#ifndef UNICODE
#define UNICODE
#endif

#include <windows.h>
#include <shlwapi.h>
#include <ntsecapi.h>
#include <stdint.h>
#include <stdio.h>
#include <msi.h>
#include "resource.h"

typedef struct {
    LCID id;
    TCHAR name[16];
    char code[10];
} Language;

static TCHAR msiFile[MAX_PATH];
static HANDLE hFile = INVALID_HANDLE_VALUE;
static Language languages[] = {
        {2052, TEXT("简体中文"),    "zh-CN"},
        {1028, TEXT("繁體中文"),    "zh-TW"},
        {1033, TEXT("English"), "en-US"},
        {1041, TEXT("日本語"),     "ja-JP"},
        {1042, TEXT("한국어"),     "ko-KR"},
        {3082, TEXT("Español"), "es-ES"},
};

static BOOL RandomString(TCHAR ss[32]) {
    uint8_t bytes[32];
    if (!RtlGenRandom(bytes, sizeof(bytes)))
        return FALSE;
    for (int i = 0; i < 31; ++i) {
        ss[i] = (TCHAR) (bytes[i] % 26 + 97);
    }
    ss[31] = '\0';
    return TRUE;
}

static int Cleanup(void) {
    if (hFile != INVALID_HANDLE_VALUE) {
        for (int i = 0; i < 200 && !DeleteFile(msiFile) && GetLastError() != ERROR_FILE_NOT_FOUND; ++i)
            Sleep(200);
    }
    return 0;
}

static Language *GetPreferredLang(TCHAR *folder) {
    TCHAR langPath[MAX_PATH];
    if (PathCombine(langPath, folder, L"lang.config") == NULL) {
        return NULL;
    }
    FILE *file;
    if (_wfopen_s(&file, langPath, L"r") != 0) {
        return NULL;
    }
    char lang[sizeof(languages[0].code)];
    while (fgets(lang, sizeof(lang), file) != NULL) {
        for (int i = 0; i < sizeof(languages) / sizeof(languages[0]); i++) {
            if (strncmp(lang, languages[i].code, strlen(languages[i].code)) == 0) {
                fclose(file);
                return &languages[i];
            }
        }
    }
    fclose(file);
    return NULL;
}

INT_PTR CALLBACK LangDialog(HWND hDlg, UINT message, WPARAM wParam, LPARAM lParam) {
    switch (message) {
        case WM_INITDIALOG:
            for (int i = 0; i < sizeof(languages) / sizeof(languages[0]); i++) {
                SendDlgItemMessage(hDlg, IDC_LANG_COMBO, CB_ADDSTRING, 0, (LPARAM) languages[i].name);
            }
            SendDlgItemMessage(hDlg, IDC_LANG_COMBO, CB_SETCURSEL, 0, 0);
            return (INT_PTR) TRUE;

        case WM_COMMAND:
            if (LOWORD(wParam) == IDOK || LOWORD(wParam) == IDCANCEL) {
                INT_PTR nResult = LOWORD(wParam);
                if (LOWORD(wParam) == IDOK) {
                    int idx = SendDlgItemMessage(hDlg, IDC_LANG_COMBO, CB_GETCURSEL, 0, 0);
                    nResult = (INT_PTR) &languages[idx];
                }
                EndDialog(hDlg, nResult);
                return (INT_PTR) TRUE;
            }
            break;
    }
    return (INT_PTR) FALSE;
}

int WINAPI WinMain(HINSTANCE hInstance, HINSTANCE hPrevInstance, PSTR pCmdLine, int nCmdShow) {
    _onexit(Cleanup);
    // Retrieve install location
    TCHAR installPath[MAX_PATH];
    DWORD dwSize = MAX_PATH;
    memset(installPath, 0, dwSize);
    Language *lang = NULL;
    if (MsiLocateComponent(L"{E39EABEF-A7EB-4EAF-AD3E-A1254450BBE1}", installPath, &dwSize) >= 0 && wcslen(installPath) > 0) {
        PathRemoveFileSpec(installPath);
        lang = GetPreferredLang(installPath);
    }
    if (lang == NULL) {
        INT_PTR nResult = DialogBox(hInstance, MAKEINTRESOURCE(IDD_LANG_DIALOG), NULL, LangDialog);
        if (nResult == IDCANCEL) {
            return 0;
        }
        lang = (Language *) nResult;
    }
    TCHAR randFile[32];
    if (!GetWindowsDirectory(msiFile, sizeof(msiFile)) || !PathAppend(msiFile, L"Temp"))
        return 1;
    if (!RandomString(randFile))
        return 1;
    if (!PathAppend(msiFile, randFile))
        return 1;
    HRSRC hRes = FindResource(NULL, MAKEINTRESOURCE(IDR_MSI), RT_RCDATA);
    if (hRes == NULL) {
        return 1;
    }
    HGLOBAL msiData = LoadResource(NULL, hRes);
    if (msiData == NULL) {
        return 1;
    }
    DWORD msiSize = SizeofResource(NULL, hRes);
    if (msiSize == 0) {
        return 1;
    }
    LPVOID pMsiData = LockResource(msiData);
    if (pMsiData == NULL) {
        return 1;
    }
    SECURITY_ATTRIBUTES security_attributes = {.nLength = sizeof(security_attributes)};
    hFile = CreateFile(msiFile, GENERIC_WRITE | DELETE, 0, &security_attributes, CREATE_NEW,
                       FILE_ATTRIBUTE_TEMPORARY, NULL);
    if (hFile == INVALID_HANDLE_VALUE) {
        return 1;
    }
    DWORD bytesWritten;
    if (!WriteFile(hFile, pMsiData, msiSize, &bytesWritten, NULL) || bytesWritten != msiSize) {
        CloseHandle(hFile);
        return 1;
    }
    CloseHandle(hFile);
    MsiSetInternalUI(INSTALLUILEVEL_FULL, NULL);
    TCHAR cmd[500];
    wsprintf(cmd, L"ProductLanguage=%d PREVINSTALLFOLDER=\"%s\"", lang->id, installPath);
    return MsiInstallProduct(msiFile, cmd);
}
