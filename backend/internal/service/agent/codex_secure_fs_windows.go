//go:build windows

package agent

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	codexWindowsAccessDeniedObjectACEType         uint8 = 0x06
	codexWindowsAccessDeniedCallbackACEType       uint8 = 0x0a
	codexWindowsAccessDeniedCallbackObjectACEType uint8 = 0x0c
)

type codexWindowsACLPolicy uint8

const (
	codexWindowsACLPolicyAncestor codexWindowsACLPolicy = iota
	codexWindowsACLPolicyPrivate
	codexWindowsACLPolicyDeviceCredential
)

func codexPrivateFileMode(os.FileInfo) bool { return true }

func openCodexFileNoFollow(path string) (*os.File, error) {
	handle, info, security, err := openCodexWindowsPath(path, false, true)
	if err != nil {
		return nil, err
	}
	if !codexWindowsPathMetadataIsSafe(codexWindowsMetadata(info, security.ownerCurrent, security.aclSafe), false, true) {
		_ = windows.CloseHandle(handle)
		return nil, errors.New("codex file handle is unsafe")
	}
	return os.NewFile(uintptr(handle), filepath.Base(path)), nil
}

func openCodexDeviceFileNoFollow(path string) (*os.File, error) {
	handle, info, security, err := openCodexWindowsPathWithSecurity(
		path,
		false,
		windows.GENERIC_READ|windows.READ_CONTROL,
		codexWindowsACLPolicyDeviceCredential,
	)
	if err != nil {
		return nil, err
	}
	if !codexWindowsPathMetadataIsSafe(codexWindowsMetadata(info, security.ownerTrusted, true), false, true) {
		_ = windows.CloseHandle(handle)
		return nil, errors.New("codex device credential handle is unsafe")
	}
	if !security.aclSafe {
		repairable, repairErr := codexWindowsHandleDeviceCredentialRepairable(handle)
		if repairErr != nil || !repairable {
			_ = windows.CloseHandle(handle)
			return nil, errors.New("codex device credential ACL is unsafe")
		}
		repairHandle, repairInfo, repairSecurity, repairErr := openCodexWindowsPathWithSecurityAndShare(
			path,
			false,
			windows.GENERIC_READ|windows.READ_CONTROL|windows.WRITE_DAC,
			codexWindowsACLPolicyDeviceCredential,
			windows.FILE_SHARE_READ,
		)
		if repairErr != nil {
			_ = windows.CloseHandle(handle)
			return nil, repairErr
		}
		if !repairSecurity.ownerCurrent || !codexWindowsPathMetadataIsSafe(codexWindowsMetadata(repairInfo, repairSecurity.ownerTrusted, true), false, true) ||
			!codexWindowsSameStableIdentity(codexWindowsMetadata(info, true, true), codexWindowsMetadata(repairInfo, true, true)) {
			_ = windows.CloseHandle(repairHandle)
			_ = windows.CloseHandle(handle)
			return nil, errors.New("codex device credential changed before ACL repair")
		}
		if repairErr := setCodexWindowsDeviceCredentialDACL(repairHandle); repairErr != nil {
			_ = windows.CloseHandle(repairHandle)
			_ = windows.CloseHandle(handle)
			return nil, repairErr
		}
		_ = windows.CloseHandle(handle)
		handle = repairHandle
	}
	return os.NewFile(uintptr(handle), filepath.Base(path)), nil
}

func validateCodexDirectory(path string, requirePrivate bool) error {
	handle, info, security, err := openCodexWindowsPath(path, true, requirePrivate)
	if err != nil {
		return errors.New("codex directory is unsafe")
	}
	_ = windows.CloseHandle(handle)
	if !codexWindowsPathMetadataIsSafe(codexWindowsMetadata(info, security.ownerCurrent, security.aclSafe), true, false) {
		return errors.New("codex directory owner or ACL is unsafe")
	}
	return validateCodexDirectoryAncestors(path)
}

func validateCodexDirectoryAncestors(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return errors.New("codex directory path is invalid")
	}
	for current := abs; ; current = filepath.Dir(current) {
		ptr, ptrErr := windows.UTF16PtrFromString(current)
		if ptrErr != nil {
			return errors.New("codex directory path is invalid")
		}
		attributes, attrErr := windows.GetFileAttributes(ptr)
		if errors.Is(attrErr, windows.ERROR_FILE_NOT_FOUND) || errors.Is(attrErr, windows.ERROR_PATH_NOT_FOUND) {
			if parent := filepath.Dir(current); parent != current {
				continue
			}
			return errors.New("codex directory has no trusted ancestor")
		}
		if attrErr != nil || attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || attributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
			return errors.New("codex directory has an unsafe ancestor")
		}
		handle, info, security, openErr := openCodexWindowsPath(current, true, false)
		if openErr != nil {
			return errors.New("codex directory ancestor could not be verified")
		}
		_ = windows.CloseHandle(handle)
		if !codexWindowsPathMetadataIsSafe(codexWindowsMetadata(info, security.ownerTrusted, security.aclSafe), true, false) {
			return errors.New("codex directory ancestor ACL is unsafe")
		}
		if parent := filepath.Dir(current); parent == current {
			return nil
		}
	}
}

type codexWindowsSecurity struct {
	ownerCurrent bool
	ownerTrusted bool
	aclSafe      bool
}

func openCodexWindowsPath(path string, directory, requirePrivate bool) (windows.Handle, windows.ByHandleFileInformation, codexWindowsSecurity, error) {
	return openCodexWindowsPathWithAccess(path, directory, windows.GENERIC_READ|windows.READ_CONTROL, requirePrivate)
}

func openCodexWindowsPathWithAccess(path string, directory bool, access uint32, requirePrivate bool) (windows.Handle, windows.ByHandleFileInformation, codexWindowsSecurity, error) {
	policy := codexWindowsACLPolicyAncestor
	if requirePrivate {
		policy = codexWindowsACLPolicyPrivate
	}
	return openCodexWindowsPathWithSecurity(path, directory, access, policy)
}

func openCodexWindowsPathWithSecurity(path string, directory bool, access uint32, policy codexWindowsACLPolicy) (windows.Handle, windows.ByHandleFileInformation, codexWindowsSecurity, error) {
	return openCodexWindowsPathWithSecurityAndShare(
		path,
		directory,
		access,
		policy,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
	)
}

func openCodexWindowsPathWithSecurityAndShare(path string, directory bool, access uint32, policy codexWindowsACLPolicy, shareMode uint32) (windows.Handle, windows.ByHandleFileInformation, codexWindowsSecurity, error) {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return windows.InvalidHandle, windows.ByHandleFileInformation{}, codexWindowsSecurity{}, err
	}
	handle, err := windows.CreateFile(
		ptr,
		access,
		shareMode,
		nil,
		windows.OPEN_EXISTING,
		codexWindowsNoFollowOpenFlags(),
		0,
	)
	if err != nil {
		return windows.InvalidHandle, windows.ByHandleFileInformation{}, codexWindowsSecurity{}, err
	}
	fail := func(err error) (windows.Handle, windows.ByHandleFileInformation, codexWindowsSecurity, error) {
		_ = windows.CloseHandle(handle)
		return windows.InvalidHandle, windows.ByHandleFileInformation{}, codexWindowsSecurity{}, err
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return fail(err)
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || (info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0) != directory {
		return fail(errors.New("codex path is a reparse point or has the wrong type"))
	}
	ownerCurrent, ownerTrusted, aclSafe, err := codexWindowsHandleSecurityForPolicy(handle, policy)
	if err != nil {
		return fail(err)
	}
	return handle, info, codexWindowsSecurity{ownerCurrent: ownerCurrent, ownerTrusted: ownerTrusted, aclSafe: aclSafe}, nil
}

func codexWindowsMetadata(info windows.ByHandleFileInformation, ownerTrusted, aclSafe bool) codexWindowsPathMetadata {
	return codexWindowsPathMetadata{
		Attributes: info.FileAttributes, HardLinks: info.NumberOfLinks, OwnerTrusted: ownerTrusted, ACLSafe: aclSafe,
		VolumeSerial: info.VolumeSerialNumber, FileIndexHigh: info.FileIndexHigh, FileIndexLow: info.FileIndexLow,
	}
}

func codexWindowsHandleSecurity(handle windows.Handle, requirePrivate bool) (bool, bool, bool, error) {
	policy := codexWindowsACLPolicyAncestor
	if requirePrivate {
		policy = codexWindowsACLPolicyPrivate
	}
	return codexWindowsHandleSecurityForPolicy(handle, policy)
}

func codexWindowsHandleSecurityForPolicy(handle windows.Handle, policy codexWindowsACLPolicy) (bool, bool, bool, error) {
	ownerCurrent, ownerTrusted, aces, err := codexWindowsHandleSecurityState(handle, policy == codexWindowsACLPolicyDeviceCredential)
	if err != nil {
		return false, false, false, err
	}
	aclSafe := ownerTrusted && codexWindowsAncestorACLIsSafe(aces)
	switch policy {
	case codexWindowsACLPolicyPrivate:
		aclSafe = codexWindowsVaultACLIsSafe(ownerTrusted, aces)
	case codexWindowsACLPolicyDeviceCredential:
		aclSafe = codexWindowsDeviceCredentialACLIsSafe(ownerTrusted, aces)
	}
	return ownerCurrent, ownerTrusted, aclSafe, nil
}

func codexWindowsHandleDeviceCredentialRepairable(handle windows.Handle) (bool, error) {
	ownerCurrent, _, aces, err := codexWindowsHandleSecurityState(handle, true)
	if err != nil {
		return false, err
	}
	return codexWindowsDeviceCredentialACLIsRepairable(ownerCurrent, aces), nil
}

func codexWindowsHandleSecurityState(handle windows.Handle, classifyDevicePrincipals bool) (bool, bool, []codexWindowsACE, error) {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return false, false, nil, err
	}
	defer func() { _ = token.Close() }()
	user, err := token.GetTokenUser()
	if err != nil {
		return false, false, nil, err
	}
	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return false, false, nil, err
	}
	administrators, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return false, false, nil, err
	}
	trustedInstaller, err := codexWindowsTrustedInstallerSID()
	if err != nil {
		return false, false, nil, err
	}
	sd, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil || sd == nil {
		return false, false, nil, errors.New("codex path security descriptor is unavailable")
	}
	owner, _, err := sd.Owner()
	if err != nil || owner == nil {
		return false, false, nil, errors.New("codex path owner is unavailable")
	}
	ownerCurrent := owner.Equals(user.User.Sid)
	ownerTrusted := ownerCurrent || owner.Equals(system) || owner.Equals(administrators) || owner.Equals(trustedInstaller)
	dacl, _, err := sd.DACL()
	if err != nil || dacl == nil {
		return ownerCurrent, ownerTrusted, nil, errors.New("codex path DACL is unavailable")
	}
	var sandbox *windows.SID
	if classifyDevicePrincipals {
		resolved, sandboxErr := codexWindowsLocalSandboxSID()
		if sandboxErr == nil {
			sandbox = resolved
		}
	}
	aces, err := codexWindowsDACLACEs(dacl, user.User.Sid, system, administrators, trustedInstaller, sandbox, classifyDevicePrincipals)
	if err != nil {
		return ownerCurrent, ownerTrusted, nil, err
	}
	return ownerCurrent, ownerTrusted, aces, nil
}

func codexWindowsDACLACEs(
	dacl *windows.ACL,
	user, system, administrators, trustedInstaller, sandbox *windows.SID,
	classifyDevicePrincipals bool,
) ([]codexWindowsACE, error) {

	aces := make([]codexWindowsACE, 0, dacl.AceCount)
	for index := uint32(0); index < uint32(dacl.AceCount); index++ {
		var raw *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, index, &raw); err != nil || raw == nil {
			return nil, errors.New("codex path ACL is unreadable")
		}

		if raw.Header.AceFlags&windows.INHERIT_ONLY_ACE != 0 {
			continue
		}
		switch raw.Header.AceType {
		case windows.ACCESS_ALLOWED_ACE_TYPE:
			ace := codexWindowsACE{Allowed: true, Mask: uint32(raw.Mask)}
			sid := (*windows.SID)(unsafe.Pointer(&raw.SidStart)) // #nosec G103 -- GetAce returns the Windows ACCESS_ALLOWED_ACE layout; SID validity is checked below.
			if !sid.IsValid() {
				return nil, errors.New("codex path ACL contains an invalid allow ACE")
			}
			ace.PrincipalTrusted = sid.Equals(user) || sid.Equals(system) || sid.Equals(administrators) || sid.Equals(trustedInstaller)
			ace.PrincipalSandbox = sandbox != nil && sid.Equals(sandbox)
			if classifyDevicePrincipals && !ace.PrincipalTrusted && !ace.PrincipalSandbox {
				_, _, _, lookupErr := sid.LookupAccount("")
				ace.PrincipalUnmapped = errors.Is(lookupErr, windows.ERROR_NONE_MAPPED)
			}
			aces = append(aces, ace)
		case windows.ACCESS_DENIED_ACE_TYPE,
			codexWindowsAccessDeniedObjectACEType,
			codexWindowsAccessDeniedCallbackACEType,
			codexWindowsAccessDeniedCallbackObjectACEType:
			// Deny ACEs cannot grant access. Their trustee layout is irrelevant to
			// these policies, so leave them to the Windows access check.
			continue
		default:
			// Compound, object, and callback allow ACEs have layouts or semantics
			// this decoder does not implement. Reject them rather than risk
			// treating an effective grant as a harmless non-allow ACE.
			return nil, errors.New("codex path ACL contains an unsupported effective ACE")
		}
	}
	return aces, nil
}

// codexWindowsTrustedInstallerSIDString is NT SERVICE\TrustedInstaller. Windows
// makes it the owner of the system drive root and of everything under
// %SystemRoot%, so it is an unavoidable ancestor owner for any path AO creates.
// It is a service SID that no interactive account can assume without already
// holding SYSTEM, which makes it at least as trustworthy as BUILTIN\Administrators.
const codexWindowsTrustedInstallerSIDString = "S-1-5-80-956008885-3418522649-1831038044-1853292631-2271478464"

func codexWindowsTrustedInstallerSID() (*windows.SID, error) {
	return windows.StringToSid(codexWindowsTrustedInstallerSIDString)
}

func codexWindowsLocalSandboxSID() (*windows.SID, error) {
	const capacity = 256
	nameBuffer := make([]uint16, capacity)
	nameLength := uint32(capacity)
	if err := windows.GetComputerNameEx(windows.ComputerNameNetBIOS, &nameBuffer[0], &nameLength); err != nil || nameLength == 0 {
		return nil, errors.New("local computer name is unavailable")
	}
	computerName := windows.UTF16ToString(nameBuffer[:nameLength])
	sid, domain, accountType, err := windows.LookupSID(computerName, "CodexSandboxUsers")
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(domain, computerName) || (accountType != windows.SidTypeAlias && accountType != windows.SidTypeGroup) {
		return nil, errors.New("CodexSandboxUsers is not a local group")
	}
	return sid, nil
}

func protectCodexPrivateDirectory(path string) error {
	handle, _, security, err := openCodexWindowsPathWithAccess(
		path,
		true,
		windows.WRITE_DAC|windows.READ_CONTROL,
		false,
	)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	if !security.ownerCurrent {
		return errors.New("codex private directory owner is unsafe")
	}
	return setCodexWindowsPrivateDACL(handle, true)
}

func protectCodexPrivateFile(path string, file *os.File) error {
	if file == nil {
		return errors.New("codex private file handle is unavailable")
	}
	var original windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(file.Fd()), &original); err != nil {
		return err
	}
	handle, opened, security, err := openCodexWindowsPathWithAccess(
		path,
		false,
		windows.WRITE_DAC|windows.READ_CONTROL,
		false,
	)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	if !security.ownerCurrent || !codexWindowsSameStableIdentity(
		codexWindowsMetadata(original, true, true),
		codexWindowsMetadata(opened, true, true),
	) {
		return errors.New("codex private file changed before ACL protection")
	}
	return setCodexWindowsPrivateDACL(handle, false)
}

func protectCodexDeviceCredentialFile(path string) error {
	handle, _, security, err := openCodexWindowsPathWithSecurity(
		path,
		false,
		windows.WRITE_DAC|windows.READ_CONTROL,
		codexWindowsACLPolicyDeviceCredential,
	)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	if !security.ownerCurrent {
		return errors.New("codex device credential owner is unsafe")
	}
	return setCodexWindowsDeviceCredentialDACL(handle)
}

func setCodexWindowsPrivateDACL(handle windows.Handle, directory bool) error {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return err
	}
	defer func() { _ = token.Close() }()
	user, err := token.GetTokenUser()
	if err != nil {
		return err
	}
	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return err
	}
	administrators, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return err
	}
	inheritance := uint32(windows.NO_INHERITANCE)
	if directory {
		inheritance = windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT
	}
	entry := func(sid *windows.SID, trusteeType windows.TRUSTEE_TYPE) windows.EXPLICIT_ACCESS {
		return windows.EXPLICIT_ACCESS{
			AccessPermissions: windows.GENERIC_ALL,
			AccessMode:        windows.SET_ACCESS,
			Inheritance:       inheritance,
			Trustee: windows.TRUSTEE{
				TrusteeForm:  windows.TRUSTEE_IS_SID,
				TrusteeType:  trusteeType,
				TrusteeValue: windows.TrusteeValueFromSID(sid),
			},
		}
	}
	dacl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{
		entry(user.User.Sid, windows.TRUSTEE_IS_USER),
		entry(system, windows.TRUSTEE_IS_USER),
		entry(administrators, windows.TRUSTEE_IS_WELL_KNOWN_GROUP),
	}, nil)
	if err != nil {
		return err
	}
	if err := windows.SetSecurityInfo(
		handle,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		dacl,
		nil,
	); err != nil {
		return err
	}
	ownerCurrent, _, aclSafe, err := codexWindowsHandleSecurity(handle, true)
	if err != nil || !ownerCurrent || !aclSafe {
		return errors.New("codex private ACL could not be verified")
	}
	sd, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil || sd == nil {
		return errors.New("codex private ACL protection is unavailable")
	}
	control, _, err := sd.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return errors.New("codex private ACL inheritance is not protected")
	}
	return nil
}

func setCodexWindowsDeviceCredentialDACL(handle windows.Handle) error {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return err
	}
	defer func() { _ = token.Close() }()
	user, err := token.GetTokenUser()
	if err != nil {
		return err
	}
	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return err
	}
	administrators, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return err
	}
	entry := func(sid *windows.SID, permissions windows.ACCESS_MASK, trusteeType windows.TRUSTEE_TYPE) windows.EXPLICIT_ACCESS {
		return windows.EXPLICIT_ACCESS{
			AccessPermissions: permissions,
			AccessMode:        windows.SET_ACCESS,
			Inheritance:       windows.NO_INHERITANCE,
			Trustee: windows.TRUSTEE{
				TrusteeForm:  windows.TRUSTEE_IS_SID,
				TrusteeType:  trusteeType,
				TrusteeValue: windows.TrusteeValueFromSID(sid),
			},
		}
	}
	entries := []windows.EXPLICIT_ACCESS{
		entry(user.User.Sid, windows.GENERIC_ALL, windows.TRUSTEE_IS_USER),
		entry(system, windows.GENERIC_ALL, windows.TRUSTEE_IS_USER),
		entry(administrators, windows.GENERIC_ALL, windows.TRUSTEE_IS_WELL_KNOWN_GROUP),
	}
	if sandbox, lookupErr := codexWindowsLocalSandboxSID(); lookupErr == nil {
		entries = append(entries, entry(sandbox, windows.GENERIC_READ, windows.TRUSTEE_IS_GROUP))
	}
	dacl, err := windows.ACLFromEntries(entries, nil)
	if err != nil {
		return err
	}
	if err := windows.SetSecurityInfo(
		handle,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		dacl,
		nil,
	); err != nil {
		return err
	}
	ownerCurrent, _, aclSafe, err := codexWindowsHandleSecurityForPolicy(handle, codexWindowsACLPolicyDeviceCredential)
	if err != nil || !ownerCurrent || !aclSafe {
		return errors.New("codex device credential ACL could not be verified")
	}
	sd, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil || sd == nil {
		return errors.New("codex device credential ACL protection is unavailable")
	}
	control, _, err := sd.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return errors.New("codex device credential ACL inheritance is not protected")
	}
	return nil
}

func syncDirectory(path string) error {
	handle, info, security, err := openCodexWindowsPathWithAccess(path, true, codexWindowsDirectoryFlushAccess(), false)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	if !codexWindowsPathMetadataIsSafe(codexWindowsMetadata(info, security.ownerCurrent, security.aclSafe), true, false) {
		return errors.New("codex directory owner or ACL is unsafe")
	}
	return windows.FlushFileBuffers(handle)
}
