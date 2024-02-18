package services

import (
	"context"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	modKernel32       = syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole = modKernel32.NewProc("AttachConsole")
)

type FrpClientExternalService struct {
	jo     windows.Handle
	ctx    context.Context
	cancel context.CancelFunc
	si     windows.StartupInfo
	pi     windows.ProcessInformation
	cmd    *uint16
}

func NewFrpClientExternalService(path, args string, cfgFile string) (*FrpClientExternalService, error) {
	jo, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	var info = windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	if _, err = windows.SetInformationJobObject(jo, windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		return nil, err
	}
	if err = os.Setenv("FRP_CONF", syscall.EscapeArg(cfgFile)); err != nil {
		return nil, err
	}
	if args == "" {
		args = "-c %FRP_CONF%"
	}
	args, err = registry.ExpandString(args)
	if err != nil {
		return nil, err
	}
	cmd, err := windows.UTF16PtrFromString(syscall.EscapeArg(path) + " " + args)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &FrpClientExternalService{
		jo:     jo,
		ctx:    ctx,
		cancel: cancel,
		si:     windows.StartupInfo{Cb: uint32(unsafe.Sizeof(windows.StartupInfo{}))},
		cmd:    cmd,
	}, nil
}

func (s *FrpClientExternalService) Run() {
	defer s.cancel()
	if err := windows.AssignProcessToJobObject(s.jo, windows.CurrentProcess()); err != nil {
		panic(err)
	}
	if err := windows.CreateProcess(nil, s.cmd, nil, nil, false,
		0, nil, nil, &s.si, &s.pi); err != nil {
		panic(err)
	}
	windows.WaitForSingleObject(s.pi.Process, windows.INFINITE)
}

func (s *FrpClientExternalService) Reload() error {
	return nil
}

func (s *FrpClientExternalService) Stop(wait bool) {
	var object uint32
	r, _, err := procAttachConsole.Call(uintptr(s.pi.ProcessId))
	if r == 0 || err != nil {
		goto kill
	}
	if err = windows.GenerateConsoleCtrlEvent(windows.CTRL_C_EVENT, 0); err != nil {
		goto kill
	}
	if object, err = windows.WaitForSingleObject(s.pi.Process, windows.INFINITE); err != nil {
		goto kill
	}
	if object == windows.WAIT_OBJECT_0 {
		return
	}
kill:
	windows.TerminateProcess(s.pi.Process, 0)
}

func (s *FrpClientExternalService) Done() <-chan struct{} {
	return s.ctx.Done()
}

func (s *FrpClientExternalService) Close() error {
	s.cancel()
	return windows.CloseHandle(s.jo)
}
