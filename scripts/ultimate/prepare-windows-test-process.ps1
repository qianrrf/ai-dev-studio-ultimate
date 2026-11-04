# GitHub's elevated Windows runner otherwise makes Administrators the default
# owner of every fixture, unlike the ordinary desktop user being tested.
$ErrorActionPreference = 'Stop'
Add-Type @'
using System;
using System.ComponentModel;
using System.Runtime.InteropServices;
public static class UltimateTestOwner {
    [DllImport("advapi32.dll", SetLastError=true)] static extern bool OpenProcessToken(IntPtr process, uint access, out IntPtr token);
    [DllImport("advapi32.dll", SetLastError=true)] static extern bool GetTokenInformation(IntPtr token, int kind, IntPtr value, int length, out int needed);
    [DllImport("advapi32.dll", SetLastError=true)] static extern bool SetTokenInformation(IntPtr token, int kind, ref IntPtr value, int length);
    [DllImport("kernel32.dll")] static extern IntPtr GetCurrentProcess();
    [DllImport("kernel32.dll")] static extern bool CloseHandle(IntPtr handle);
    public static void SetCurrentUser() {
        IntPtr token;
        if (!OpenProcessToken(GetCurrentProcess(), 0x0088, out token)) throw new Win32Exception();
        IntPtr user = IntPtr.Zero;
        try {
            int needed;
            GetTokenInformation(token, 1, IntPtr.Zero, 0, out needed);
            user = Marshal.AllocHGlobal(needed);
            if (!GetTokenInformation(token, 1, user, needed, out needed)) throw new Win32Exception();
            IntPtr sid = Marshal.ReadIntPtr(user);
            if (!SetTokenInformation(token, 4, ref sid, IntPtr.Size)) throw new Win32Exception();
        } finally {
            if (user != IntPtr.Zero) Marshal.FreeHGlobal(user);
            CloseHandle(token);
        }
    }
}
'@
[UltimateTestOwner]::SetCurrentUser()
$env:TEMP = Join-Path $env:USERPROFILE 'AppData\Local\Temp\ultimate-validation'
$env:TMP = $env:TEMP
New-Item -ItemType Directory -Path $env:TEMP -Force | Out-Null
$fixture = New-Item -ItemType Directory -Path (Join-Path $env:TEMP ([guid]::NewGuid().ToString()))
try {
    $expectedOwner = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value
    $actualOwner = (Get-Acl -LiteralPath $fixture.FullName).GetOwner([System.Security.Principal.SecurityIdentifier]).Value
    if ($actualOwner -ne $expectedOwner) { throw "Test fixture owner mismatch: $actualOwner" }
} finally {
    Remove-Item -LiteralPath $fixture.FullName
}
