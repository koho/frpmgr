package sec

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var entropy = bytesToBlob([]byte{
	0x3c, 0x08, 0x7c, 0xa8, 0x13, 0x61, 0xb2, 0x0a,
	0xd1, 0x87, 0x58, 0xcf, 0xa8, 0x2f, 0x10, 0x16,
})

func bytesToBlob(bytes []byte) *windows.DataBlob {
	blob := &windows.DataBlob{Size: uint32(len(bytes))}
	if len(bytes) > 0 {
		blob.Data = &bytes[0]
	}
	return blob
}

func Encrypt(data []byte) ([]byte, error) {
	out := windows.DataBlob{}
	err := windows.CryptProtectData(bytesToBlob(data), nil, entropy, 0, nil,
		windows.CRYPTPROTECT_LOCAL_MACHINE|windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt DPAPI protected data: %w", err)
	}
	ret := make([]byte, out.Size)
	copy(ret, unsafe.Slice(out.Data, out.Size))
	windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return ret, nil
}

func Decrypt(data []byte) ([]byte, error) {
	out := windows.DataBlob{}
	err := windows.CryptUnprotectData(bytesToBlob(data), nil, entropy, 0, nil,
		windows.CRYPTPROTECT_LOCAL_MACHINE|windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt DPAPI protected data: %w", err)
	}
	ret := make([]byte, out.Size)
	copy(ret, unsafe.Slice(out.Data, out.Size))
	windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return ret, nil
}

// EncryptFile encrypts the input file and writes to the output file.
func EncryptFile(input string, output string) error {
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	encrypted, err := Encrypt(data)
	if err != nil {
		return err
	}
	return os.WriteFile(output, encrypted, 0666)
}
