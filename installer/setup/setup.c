#define UNICODE
#define _UNICODE

#include <windows.h>
#include <msi.h>
#include <shlwapi.h>
#include <sddl.h>
#include <stdio.h>
#include "resource.h"

static WCHAR msiPath[MAX_PATH];
static HANDLE msiFile = INVALID_HANDLE_VALUE;

typedef struct {
    WCHAR id[5];
    WCHAR name[16];
    WCHAR code[10];
} Language;

typedef struct {
    WCHAR path[MAX_PATH];
    DWORD pathLen;
    WCHAR lang[10];
    DWORD langLen;
    WCHAR version[20];
    DWORD versionLen;
} Product;

static Language languages[] = {
    {L"2052", L"简体中文", L"zh-CN"},
    {L"1028", L"繁體中文", L"zh-TW"},
    {L"1033", L"English", L"en-US"},
    {L"1041", L"日本語", L"ja-JP"},
    {L"1042", L"한국어", L"ko-KR"},
    {L"3082", L"Español", L"es-ES"},
};

static INT MatchLanguageCode(LPWSTR langCode)
{
    for (size_t i = 0; i < _countof(languages); i++)
    {
        if (wcscmp(languages[i].code, langCode) == 0)
            return i;
    }
    return -1;
}

static INT GetApplicationLanguage(LPWSTR path, DWORD pathLen)
{
    if (!PathAppendW(path, L"lang.config"))
        return -1;
    DWORD bytesRead = 0;
    HANDLE hFile = CreateFileW(path, GENERIC_READ, 0, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
    path[pathLen] = L'\0';
    if (hFile != INVALID_HANDLE_VALUE)
    {
        CHAR buf[LOCALE_NAME_MAX_LENGTH];
        WCHAR localeName[LOCALE_NAME_MAX_LENGTH];
        BOOL ok = ReadFile(hFile, buf, sizeof(buf) - 1, &bytesRead, NULL);
        CloseHandle(hFile);
        if (ok && bytesRead != 0)
        {
            buf[bytesRead] = 0;
            if (MultiByteToWideChar(CP_UTF8, 0, buf, -1, localeName, _countof(localeName)) > 0)
            {
                INT i = MatchLanguageCode(localeName);
                if (i >= 0)
                    return i;
            }
        }
    }
    if (!PathAppendW(path, L"app.json"))
        return -1;
    hFile = CreateFileW(path, GENERIC_READ, 0, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
    path[pathLen] = L'\0';
    if (hFile == INVALID_HANDLE_VALUE)
        return -1;
    CHAR buf[100];
    // To avoid JSON dependency, we require the first field to be the language setting.
    static const CHAR* langKey = "{\"lang\":\"*\"";
    WCHAR langCode[10];
    DWORD langCodeLen = 0;
    INT j = 0;
    while (ReadFile(hFile, buf, sizeof(buf), &bytesRead, NULL) && bytesRead != 0)
    {
        for (DWORD i = 0; i < bytesRead; i++)
        {
            if (langKey[j] == '*')
            {
                if (buf[i] == '"')
                    j++;
                else
                {
                    langCode[langCodeLen++] = buf[i];
                    if (langCodeLen >= sizeof(langCode) - 1)
                        goto out;
                    continue;
                }
            }
            if (buf[i] == langKey[j])
            {
                j++;
                if (langKey[j] == 0)
                    goto out;
            }
            else if (buf[i] != '\t' && buf[i] != ' ' && buf[i] != '\r' && buf[i] != '\n')
                goto out;
            else if (langKey[j] != '{' && langKey[j] != ':' && j > 0 && langKey[j - 1] != '{' && langKey[j - 1] != ':')
                goto out;
        }
    }

out:
    CloseHandle(hFile);
    if (langKey[j] != 0 || langCodeLen == 0)
        return -1;
    langCode[langCodeLen] = 0;
    return MatchLanguageCode(langCode);
}

INT_PTR CALLBACK LanguageDialog(HWND hDlg, UINT message, WPARAM wParam, LPARAM lParam)
{
    switch (message) {
    case WM_INITDIALOG:
        for (size_t i = 0; i < _countof(languages); i++)
            SendDlgItemMessageW(hDlg, IDC_LANG_COMBO, CB_ADDSTRING, 0, (LPARAM)languages[i].name);
        SendDlgItemMessageW(hDlg, IDC_LANG_COMBO, CB_SETCURSEL, lParam, 0);
        return (INT_PTR)TRUE;

    case WM_COMMAND:
        INT_PTR nResult = LOWORD(wParam);
        if (nResult == IDOK || nResult == IDCANCEL)
        {
            if (nResult == IDOK)
            {
                LRESULT i = SendDlgItemMessageW(hDlg, IDC_LANG_COMBO, CB_GETCURSEL, 0, 0);
                nResult = (i >= 0 && i < _countof(languages)) ? (INT_PTR)&languages[i] : 0;
            }
            else
                nResult = 0;
            EndDialog(hDlg, nResult);
            return (INT_PTR)TRUE;
        }
        break;
    }
    return (INT_PTR)FALSE;
}

static int cleanup(void)
{
    if (msiFile != INVALID_HANDLE_VALUE)
    {
        CloseHandle(msiFile);
        msiFile = INVALID_HANDLE_VALUE;
    }
    for (INT i = 0; i < 200 && !DeleteFileW(msiPath) && GetLastError() != ERROR_FILE_NOT_FOUND; i++)
        Sleep(200);
    return 0;
}

int WINAPI WinMain(HINSTANCE hInstance, HINSTANCE hPrevInstance, LPSTR lpCmdLine, int nShowCmd)
{
    INT langIndex = -1;
    BOOL installed = FALSE, showDlg = TRUE;
    Product product = {
        .path = { 0 },
        .pathLen = _countof(product.path),
        .lang = { 0 },
        .langLen = _countof(product.lang),
        .version = { 0 },
        .versionLen = _countof(product.version)
    };

#ifdef UPGRADE_CODE
    WCHAR productCode[39];
    if (MsiEnumRelatedProductsW(UPGRADE_CODE, 0, 0, productCode) == ERROR_SUCCESS)
    {
        MsiGetProductInfo(productCode, INSTALLPROPERTY_VERSIONSTRING, product.version, &product.versionLen);
        if (MsiGetProductInfo(productCode, INSTALLPROPERTY_INSTALLLOCATION, product.path, &product.pathLen) == ERROR_SUCCESS && product.path[0])
            langIndex = GetApplicationLanguage(product.path, product.pathLen);
        if (MsiGetProductInfo(productCode, INSTALLPROPERTY_INSTALLEDLANGUAGE, product.lang, &product.langLen) == ERROR_SUCCESS && langIndex < 0)
        {
            for (size_t i = 0; i < _countof(languages); i++)
            {
                if (wcscmp(languages[i].id, product.lang) == 0)
                {
                    langIndex = i;
                    break;
                }
            }
        }
    }
#ifdef VERSION
    installed = wcscmp(VERSION, product.version) == 0;
    showDlg = !product.path[0] || installed || langIndex < 0;
#endif
#endif

    if (langIndex < 0)
    {
        PZZWSTR langList = NULL;
        ULONG langNum, langLen = 0;
        if (GetUserPreferredUILanguages(MUI_LANGUAGE_NAME, &langNum, NULL, &langLen))
        {
            langList = (PZZWSTR)LocalAlloc(LMEM_FIXED, langLen * sizeof(WCHAR));
            if (langList)
            {
                if (GetUserPreferredUILanguages(MUI_LANGUAGE_NAME, &langNum, langList, &langLen) && langNum > 0)
                {
                    for (size_t i = 0; i < langLen && langList[i] != L'\0'; i += wcsnlen_s(&langList[i], langLen - i) + 1)
                    {
                        langIndex = MatchLanguageCode(&langList[i]);
                        if (langIndex >= 0)
                            break;
                    }
                }
                LocalFree(langList);
            }
        }
    }

    Language* lang = showDlg ? (Language*)DialogBoxParamW(
        hInstance, MAKEINTRESOURCE(IDD_LANG_DIALOG),
        NULL, LanguageDialog, langIndex
    ) : &languages[langIndex];
    if (lang == NULL)
        return 0;

    if (!GetWindowsDirectoryW(msiPath, _countof(msiPath)) || !PathAppendW(msiPath, L"Temp"))
        return 1;
    GUID guid;
    if (FAILED(CoCreateGuid(&guid)))
        return 1;
    WCHAR identifier[40];
    if (StringFromGUID2(&guid, identifier, _countof(identifier)) == 0 || !PathAppendW(msiPath, identifier))
        return 1;

    HRSRC hRes = FindResourceW(NULL, MAKEINTRESOURCE(IDR_MSI), RT_RCDATA);
    if (hRes == NULL)
        return 1;
    HGLOBAL hResData = LoadResource(NULL, hRes);
    if (hResData == NULL)
        return 1;
    DWORD resSize = SizeofResource(NULL, hRes);
    if (resSize == 0)
        return 1;
    LPVOID pResData = LockResource(hResData);
    if (pResData == NULL)
        return 1;

    SECURITY_ATTRIBUTES sa = { .nLength = sizeof(sa) };
    if (!ConvertStringSecurityDescriptorToSecurityDescriptorA("O:BAD:PAI(A;;FA;;;BA)", SDDL_REVISION_1, &sa.lpSecurityDescriptor, NULL))
        return 1;
    msiFile = CreateFileW(msiPath, GENERIC_WRITE, 0, &sa, CREATE_NEW, FILE_ATTRIBUTE_TEMPORARY, NULL);
    if (sa.lpSecurityDescriptor)
        LocalFree(sa.lpSecurityDescriptor);
    if (msiFile == INVALID_HANDLE_VALUE)
        return 1;
    _onexit(cleanup);
    DWORD bytesWritten;
    BOOL ok = WriteFile(msiFile, pResData, resSize, &bytesWritten, NULL);
    CloseHandle(msiFile);
    msiFile = INVALID_HANDLE_VALUE;
    if (!ok || bytesWritten != resSize)
        return 1;

    MsiSetInternalUI(INSTALLUILEVEL_FULL, NULL);
#define CMD_FORMAT L"ProductLanguage=%s PREVINSTALLFOLDER=\"%s\""
    WCHAR cmd[_countof(CMD_FORMAT) + _countof(product.path)];
    if (swprintf_s(cmd, _countof(cmd), CMD_FORMAT, lang->id, product.path) < 0)
        return 1;
    return MsiInstallProductW(msiPath, cmd);
}
