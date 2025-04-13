package winservice

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

// DACL_SECURITY_INFORMATION Security descriptor information flags
const (
	DACL_SECURITY_INFORMATION = 0x00000004
)

// SetServiceHidden makes a service hidden from regular users
func SetServiceHidden(serviceName string) error {
	return setServiceSDDL(serviceName, hiddenServiceSDDL)
}

// SetServiceUnHidden makes a service visible to regular users
func SetServiceUnHidden(serviceName string) error {
	return setServiceSDDL(serviceName, defaultServiceSDDL)
}

// setServiceSDDL modifies the service's Discretionary Access Control List
func setServiceSDDL(serviceName, sddl string) error {
	m, err := mgr.Connect()
	if err != nil {
		return wrapError(ErrServiceManagerConnectFail, err.Error())
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return wrapError(ErrServiceOpenFail, "opening service")
	}
	defer s.Close()

	sd, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return wrapError(ErrServiceSecuritySetFail, fmt.Sprintf("parsing SDDL: %s", err))
	}

	// Call the Windows API to set the security descriptor
	r1, _, e1 := procSetServiceObjectSecurity.Call(
		uintptr(s.Handle),
		uintptr(DACL_SECURITY_INFORMATION),
		uintptr(unsafe.Pointer(sd)),
	)
	if r1 == 0 {
		return wrapError(ErrServiceSecuritySetFail, fmt.Sprintf("setting security: %s", e1))
	}
	return nil
}
