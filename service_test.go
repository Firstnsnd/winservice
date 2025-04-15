package winservice

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testServiceName = "TestWinSvc"
)

func TestFullServiceLifecycle(t *testing.T) {
	require.True(t, IsAdmin(), "Test must be run as administrator")

	exePath := filepath.Join("example", "TestWinSvc.exe")
	exeAbsPath, err := filepath.Abs(exePath)
	require.NoError(t, err)

	// make sure file exist
	_, err = os.Stat(exeAbsPath)
	require.NoError(t, err, "Executable not found at: "+exeAbsPath)
	// 1. Create the service
	err = CreateService(testServiceName, exeAbsPath, true)
	require.NoError(t, err, "Failed to create service")

	// 2. Set recovery actions
	err = setRecoveryActions(testServiceName)
	require.NoError(t, err, "Failed to set service recovery policy")

	// 3. Hide the service
	err = SetServiceHidden(testServiceName)
	require.NoError(t, err, "Failed to hide service")

	// 4. Start the service
	err = StartService(testServiceName)
	require.NoError(t, err, "Failed to start service")

	// 5. Verify service existence
	exists, err := ServiceExists(testServiceName)
	require.NoError(t, err, "Failed to check if service exists")
	require.True(t, exists, "Service should exist")

	// 6. Query service status
	status, err := QueryServiceStatus(testServiceName)
	require.NoError(t, err, "Failed to query service status")
	require.Equal(t, SERVICE_RUNNING, status, "Service should be running")

	// 7. Stop the service
	err = StopService(testServiceName)
	require.NoError(t, err, "Failed to stop service")

	// 8. Unhide the service
	err = SetServiceUnHidden(testServiceName)
	require.NoError(t, err, "Failed to unhide service")

	// 9. Delete the service
	err = DeleteService(testServiceName)
	require.NoError(t, err, "Failed to delete service")
}
