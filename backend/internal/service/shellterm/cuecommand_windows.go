package shellterm

import (
	"crypto/rand"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

func newCueCommandDirectory(root string) (string, error) {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return "", err
	}
	defer func() { _ = token.Close() }()
	user, err := token.GetTokenUser()
	if err != nil {
		return "", err
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;" + user.User.Sid.String() + ")")
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "command-"+rand.Text())
	ptr, err := windows.UTF16PtrFromString(dir)
	if err != nil {
		return "", err
	}
	attributes := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	if err := windows.CreateDirectory(ptr, &attributes); err != nil {
		return "", err
	}
	return dir, nil
}
