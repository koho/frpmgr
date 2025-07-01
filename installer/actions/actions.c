#include <windows.h>
#include <msi.h>
#include <msidefs.h>
#include <msiquery.h>
#include <tlhelp32.h>
#include <shlwapi.h>

#define LEGACY_SERVICE_PREFIX L"FRPC$"
#define SERVICE_PREFIX        L"frpmgr_"

static void Log(MSIHANDLE installer, INSTALLMESSAGE messageType, const WCHAR* format, ...)
{
    MSIHANDLE record = MsiCreateRecord(0);
    if (!record)
        return;
    LPWSTR pBuffer = NULL;
    va_list args = NULL;
    va_start(args, format);
    FormatMessageW(FORMAT_MESSAGE_FROM_STRING | FORMAT_MESSAGE_ALLOCATE_BUFFER, format,
        0, 0, (LPWSTR)&pBuffer, 0, &args);
    va_end(args);
    if (pBuffer)
    {
        MsiRecordSetStringW(record, 0, pBuffer);
        MsiProcessMessage(installer, messageType, record);
        LocalFree(pBuffer);
    }
    MsiCloseHandle(record);
}

static BOOL GetFileInformation(const LPWSTR path, BY_HANDLE_FILE_INFORMATION* fileInfo)
{
    HANDLE file = CreateFileW(path, 0, 0, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
    if (file == INVALID_HANDLE_VALUE)
        return FALSE;
    BOOL ret = GetFileInformationByHandle(file, fileInfo);
    CloseHandle(file);
    return ret;
}

static void KillProcessesEx(LPWSTR path, BOOL uiOnly)
{
    HANDLE snapshot, process;
    BY_HANDLE_FILE_INFORMATION fileInfo = { 0 }, procFileInfo = { 0 };
    WCHAR procPath[MAX_PATH];
    PROCESSENTRY32W entry;
    entry.dwSize = sizeof(PROCESSENTRY32W);

    LPWSTR filename = PathFindFileNameW(path);
    if (!GetFileInformation(path, &fileInfo))
        return;

    snapshot = CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0);
    if (snapshot == INVALID_HANDLE_VALUE)
        return;
    for (BOOL ret = Process32FirstW(snapshot, &entry); ret; ret = Process32NextW(snapshot, &entry))
    {
        if (_wcsicmp(entry.szExeFile, filename))
            continue;
        process = OpenProcess(PROCESS_TERMINATE | PROCESS_QUERY_LIMITED_INFORMATION, FALSE, entry.th32ProcessID);
        if (!process)
            continue;
        DWORD procPathLen = _countof(procPath);
        DWORD sessionId = 0;
        if (!QueryFullProcessImageNameW(process, 0, procPath, &procPathLen))
            goto next;
        if (!GetFileInformation(procPath, &procFileInfo))
            goto next;
        if (fileInfo.dwVolumeSerialNumber != procFileInfo.dwVolumeSerialNumber ||
            fileInfo.nFileIndexHigh != procFileInfo.nFileIndexHigh ||
            fileInfo.nFileIndexLow != procFileInfo.nFileIndexLow)
            goto next;
        if (!ProcessIdToSessionId(entry.th32ProcessID, &sessionId))
            goto next;
        if (uiOnly && sessionId == 0)
            goto next;
        if (TerminateProcess(process, 1))
            WaitForSingleObject(process, INFINITE);
    next:
        CloseHandle(process);
    }
    CloseHandle(snapshot);
    return;
}

__declspec(dllexport) UINT __stdcall KillFrpProcesses(MSIHANDLE installer)
{
    WCHAR path[MAX_PATH];
    DWORD pathLen = _countof(path);
    UINT ret = MsiGetPropertyW(installer, L"CustomActionData", path, &pathLen);
    if (ret != ERROR_SUCCESS)
    {
        Log(installer, INSTALLMESSAGE_ERROR, L"Failed to load CustomActionData");
        return ERROR_SUCCESS;
    }
    if (path[0])
        KillProcessesEx(path, FALSE);
    return ERROR_SUCCESS;
}

__declspec(dllexport) UINT __stdcall KillFrpGUIProcesses(MSIHANDLE installer)
{
    WCHAR path[MAX_PATH];
    DWORD pathLen = _countof(path);
    MSIHANDLE record = MsiCreateRecord(0);
    if (!record)
        return ERROR_SUCCESS;
    MsiRecordSetStringW(record, 0, L"[#frpmgr.exe]");
    UINT ret = MsiFormatRecordW(installer, record, path, &pathLen);
    MsiCloseHandle(record);
    if (ret != ERROR_SUCCESS)
    {
        Log(installer, INSTALLMESSAGE_ERROR, L"Failed to load application path");
        return ERROR_SUCCESS;
    }
    if (path[0])
        KillProcessesEx(path, TRUE);
    return ERROR_SUCCESS;
}

__declspec(dllexport) UINT __stdcall EvaluateFrpServices(MSIHANDLE installer)
{
    SC_HANDLE scm = NULL;
    LPENUM_SERVICE_STATUS_PROCESSW services = NULL;
    DWORD SERVICE_STATUS_PROCESS_SIZE = 0x10000;
    DWORD resume = 0;
    LPQUERY_SERVICE_CONFIGW cfg = NULL;
    DWORD cfgSize = 0;

    MSIHANDLE db = 0, view = 0;
    WCHAR path[MAX_PATH];
    DWORD pathLen = _countof(path);
    BY_HANDLE_FILE_INFORMATION fileInfo = { 0 }, svcFileInfo = { 0 };
    BOOL fileInfoExists = FALSE;
    MSIHANDLE record = MsiCreateRecord(0);
    if (!record)
        goto out;
    MsiRecordSetStringW(record, 0, L"[#frpmgr.exe]");
    UINT ret = MsiFormatRecordW(installer, record, path, &pathLen);
    MsiCloseHandle(record);
    if (ret != ERROR_SUCCESS)
    {
        Log(installer, INSTALLMESSAGE_ERROR, L"Failed to load application path");
        goto out;
    }
    if (!path[0])
        goto out;
    fileInfoExists = GetFileInformation(path, &fileInfo);

    db = MsiGetActiveDatabase(installer);
    if (!db)
    {
        Log(installer, INSTALLMESSAGE_ERROR, L"MsiGetActiveDatabase failed");
        goto out;
    }
    ret = MsiDatabaseOpenViewW(db,
        L"INSERT INTO `ServiceControl` (`ServiceControl`, `Name`, `Event`, `Component_`, `Wait`) VALUES(?, ?, ?, ?, ?) TEMPORARY",
        &view);
    if (ret != ERROR_SUCCESS)
    {
        Log(installer, INSTALLMESSAGE_ERROR, L"MsiDatabaseOpenView failed (%1)", ret);
        goto out;
    }

    scm = OpenSCManagerW(NULL, SERVICES_ACTIVE_DATABASEW, SC_MANAGER_CONNECT | SC_MANAGER_ENUMERATE_SERVICE);
    if (!scm)
    {
        Log(installer, INSTALLMESSAGE_ERROR, L"OpenSCManager failed (%1)", GetLastError());
        goto out;
    }

    services = (LPENUM_SERVICE_STATUS_PROCESSW)LocalAlloc(LMEM_FIXED, SERVICE_STATUS_PROCESS_SIZE);
    if (!services)
    {
        Log(installer, INSTALLMESSAGE_ERROR, L"LocalAlloc failed (%1)", GetLastError());
        goto out;
    }
    for (BOOL more = TRUE; more;)
    {
        DWORD bytesNeeded = 0, count = 0;
        if (EnumServicesStatusExW(scm, SC_ENUM_PROCESS_INFO, SERVICE_WIN32, SERVICE_STATE_ALL, (LPBYTE)services,
            SERVICE_STATUS_PROCESS_SIZE, &bytesNeeded, &count, &resume, NULL))
            more = FALSE;
        else
        {
            ret = GetLastError();
            if (ret != ERROR_MORE_DATA)
            {
                Log(installer, INSTALLMESSAGE_ERROR, L"EnumServicesStatusEx failed (%1)", ret);
                break;
            }
        }

        for (DWORD i = 0; i < count; ++i)
        {
            INT legacy;
            if ((legacy = _wcsnicmp(services[i].lpServiceName, LEGACY_SERVICE_PREFIX, _countof(LEGACY_SERVICE_PREFIX) - 1)) &&
                _wcsnicmp(services[i].lpServiceName, SERVICE_PREFIX, _countof(SERVICE_PREFIX) - 1))
                continue;

            SC_HANDLE service = OpenServiceW(scm, services[i].lpServiceName, SERVICE_QUERY_CONFIG);
            if (!service)
                continue;
            BOOL ok = FALSE;
            while (!(ok = QueryServiceConfigW(service, cfg, cfgSize, &bytesNeeded)) && GetLastError() == ERROR_INSUFFICIENT_BUFFER)
            {
                if (cfg)
                    LocalFree(cfg);
                else
                    bytesNeeded += sizeof(path) + 256 * sizeof(WCHAR); // Additional size for path and display name.
                cfgSize = bytesNeeded;
                cfg = (LPQUERY_SERVICE_CONFIGW)LocalAlloc(LMEM_FIXED, cfgSize);
                if (!cfg)
                {
                    Log(installer, INSTALLMESSAGE_ERROR, L"LocalAlloc failed (%1)", GetLastError());
                    break;
                }
            }
            CloseServiceHandle(service);

            if (!ok || cfg == NULL)
                continue;
            INT nArgs = 0;
            LPWSTR* args = CommandLineToArgvW(cfg->lpBinaryPathName, &nArgs);
            if (!args)
                continue;
            ok = nArgs >= 1 && (fileInfoExists ?
                (GetFileInformation(args[0], &svcFileInfo) &&
                    fileInfo.dwVolumeSerialNumber == svcFileInfo.dwVolumeSerialNumber &&
                    fileInfo.nFileIndexHigh == svcFileInfo.nFileIndexHigh &&
                    fileInfo.nFileIndexLow == svcFileInfo.nFileIndexLow) : _wcsicmp(args[0], path) == 0);
            LocalFree(args);
            if (!ok)
                continue;

            Log(installer, INSTALLMESSAGE_INFO, L"Scheduling stop on upgrade or removal on uninstall of service %1", services[i].lpServiceName);
            GUID guid;
            if (FAILED(CoCreateGuid(&guid)))
                continue;
            WCHAR identifier[40];
            if (StringFromGUID2(&guid, identifier, _countof(identifier)) == 0)
                continue;
            record = MsiCreateRecord(5);
            if (!record)
                continue;
            MsiRecordSetStringW(record, 1, identifier);
            MsiRecordSetStringW(record, 2, services[i].lpServiceName);
            MsiRecordSetInteger(record, 3, msidbServiceControlEventStop | msidbServiceControlEventUninstallStop | (legacy == 0 ? msidbServiceControlEventDelete : 0) | msidbServiceControlEventUninstallDelete);
            MsiRecordSetStringW(record, 4, L"frpmgr.exe");
            MsiRecordSetInteger(record, 5, 1);
            ret = MsiViewExecute(view, record);
            MsiCloseHandle(record);
            if (ret != ERROR_SUCCESS)
                Log(installer, INSTALLMESSAGE_ERROR, L"MsiViewExecute failed for service %1 (%2)", services[i].lpServiceName, ret);
        }
    }
    LocalFree(services);
    if (cfg)
        LocalFree(cfg);

out:
    if (scm)
        CloseServiceHandle(scm);
    if (view)
        MsiCloseHandle(view);
    if (db)
        MsiCloseHandle(db);
    return ERROR_SUCCESS;
}

__declspec(dllexport) UINT __stdcall SetLangConfig(MSIHANDLE installer)
{
    WCHAR path[MAX_PATH];
    DWORD pathLen = _countof(path);
    UINT ret = MsiGetPropertyW(installer, L"CustomActionData", path, &pathLen);
    if (ret != ERROR_SUCCESS)
    {
        Log(installer, INSTALLMESSAGE_ERROR, L"Failed to load CustomActionData");
        goto out;
    }
    if (!path[0] || !PathAppendW(path, L"lang.config"))
        goto out;

    WCHAR localeName[LOCALE_NAME_MAX_LENGTH];
    if (LCIDToLocaleName(MsiGetLanguage(installer), localeName, _countof(localeName), 0) == 0)
        goto out;

    HANDLE langFile = CreateFileW(path, GENERIC_WRITE, 0, NULL, CREATE_ALWAYS, FILE_ATTRIBUTE_NORMAL, NULL);
    if (langFile == INVALID_HANDLE_VALUE)
        goto out;
    CHAR buf[LOCALE_NAME_MAX_LENGTH];
    DWORD bytesWritten = WideCharToMultiByte(CP_UTF8, 0, localeName, -1, buf, sizeof(buf), NULL, NULL);
    if (bytesWritten > 0)
        WriteFile(langFile, buf, bytesWritten - 1, &bytesWritten, NULL);
    CloseHandle(langFile);

out:
    return ERROR_SUCCESS;
}

__declspec(dllexport) UINT __stdcall MoveFrpProfiles(MSIHANDLE installer)
{
    WIN32_FIND_DATAW findData;
    const WCHAR* dirs[] = { NULL, L"profiles" };

    WCHAR path[MAX_PATH], newPath[MAX_PATH];
    DWORD pathLen = _countof(path), newPathLen = _countof(newPath);
    UINT ret = MsiGetPropertyW(installer, L"CustomActionData", path, &pathLen);
    if (ret != ERROR_SUCCESS)
    {
        Log(installer, INSTALLMESSAGE_ERROR, L"Failed to load CustomActionData");
        goto out;
    }
    if (!path[0] || !PathCombineW(newPath, path, L"profiles"))
        goto out;
    if (CreateDirectoryW(newPath, NULL) == 0 && GetLastError() != ERROR_ALREADY_EXISTS)
        goto out;
    newPathLen = wcsnlen_s(newPath, _countof(newPath) - 1);
    for (size_t i = 0; i < _countof(dirs); i++)
    {
        path[pathLen] = L'\0';
        if (dirs[i] && !PathAppendW(path, dirs[i]))
            continue;
        pathLen = wcsnlen_s(path, _countof(path) - 1);
        if (!PathAppendW(path, L"*.ini"))
            continue;
        HANDLE hFind = FindFirstFileExW(path, FindExInfoBasic, &findData, FindExSearchNameMatch, NULL, 0);
        if (hFind != INVALID_HANDLE_VALUE)
        {
            do
            {
                if (findData.dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY || !PathMatchSpecW(findData.cFileName, L"*.ini"))
                    continue;
                path[pathLen] = L'\0';
                newPath[newPathLen] = L'\0';
                if (!PathAppendW(path, findData.cFileName) ||
                    !PathRenameExtensionW(findData.cFileName, L".conf") ||
                    !PathAppendW(newPath, findData.cFileName))
                    continue;
                MoveFileW(path, newPath);
            } while (FindNextFileW(hFind, &findData));
            FindClose(hFind);
        }
    }
out:
    return ERROR_SUCCESS;
}

__declspec(dllexport) UINT __stdcall RemoveFrpFiles(MSIHANDLE installer)
{
    WCHAR path[MAX_PATH];
    DWORD pathLen = _countof(path);
    UINT ret = MsiGetPropertyW(installer, L"CustomActionData", path, &pathLen);
    if (ret != ERROR_SUCCESS)
    {
        Log(installer, INSTALLMESSAGE_ERROR, L"Failed to load CustomActionData");
        return ERROR_SUCCESS;
    }

    const WCHAR* appFiles[] = { L"app.json", L"lang.config" };
    for (size_t i = 0; i < _countof(appFiles); i++)
    {
        path[pathLen] = L'\0';
        if (!PathAppendW(path, appFiles[i]))
            return ERROR_SUCCESS;
        DeleteFileW(path);
    }

    WIN32_FIND_DATAW findData;
    const WCHAR* files[][2] = { {L"profiles", L"*.conf"}, {L"logs", L"*.log"} };
    for (size_t i = 0; i < _countof(files); i++)
    {
        path[pathLen] = L'\0';
        if (!PathAppendW(path, files[i][0]))
            continue;
        SIZE_T dirLen = wcsnlen_s(path, _countof(path) - 1);
        if (!PathAppendW(path, files[i][1]))
            continue;
        HANDLE hFind = FindFirstFileExW(path, FindExInfoBasic, &findData, FindExSearchNameMatch, NULL, 0);
        if (hFind != INVALID_HANDLE_VALUE)
        {
            do
            {
                if (findData.dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY || !PathMatchSpecW(findData.cFileName, files[i][1]))
                    continue;
                path[dirLen] = L'\0';
                if (!PathAppendW(path, findData.cFileName))
                    continue;
                DeleteFileW(path);
            } while (FindNextFileW(hFind, &findData));
            FindClose(hFind);
        }
        path[dirLen] = L'\0';
        RemoveDirectoryW(path);
    }
    return ERROR_SUCCESS;
}
