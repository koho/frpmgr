#include <windows.h>
#include <shlwapi.h>
#include <ntsecapi.h>
#include <stdint.h>
#include <msi.h>

static char msi_filename[MAX_PATH];
static HANDLE hFile = INVALID_HANDLE_VALUE;

static BOOL random_string(char ss[32]) {
    uint8_t bytes[32];
    if (!RtlGenRandom(bytes, sizeof(bytes)))
        return FALSE;
    for (int i = 0; i < 31; ++i) {
        ss[i] = (char)(bytes[i] % 26 + 97);
    }
    ss[31] = '\0';
    return TRUE;
}

static int cleanup(void)
{
    if (hFile != INVALID_HANDLE_VALUE) {
        for (int i = 0; i < 200 && !DeleteFileA(msi_filename) && GetLastError() != ERROR_FILE_NOT_FOUND; ++i)
            Sleep(200);
    }
    return 0;
}

int WINAPI WinMain(HINSTANCE hInstance, HINSTANCE hPrevInstance, PSTR pCmdLine, int nCmdShow) {
    _onexit(cleanup);
    DWORD ret = -1;
    char random_filename[32];
    if (!GetWindowsDirectoryA(msi_filename, sizeof(msi_filename)) || !PathAppendA(msi_filename, "Temp"))
        goto out;
    if (!random_string(random_filename))
        goto out;
    if (!PathAppendA(msi_filename, random_filename))
        goto out;
    HRSRC hRes = FindResource(NULL, MAKEINTRESOURCE(11), RT_RCDATA);
    if (hRes == NULL) {
        goto out;
    }
    HGLOBAL msiData = LoadResource(NULL, hRes);
    if (msiData == NULL) {
        goto out;
    }
    DWORD msiSize = SizeofResource(NULL, hRes);
    if (msiSize == 0) {
        goto out;
    }
    LPVOID pMsiData = LockResource(msiData);
    if (pMsiData == NULL) {
        goto out;
    }
    SECURITY_ATTRIBUTES security_attributes = { .nLength = sizeof(security_attributes) };
    hFile = CreateFileA(msi_filename, GENERIC_WRITE | DELETE, 0, &security_attributes, CREATE_NEW, FILE_ATTRIBUTE_TEMPORARY, NULL);
    if (hFile == INVALID_HANDLE_VALUE) {
        goto out;
    }
    DWORD bytesWritten;
    if (!WriteFile(hFile, pMsiData, msiSize, &bytesWritten, NULL) || bytesWritten != msiSize) {
        goto out;
    }
    CloseHandle(hFile);
    MsiSetInternalUI(INSTALLUILEVEL_FULL, NULL);
    MsiInstallProductA(msi_filename, NULL);
    ret = 0;
out:
    if (ret) {
        MessageBoxW(NULL, L"准备安装包时出现错误。", L"FRP 管理器", MB_OK | MB_ICONERROR);
    }
    return 0;
}
