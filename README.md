# WinService Manager (Go)

A lightweight Go library to manage Windows services, including creation, deletion, start/stop, auto-restart configuration, visibility control (hide/unhide), and service status query.

## ✨ Features

- Create and delete Windows services
- Start and stop services
- Set service auto-recovery policies
- Hide/unhide services via SDDL
- Query service status
- Detect if service exists
- Written in pure Go, using [`golang.org/x/sys/windows`](https://pkg.go.dev/golang.org/x/sys/windows)

## 🛠 Installation

```bash
go get github.com/Firstnsnd/winservice
```

## 🚀 Usage

> ⚠ Requires administrator privileges.

### Create, Start and Hide Service

```go

package main

import "github.com/Firstnsnd/winservice"

func main() {
    name := "MyHiddenService"
    exe := "C:\\path\\to\\my-service.exe"

    err := winservice.CreateService(name, exe, true)
    if err != nil {
        panic(err)
    }

    _ = winservice.SetServiceHidden(name)
    _ = winservice.StartService(name)
}
```

### Stop, Unhide and Delete Service

```go
_ = winservice.StopService("MyHiddenService")
_ = winservice.SetServiceUnHidden("MyHiddenService")
_ = winservice.DeleteService("MyHiddenService")
```

### Check Service Existence and Status

```go
exists, _ := winservice.ServiceExists("MyHiddenService")
status, _ := winservice.QueryServiceStatus("MyHiddenService")
```

### Set Recovery Policy

```go
_ = winservice.SetRecoveryActions("MyHiddenService")
```

## 🧪 Testing

Run the unit test (requires admin permission):

```bash
go test -v ./...
```

Or specifically:

```bash
go test -run TestFullServiceLifecycle
```

## 🧩 Error Handling

All exported functions return wrapped `error`s. You can use `errors.Is` for specific error types:

```
err := winservice.StartService("svc")
if errors.Is(err, winservice.ErrServiceStartFail) {
	log.Println("Start failed with known reason")
}
```

## 📁 Project Structure

```

winservice/
├── service.go         # Core service operations
├── sddl.go            # Service hiding via SDDL
├── utils.go           # Some helper functions
└── service_test.go    # Full lifecycle test
```

## ✅ Requirements

- Windows OS
- Go 1.16+

## 📄 License

Apache License 2.0