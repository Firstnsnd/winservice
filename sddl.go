package winservice

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc/mgr"
)

// DACL_SECURITY_INFORMATION Security descriptor information flags
const (
	DACL_SECURITY_INFORMATION = 0x00000004
)

// IsServiceHidden check service is hidden
func IsServiceHidden(serviceName string) (bool, error) {
	// 1. Check if DisplayName exists in the registry
	keyPath := `SYSTEM\CurrentControlSet\Services\` + serviceName
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.READ)
	if err != nil {
		return false, fmt.Errorf("could not open registry key for service %s: %v", serviceName, err)
	}
	defer k.Close()

	// Check DisplayName
	_, _, err = k.GetStringValue("DisplayName")
	if errors.Is(err, registry.ErrNotExist) {
		// If DisplayName is not found, the service may be hidden
		return true, nil
	}

	// 2. Check if ImagePath is valid
	imagePath, _, err := k.GetStringValue("ImagePath")
	if err != nil {
		return false, fmt.Errorf("could not get ImagePath for service %s: %v", serviceName, err)
	}

	// Check if the file pointed by ImagePath exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		// If the file pointed by ImagePath does not exist, the service might be abnormal
		return true, nil
	}

	// 3. Use Get-Service to confirm if the service can be listed
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Get-Service -Name %s", serviceName))
	err = cmd.Run()
	if err != nil {
		// If Get-Service fails, the service may be hidden
		return true, nil
	}

	// If DisplayName exists, ImagePath is valid, and the service can be listed by Get-Service, it is not hidden
	return false, nil
}

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
