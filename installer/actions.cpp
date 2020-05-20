#include <windows.h>
#include <strsafe.h>
#include <msiquery.h>
#include <wcautil.h>
#include <tlhelp32.h>
#include <tchar.h>
#include <msidefs.h>
#include <stdlib.h>
#include <io.h>
#define CA //extern "C" _declspec(dllexport)

struct IDField { DWORD dwVolume, dwIndexHigh, dwIndexLow; };

static bool CalculateFileId(const TCHAR* path, struct IDField* id)
{
	BY_HANDLE_FILE_INFORMATION file_info = { 0 };
	HANDLE file;
	bool ret;

	file = CreateFile(path, 0, 0, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
	if (file == INVALID_HANDLE_VALUE)
		return false;
	ret = GetFileInformationByHandle(file, &file_info);
	CloseHandle(file);
	if (!ret)
		return false;
	id->dwVolume = file_info.dwVolumeSerialNumber;
	id->dwIndexHigh = file_info.nFileIndexHigh;
	id->dwIndexLow = file_info.nFileIndexLow;
	return true;
}

static bool GetFrpMgrPath(MSIHANDLE hInstall, WCHAR* path, DWORD* dwSize) {
	WCHAR productCode[MAX_GUID_CHARS + 1];
	DWORD dwPCSize = _countof(productCode);
	UINT ret;
	ret = MsiGetProperty(hInstall, L"ProductCode", productCode, &dwPCSize);
	if (ret != ERROR_SUCCESS) {
		return false;
	}
	MsiGetComponentPath(productCode, L"{E39EABEF-A7EB-4EAF-AD3E-A1254450BBE1}", path, dwSize);
	return true;
}

static bool GetInstallPath(MSIHANDLE hInstall, WCHAR* path, DWORD* dwSize) {
	WCHAR productCode[MAX_GUID_CHARS + 1];
	DWORD dwPCSize = _countof(productCode);
	UINT ret;
	ret = MsiGetProperty(hInstall, L"ProductCode", productCode, &dwPCSize);
	if (ret != ERROR_SUCCESS) {
		return false;
	}
	ret = MsiGetProductInfo(productCode, INSTALLPROPERTY_INSTALLLOCATION, path, dwSize);
	return ret == ERROR_SUCCESS;
}


CA UINT __stdcall KillProcesses(MSIHANDLE hInstall)
{
	HRESULT hr = S_OK;
	UINT er = ERROR_SUCCESS;

	hr = WcaInitialize(hInstall, "KillProcesses");
	ExitOnFailure(hr, "Failed to initialize KillProcesses()");

	WcaLog(LOGMSG_STANDARD, "Initialized KillProcesses().");

	WCHAR binPath[MAX_PATH];
	memset(binPath, 0, sizeof(binPath));
	DWORD dwBinPathSize = _countof(binPath);
	if (!GetFrpMgrPath(hInstall, binPath, &dwBinPathSize) || wcslen(binPath) == 0)
		goto LExit;

	HANDLE hSnapshot, hProcess;
	PROCESSENTRY32 entry;
	entry.dwSize = sizeof(PROCESSENTRY32);
	TCHAR processPath[MAX_PATH + 1];
	DWORD dwProcessPathLen = _countof(processPath);
	struct IDField appFileId, fileId;

	if (!CalculateFileId(binPath, &appFileId))
		goto LExit;

	hSnapshot = CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0);
	if (hSnapshot == INVALID_HANDLE_VALUE)
		goto LExit;
	for (bool ret = Process32First(hSnapshot, &entry); ret; ret = Process32Next(hSnapshot, &entry)) {
		if (_tcsicmp(entry.szExeFile, TEXT("frpmgr.exe")))
			continue;
		hProcess = OpenProcess(PROCESS_TERMINATE | PROCESS_QUERY_LIMITED_INFORMATION, false, entry.th32ProcessID);
		if (!hProcess)
			continue;

		if (!QueryFullProcessImageName(hProcess, 0, processPath, &dwProcessPathLen))
			goto next;
		if (!CalculateFileId(processPath, &fileId))
			goto next;
		ret = false;
		if (!memcmp(&fileId, &appFileId, sizeof(fileId))) {
			ret = true;
		}
		if (!ret)
			goto next;
		if (TerminateProcess(hProcess, 0)) {
			WaitForSingleObject(hProcess, INFINITE);
		}
	next:
		CloseHandle(hProcess);
	}
	CloseHandle(hSnapshot);

LExit:
	er = SUCCEEDED(hr) ? ERROR_SUCCESS : ERROR_INSTALL_FAILURE;
	return WcaFinalize(er);
}

static UINT InsertServiceControl(MSIHANDLE hView, const TCHAR* szServiceName)
{
	static unsigned int index = 0;
	UINT ret;
	MSIHANDLE record = NULL;
	TCHAR row_identifier[_countof(TEXT("frpmgr_service_control_4294967296"))];

	if (_sntprintf_s(row_identifier, _countof(row_identifier), TEXT("frpmgr_service_control_%u"), ++index) >= _countof(row_identifier)) {
		ret = ERROR_INSTALL_FAILURE;
		goto LExit;
	}
	record = MsiCreateRecord(5);
	if (!record) {
		ret = ERROR_INSTALL_FAILURE;
		goto LExit;
	}

	MsiRecordSetString(record, 1/*ServiceControl*/, row_identifier);
	MsiRecordSetString(record, 2/*Name          */, szServiceName);
	MsiRecordSetInteger(record, 3/*Event         */, msidbServiceControlEventStop | msidbServiceControlEventUninstallStop | msidbServiceControlEventUninstallDelete);
	MsiRecordSetString(record, 4/*Component_    */, TEXT("MainApplication"));
	MsiRecordSetInteger(record, 5/*Wait          */, 1); /* Waits 30 seconds. */
	WcaLog(LOGMSG_STANDARD, "Uninstalling frpc service: %s", szServiceName);
	ret = MsiViewExecute(hView, record);
	if (ret != ERROR_SUCCESS) {
		WcaLog(LOGMSG_STANDARD, "MsiViewExecute failed(%d).", ret);
		goto LExit;
	}

LExit:
	if (record)
		MsiCloseHandle(record);
	return ret;
}

CA UINT __stdcall EvaluateFrpServices(MSIHANDLE hInstall)
{
	HRESULT hr = S_OK;
	UINT ret = ERROR_INSTALL_FAILURE;
	hr = WcaInitialize(hInstall, "EvaluateFrpServices");
	if (!SUCCEEDED(hr))
		return WcaFinalize(ERROR_INSTALL_FAILURE);

	WCHAR binPath[MAX_PATH];
	DWORD dwBinPathSize = _countof(binPath);
	MSIHANDLE hDb = NULL, hView = NULL;
	SC_HANDLE hScm = NULL;
	SC_HANDLE hService = NULL;
	ENUM_SERVICE_STATUS_PROCESS* serviceStatus = NULL;
	QUERY_SERVICE_CONFIG* serviceConfig = NULL;
	DWORD dwServiceStatusResume = 0;
	enum { SERVICE_STATUS_PROCESS_SIZE = 0x10000, SERVICE_CONFIG_SIZE = 8000 };

	memset(binPath, 0, sizeof(binPath));
	if (!GetFrpMgrPath(hInstall, binPath, &dwBinPathSize) || wcslen(binPath) == 0)
		goto LExit;

	hDb = MsiGetActiveDatabase(hInstall);
	if (!hDb) {
		WcaLog(LOGMSG_STANDARD, "MsiGetActiveDatabase failed.");
		goto LExit;
	}
	ret = MsiDatabaseOpenView(hDb,
		TEXT("INSERT INTO `ServiceControl` (`ServiceControl`, `Name`, `Event`, `Component_`, `Wait`) VALUES(?, ?, ?, ?, ?) TEMPORARY"),
		&hView);
	if (ret != ERROR_SUCCESS) {
		WcaLog(LOGMSG_STANDARD, "MsiDatabaseOpenView failed.");
		goto LExit;
	}
	hScm = OpenSCManager(NULL, SERVICES_ACTIVE_DATABASE, SC_MANAGER_CONNECT | SC_MANAGER_ENUMERATE_SERVICE);
	if (!hScm) {
		ret = GetLastError();
		WcaLog(LOGMSG_STANDARD, "MsiDatabaseOpenView failed(%d).", ret);
		goto LExit;
	}

	serviceStatus = (ENUM_SERVICE_STATUS_PROCESS*)LocalAlloc(LMEM_FIXED, SERVICE_STATUS_PROCESS_SIZE);
	serviceConfig = (QUERY_SERVICE_CONFIG*)LocalAlloc(LMEM_FIXED, SERVICE_CONFIG_SIZE);
	if (!serviceStatus || !serviceConfig) {
		ret = GetLastError();
		WcaLog(LOGMSG_STANDARD, "LocalAlloc failed(%d).", ret);
		goto LExit;
	}

	for (bool more_services = true; more_services;) {
		DWORD dwServiceStatusSize = 0, serviceStatusCount = 0;
		DWORD dwBytesNeeded;
		if (EnumServicesStatusEx(hScm, SC_ENUM_PROCESS_INFO, SERVICE_WIN32, SERVICE_STATE_ALL, (LPBYTE)serviceStatus,
			SERVICE_STATUS_PROCESS_SIZE, &dwServiceStatusSize, &serviceStatusCount,
			&dwServiceStatusResume, NULL))
			more_services = false;
		else {
			ret = GetLastError();
			if (ret != ERROR_MORE_DATA) {
				WcaLog(LOGMSG_STANDARD, "EnumServicesStatusEx failed(%d).", ret);
				break;
			}
		}

		for (DWORD i = 0; i < serviceStatusCount; ++i) {
			hService = OpenService(hScm, serviceStatus[i].lpServiceName, SERVICE_QUERY_CONFIG);
			if (hService == NULL) {
				continue;
			}
			if (!QueryServiceConfig(hService, serviceConfig, SERVICE_CONFIG_SIZE, &dwBytesNeeded)) {
				continue;
			}

			if (_tcsnicmp(serviceConfig->lpBinaryPathName, binPath, wcslen(binPath)) && _tcsnicmp(serviceConfig->lpBinaryPathName + 1, binPath, wcslen(binPath)))
				continue;
			InsertServiceControl(hView, serviceStatus[i].lpServiceName);
		}
	}
	ret = ERROR_SUCCESS;

LExit:
	if (serviceStatus) {
		LocalFree(serviceStatus);
	}
	if (serviceConfig) {
		LocalFree(serviceConfig);
	}
	if (hScm)
		CloseServiceHandle(hScm);
	if (hView)
		MsiCloseHandle(hView);
	if (hDb)
		MsiCloseHandle(hDb);
	return WcaFinalize(ret == ERROR_SUCCESS ? ret : ERROR_INSTALL_FAILURE);
}

static bool RemoveDirectoryRecursive(TCHAR path[MAX_PATH]) {
	HANDLE hFind;
	WIN32_FIND_DATA findData;
	TCHAR* pathEnd;

	pathEnd = path + _tcsnlen(path, MAX_PATH);
	wcscat_s(path, MAX_PATH, TEXT("\\*.*"));

	hFind = FindFirstFileEx(path, FindExInfoBasic, &findData, FindExSearchNameMatch, NULL, 0);
	if (hFind == INVALID_HANDLE_VALUE) {
		return false;
	}
	do {
		if (findData.cFileName[0] == TEXT('.') && (findData.cFileName[1] == TEXT('\0') || (findData.cFileName[1] == TEXT('.') && findData.cFileName[2] == TEXT('\0'))))
			continue;

		pathEnd[0] = TEXT('\0');
		wcscat_s(path, MAX_PATH, TEXT("\\"));
		wcscat_s(path, MAX_PATH, findData.cFileName);

		if (findData.dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY) {
			RemoveDirectoryRecursive(path);
			continue;
		}

		if ((findData.dwFileAttributes & FILE_ATTRIBUTE_READONLY) && !SetFileAttributes(path, findData.dwFileAttributes & ~FILE_ATTRIBUTE_READONLY))
			WcaLog(LOGMSG_STANDARD, "SetFileAttributes failed");

		DeleteFile(path);
	} while (FindNextFile(hFind, &findData));
	FindClose(hFind);

	pathEnd[0] = TEXT('\0');
	if (RemoveDirectory(path)) {
		return true;
	}
	else {
		return false;
	}
}

CA UINT __stdcall RemoveConfigFiles(MSIHANDLE hInstall) {
	HRESULT hr = S_OK;
	UINT ret = ERROR_INSTALL_FAILURE;
	hr = WcaInitialize(hInstall, "RemoveConfigFiles");
	if (!SUCCEEDED(hr))
		return WcaFinalize(ERROR_INSTALL_FAILURE);

	WCHAR installPath[MAX_PATH];
	DWORD dwBinPathSize = _countof(installPath);
	memset(installPath, 0, sizeof(installPath));
	if (!GetInstallPath(hInstall, installPath, &dwBinPathSize) || wcslen(installPath) == 0) {
		goto LExit;
	}

	wchar_t warnText[500];
	wsprintf(warnText, TEXT("是否删除配置文件?\n\n注意：若要重新使用配置文件，下次安装时必须安装到此目录：\n\n%s"), installPath);
	int result = MessageBox(FindWindow(NULL, TEXT("FRP")), warnText, TEXT("卸载提示"), MB_YESNO|MB_ICONQUESTION);
	if (result == IDYES) {
		// Delete config files
		WCHAR configFiles[MAX_PATH];
		wcscpy_s(configFiles, installPath);
		wcscat_s(configFiles, TEXT("*.ini"));
		WIN32_FIND_DATA fileData;
		HANDLE hFindFile = FindFirstFile(configFiles, &fileData);
		if (hFindFile == INVALID_HANDLE_VALUE) {
			goto LDelLog;
		}
		do {
			WCHAR conf[MAX_PATH];
			wcscpy_s(conf, installPath);
			wcscat_s(conf, fileData.cFileName);
			DeleteFile(conf);
		} while (FindNextFile(hFindFile, &fileData));
		FindClose(hFindFile);
		hFindFile = NULL;
LDelLog:
		// Delete logs
		WCHAR logDir[MAX_PATH];
		wcscpy_s(logDir, installPath);
		wcscat_s(logDir, TEXT("logs"));
		hFindFile = FindFirstFile(logDir, &fileData);
		if ((hFindFile != INVALID_HANDLE_VALUE) && (fileData.dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY)) {
			RemoveDirectoryRecursive(logDir);
			FindClose(hFindFile);
		}
	}
	ret = ERROR_SUCCESS;

LExit:
	return WcaFinalize(ret);
}

// DllMain - Initialize and cleanup WiX custom action utils.
extern "C" BOOL WINAPI DllMain(
	__in HINSTANCE hInst,
	__in ULONG ulReason,
	__in LPVOID
	)
{
	switch(ulReason)
	{
	case DLL_PROCESS_ATTACH:
		WcaGlobalInitialize(hInst);
		break;

	case DLL_PROCESS_DETACH:
		WcaGlobalFinalize();
		break;
	}

	return TRUE;
}
