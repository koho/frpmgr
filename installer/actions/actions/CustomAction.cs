using Microsoft.Deployment.WindowsInstaller;
using System;
using System.Configuration.Install;
using System.Diagnostics;
using System.IO;
using System.Management;
using System.Runtime.InteropServices;
using System.ServiceProcess;
using System.Text;

namespace actions
{
    public class CustomActions
    {
        [DllImport("user32.dll", SetLastError = true)]
        public static extern IntPtr FindWindow(string lpClassName, string lpWindowName);

        [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
        public static extern int MessageBox(int hWnd, String text, String caption, uint type);

        [DllImport("Kernel32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
        public static extern IntPtr CreateFile(string lpFileName, uint dwDesiredAccess, int dwShareMode, IntPtr lpSECURITY_ATTRIBUTES, int dwCreationDisposition, int dwFlagsAndAttributes, IntPtr hTemplateFile);

        [DllImport("Kernel32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
        public static extern bool CloseHandle(IntPtr hObject);

        [DllImport("Kernel32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
        public static extern bool GetFileInformationByHandle(IntPtr handle, ref BY_HANDLE_FILE_INFORMATION hfi);

        [DllImport("Kernel32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
        public static extern int LCIDToLocaleName(uint Locale, StringBuilder lpName, int cchName, int dwFlags);

        [StructLayout(LayoutKind.Sequential)]
        public struct BY_HANDLE_FILE_INFORMATION
        {
            public uint dwFileAttributes;
            public System.Runtime.InteropServices.ComTypes.FILETIME ftCreationTime;
            public System.Runtime.InteropServices.ComTypes.FILETIME ftLastAccessTime;
            public System.Runtime.InteropServices.ComTypes.FILETIME ftLastWriteTime;
            public uint dwVolumeSerialNumber;
            public uint nFileSizeHigh;
            public uint nFileSizeLow;
            public uint nNumberOfLinks;
            public uint nFileIndexHigh;
            public uint nFileIndexLow;
        }

        public const int OPEN_EXISTING = 3;
        public const int INVALID_HANDLE_VALUE = -1;
        public const int FILE_ATTRIBUTE_NORMAL = 0x80;
        public const int MB_OK = 0;
        public const int MB_YESNO = 0x4;
        public const int MB_RETRYCANCEL = 0x5;
        public const int MB_ICONQUESTION = 0x20;
        public const int MB_ICONWARNING = 0x30;
        public const int IDYES = 6;
        public const int IDRETRY = 4;

        public static bool CalculateFileId(string path, out BY_HANDLE_FILE_INFORMATION hfi)
        {
            hfi = new BY_HANDLE_FILE_INFORMATION { };
            IntPtr file = CreateFile(path, 0, 0, IntPtr.Zero, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, IntPtr.Zero);
            if (file.ToInt32() == INVALID_HANDLE_VALUE)
                return false;
            bool ret = GetFileInformationByHandle(file, ref hfi);
            CloseHandle(file);
            if (!ret)
                return false;
            return true;
        }

        public static bool CompareFile(BY_HANDLE_FILE_INFORMATION f1, BY_HANDLE_FILE_INFORMATION f2)
        {
            return f1.dwVolumeSerialNumber == f2.dwVolumeSerialNumber && f1.nFileIndexHigh == f2.nFileIndexHigh && f1.nFileIndexLow == f2.nFileIndexLow;
        }

        public static int ForceDeleteDirectory(string path)
        {
            if (!Directory.Exists(path))
            {
                return -1;
            }
            using (ManagementObject dirObject = new ManagementObject("Win32_Directory.Name='" + path + "'"))
            {
                dirObject.Get();
                ManagementBaseObject outParams = dirObject.InvokeMethod("Delete", null, null);
                return Convert.ToInt32(outParams.Properties["ReturnValue"].Value);
            }
        }

        public static void RemoveServices(Session session, string prefix, string binPath)
        {
            ServiceController[] services = ServiceController.GetServices();
            foreach (ServiceController controller in services)
            {
                ManagementObject wmiService = new ManagementObject("Win32_Service.Name='" + controller.ServiceName + "'");
                wmiService.Get();
                string pathName = wmiService.GetPropertyValue("PathName").ToString();
                string path1 = pathName.Substring(0, Math.Min(binPath.Length, pathName.Length));
                string path2 = pathName.Substring(1, Math.Min(binPath.Length, pathName.Length - 1));
                if ((binPath.ToLower().Equals(path1.ToLower()) || binPath.ToLower().Equals(path2.ToLower())) && controller.ServiceName.StartsWith(prefix))
                {
                    try
                    {
                        controller.Stop();
                        controller.WaitForStatus(ServiceControllerStatus.Stopped);
                    }
                    catch (Exception)
                    {
                        session.Log("Failed to stop " + controller.ServiceName);
                    }

                    ServiceInstaller installer = new ServiceInstaller
                    {
                        Context = new InstallContext(),
                        ServiceName = controller.ServiceName
                    };
                    try
                    {
                        installer.Uninstall(null);
                    }
                    catch (Exception)
                    {
                        session.Log("Failed to uninstall " + controller.ServiceName);
                    }
                }
            }
        }

        [CustomAction]
        public static ActionResult KillProcesses(Session session)
        {
            session.Log("Killing FRP processes");
            string binPath = session["CustomActionData"];
            if (string.IsNullOrEmpty(binPath) || !CalculateFileId(binPath, out BY_HANDLE_FILE_INFORMATION binInfo))
            {
                return ActionResult.Success;
            }
            Process[] processes = Process.GetProcesses();
            foreach (Process p in processes)
            {
                try
                {
                    if (!CalculateFileId(p.MainModule.FileName, out BY_HANDLE_FILE_INFORMATION info))
                        continue;
                    if (CompareFile(binInfo, info))
                    {
                        p.Kill();
                        p.WaitForExit();
                    }
                }
                catch (Exception)
                {
                    continue;
                }
            }
            return ActionResult.Success;
        }

        [CustomAction]
        public static ActionResult RemoveFrpFiles(Session session)
        {
            session.Log("Removing files");
            string installPath = session["CustomActionData"];
            if (!string.IsNullOrEmpty(installPath))
            {
                ForceDeleteDirectory(Path.Combine(installPath, "profiles"));
                ForceDeleteDirectory(Path.Combine(installPath, "logs"));
            }
            return ActionResult.Success;
        }

        [CustomAction]
        public static ActionResult EvaluateFrpServices(Session session)
        {
            session.Log("Evaluate FRP Services");
            string binPath = session["CustomActionData"];
            if (string.IsNullOrEmpty(binPath))
            {
                return ActionResult.Success;
            }
            RemoveServices(session, "", binPath);
            return ActionResult.Success;
        }

        [CustomAction]
        public static ActionResult KillGUIProcesses(Session session)
        {
            session.Log("Killing FRP GUI processes");
            string binPath = session.Format(session["WixShellExecTarget"]);
            if (string.IsNullOrEmpty(binPath))
            {
                return ActionResult.Success;
            }
            Process process = new Process
            {
                StartInfo = new ProcessStartInfo()
            };
            process.StartInfo.WindowStyle = ProcessWindowStyle.Hidden;
            process.StartInfo.FileName = "wmic.exe";
            process.StartInfo.Arguments = "process where (executablepath = '" + binPath.Replace(@"\", @"\\") + "' and sessionid != 0) delete";
            process.StartInfo.UseShellExecute = true;
            process.StartInfo.Verb = "runas";
            process.Start();
            process.WaitForExit();
            return ActionResult.Success;
        }

        [CustomAction]
        public static ActionResult SetLangConfig(Session session)
        {
            session.Log("Set language config");
            string langPath = session["CustomActionData"];
            if (string.IsNullOrEmpty(langPath))
            {
                return ActionResult.Failure;
            }
            StringBuilder name = new StringBuilder(500);
            if (LCIDToLocaleName((uint)session.Language, name, name.Capacity, 0) == 0)
            {
                return ActionResult.Failure;
            }
            File.AppendAllText(langPath, name.ToString() + Environment.NewLine, Encoding.UTF8);
            return ActionResult.Success;
        }

        [CustomAction]
        public static ActionResult MoveFrpProfiles(Session session)
        {
            session.Log("Moving FRP profiles");
            string installPath = session["CustomActionData"];
            if (string.IsNullOrEmpty(installPath))
            {
                return ActionResult.Failure;
            }
            string profilePath = Path.Combine(installPath, "profiles");
            Directory.CreateDirectory(profilePath);
            foreach (string profile in Directory.GetFiles(installPath, "*.ini"))
            {
                try
                {
                    File.Move(profile, Path.Combine(profilePath, Path.GetFileName(profile)));
                }
                catch (Exception e)
                {
                    session.Log(e.Message);
                }
            }
            return ActionResult.Success;
        }

        [CustomAction]
        public static ActionResult RemoveOldFrpServices(Session session)
        {
            session.Log("Remove old FRP Services");
            string binPath = session["CustomActionData"];
            if (string.IsNullOrEmpty(binPath))
            {
                return ActionResult.Success;
            }
            RemoveServices(session, "FRPC$", binPath);
            return ActionResult.Success;
        }
    }
}
