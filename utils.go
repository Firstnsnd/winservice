package winservice

import (
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
)

// IsAdmin 判断当前进程是否拥有管理员权限
func IsAdmin() bool {
	var sid *windows.SID
	// Create a SID for the Administrators group.
	sid, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return false
	}

	token := windows.Token(0)
	isMember, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return isMember
}

func wrapError(err error, str string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w, %s", err, str)
}

// wrapWinAPIError trans Windows error to  Go  error
func wrapWinAPIError(err error) error {
	if errno, ok := err.(syscall.Errno); ok {
		return fmt.Errorf("windows API err (%d): %s", errno, errno.Error())
	}
	return err
}
