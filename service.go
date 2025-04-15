package winservice

import (
	"errors"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// Windows API definitions
var (
	modAdvapi32                  = windows.NewLazySystemDLL("advapi32.dll")
	procOpenSCManager            = modAdvapi32.NewProc("OpenSCManagerW")
	procOpenService              = modAdvapi32.NewProc("OpenServiceW")
	procQueryServiceStatus       = modAdvapi32.NewProc("QueryServiceStatus")
	procSetServiceObjectSecurity = modAdvapi32.NewProc("SetServiceObjectSecurity")
)

// Common errors for service operations
var (
	ErrServiceManagerConnectFail   = errors.New("service manager connect fail")
	ErrServiceOpenFail             = errors.New("service open fail")
	ErrServiceCreateFail           = errors.New("service create fail")
	ErrServiceStartFail            = errors.New("service start fail")
	ErrServiceStopFail             = errors.New("service stop fail")
	ErrServiceDeleteFail           = errors.New("service delete fail")
	ErrServiceStatusGetFail        = errors.New("service status get fail")
	ErrServiceStopTimeout          = errors.New("service stop timeout")
	ErrServiceAlreadyExists        = errors.New("service is already exist")
	ErrServiceRecoveryStrategyFail = errors.New("service recovery strategy set fail")
	ErrServiceSecuritySetFail      = errors.New("service security set fail")
)

const (
	SERVICE_QUERY_STATUS     uint32 = 0x0004
	SERVICE_QUERY_CONFIG     uint32 = 0x0001
	SERVICE_CHANGE_CONFIG    uint32 = 0x0002
	SERVICE_START            uint32 = 0x0010
	SERVICE_STOP             uint32 = 0x0020
	SERVICE_ALL_ACCESS       uint32 = 0xF01FF
	SERVICE_STOPPED          uint32 = 0x00000001
	SERVICE_START_PENDING    uint32 = 0x00000002
	SERVICE_STOP_PENDING     uint32 = 0x00000003
	SERVICE_RUNNING          uint32 = 0x00000004
	SERVICE_CONTINUE_PENDING uint32 = 0x00000005
	SERVICE_PAUSE_PENDING    uint32 = 0x00000006
	SERVICE_PAUSED           uint32 = 0x00000007

	// Default timeout for service operations
	defaultServiceTimeout = 10 * time.Second
	// Default retry interval
	defaultRetryInterval = 300 * time.Millisecond
	hiddenServiceSDDL    = "D:(A;;CCLCSWRPWPDTLOCRRC;;;SY)(A;;CCDCLCSWRPWPDTLOCRSDRCWDWO;;;BA)"
	defaultServiceSDDL   = "D:(A;;CCLCSWRPWPDTLOCRRC;;;SY)(A;;CCDCLCSWRPWPDTLOCRSDRCWDWO;;;BA)(A;;CCLCSWLOCRRC;;;IU)(A;;CCLCSWLOCRRC;;;SU)"
)

type SERVICE_STATUS struct {
	ServiceType             uint32
	ServiceState            uint32
	ControlsAccepted        uint32
	Win32ExitCode           uint32
	ServiceSpecificExitCode uint32
	CheckPoint              uint32
	WaitHint                uint32
}

// openSCManager opens a connection to the service control manager
func openSCManager() (windows.Handle, error) {
	handle, _, err := procOpenSCManager.Call(0, 0, windows.SC_MANAGER_ALL_ACCESS)
	if handle == 0 {
		return 0, wrapError(ErrServiceManagerConnectFail, err.Error())
	}
	return windows.Handle(handle), nil
}

// openService opens a handle to an existing service
func openService(scm windows.Handle, serviceName string, desiredAccess uint32) (windows.Handle, error) {
	handle, _, err := procOpenService.Call(
		uintptr(scm),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(serviceName))),
		uintptr(desiredAccess),
	)
	if handle == 0 {
		return 0, wrapError(ErrServiceOpenFail, err.Error())
	}
	return windows.Handle(handle), nil
}

// CreateService registers a new Windows service
func CreateService(serviceName, binPath string, autostart bool) error {
	// Connect to the service control manager
	m, err := mgr.Connect()
	if err != nil {
		return wrapError(ErrServiceManagerConnectFail, err.Error())
	}
	defer m.Disconnect()

	//  Check if the service already exists
	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return wrapError(ErrServiceAlreadyExists, serviceName)
	}

	// Set startup type based on autostart flag
	startType := mgr.StartManual
	if autostart {
		startType = mgr.StartAutomatic
	}

	// Create the service
	s, err = m.CreateService(
		serviceName,
		binPath,
		mgr.Config{
			DisplayName:  serviceName,
			StartType:    uint32(startType),
			ErrorControl: mgr.ErrorNormal,
		})
	if err != nil {
		return wrapError(ErrServiceCreateFail, err.Error())
	}
	defer s.Close()

	// Set recovery actions (restart on failure)
	if err = setRecoveryActions(serviceName); err != nil {
		return err
	}

	return nil
}

// setRecoveryActions configures a service to restart on failure
func setRecoveryActions(serviceName string) error {
	// Connect to the service control manager
	m, err := mgr.Connect()
	if err != nil {
		return wrapError(ErrServiceManagerConnectFail, err.Error())
	}
	defer m.Disconnect()

	// Open the target service
	s, err := m.OpenService(serviceName)
	if err != nil {
		return wrapError(ErrServiceOpenFail, err.Error())
	}
	defer s.Close()

	// Configure restart actions
	actions := []mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 15 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 15 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 15 * time.Second},
	}

	if err = s.SetRecoveryActions(actions, 60); err != nil {
		return wrapError(ErrServiceRecoveryStrategyFail, err.Error())
	}

	return nil
}

// StartService starts an existing Windows service
func StartService(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return wrapError(ErrServiceManagerConnectFail, err.Error())
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return wrapError(ErrServiceOpenFail, err.Error())
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return wrapError(ErrServiceStartFail, err.Error())
	}
	return nil
}

// StopService stops a running Windows service with timeout
func StopService(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return wrapError(ErrServiceManagerConnectFail, err.Error())
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return wrapError(ErrServiceOpenFail, err.Error())
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		return wrapError(ErrServiceStopFail, err.Error())
	}

	// Wait until the service is fully stopped or timeout
	timeout := time.Now().Add(10 * time.Second)
	for status.State != svc.Stopped {
		if time.Now().After(timeout) {
			return ErrServiceStopTimeout
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return wrapError(ErrServiceStatusGetFail, err.Error())
		}
	}
	return nil
}

// DeleteService removes a Windows service

func DeleteService(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return wrapError(ErrServiceManagerConnectFail, err.Error())
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return wrapError(ErrServiceOpenFail, err.Error())
	}
	defer s.Close()

	if err := s.Delete(); err != nil {
		return wrapError(ErrServiceDeleteFail, err.Error())
	}
	return nil
}

// QueryServiceStatus returns the current state of a service
func QueryServiceStatus(serviceName string) (uint32, error) {
	scm, err := openSCManager()
	if err != nil {
		return 0, err
	}
	defer windows.Close(scm)

	service, err := openService(scm, serviceName, SERVICE_QUERY_STATUS|SERVICE_QUERY_CONFIG)
	if err != nil {
		return 0, err
	}
	defer windows.Close(service)

	var status SERVICE_STATUS
	ret, _, err := procQueryServiceStatus.Call(
		uintptr(service),
		uintptr(unsafe.Pointer(&status)),
	)
	if ret == 0 {
		return 0, wrapError(ErrServiceStatusGetFail, err.Error())
	}

	return status.ServiceState, nil
}

// ServiceExists checks if a service with the given name exists
func ServiceExists(serviceName string) (bool, error) {
	scm, err := openSCManager()
	if err != nil {
		return false, err
	}
	defer windows.Close(scm)

	service, err := openService(scm, serviceName, SERVICE_QUERY_STATUS)
	if err != nil {
		if errors.Is(err, ErrServiceOpenFail) {
			return false, nil // Service doesn't exist
		}
		return false, err
	}
	defer windows.Close(service)

	return true, nil
}
