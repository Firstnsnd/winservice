// myGoService.go
package main

import (
	"fmt"
	"github.com/kardianos/service"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logger service.Logger

var _ service.Interface = (*Program)(nil)

// Program implement service.Interface
type Program struct{}

// Start service start
func (p *Program) Start(s service.Service) error {

	go p.run()
	return nil
}

// Stop service stop execute
func (p *Program) Stop(s service.Service) error {
	return nil
}

// service run action
func (p *Program) run() {
	// create log file
	logDir := filepath.Join(filepath.Dir(os.Args[0]), "logs")
	_ = os.MkdirAll(logDir, 0755)
	logFile := filepath.Join(logDir, "TestWinSvc.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("open log faile: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// loop
	for {
		log.Println("TestWinSvc is running...")
		time.Sleep(1 * time.Minute)
	}

}

// service config
func getServiceConfig() *service.Config {
	return &service.Config{
		Name:        "TestWinSvc",
		DisplayName: "Test Win Svc Demo",
		Description: "A demo use kardianos/service develop eg service",
	}
}

func main() {
	svcConfig := getServiceConfig()
	prg := &Program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	// if args exist，will execute it
	if len(os.Args) > 1 {
		err = service.Control(s, os.Args[1])
		if err != nil {
			fmt.Printf("service control execute %s fail: %v\n", os.Args[1], err)
		}
		return
	}

	// start service
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
