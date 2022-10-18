package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	swissknife "github.com/Sagleft/swiss-knife"
	utopiago "github.com/Sagleft/utopialib-go"
)

const (
	configPath = "config.json"
)

func main() {
	cfg := config{}
	if err := swissknife.ParseStructFromJSONFile(configPath, &cfg); err != nil {
		log.Fatalln(err)
	}

	runHealthchecks(cfg)
}

type config struct {
	Utopia                 utopiago.UtopiaClient `json:"utopia"`
	ServiceName            string                `json:"serviceName"`
	AlsoRebootService      string                `json:"alsoRebotService"`
	SleepTimeoutSeconds    int                   `json:"sleepTimeoutSeconds"`
	WaitAfterRebootSeconds int                   `json:"waitAfterRebootSeconds"`
}

func runHealthchecks(cfg config) {
	for {
		if isProblemDetected(cfg.Utopia) {
			if err := doReboot(cfg); err != nil {
				log.Printf("failed to reboot %s: %s", cfg.ServiceName, err.Error())
			}
		}
		time.Sleep(time.Duration(cfg.SleepTimeoutSeconds) * time.Second)
	}
}

func isProblemDetected(u utopiago.UtopiaClient) bool {
	if _, err := u.GetSystemInfo(); err != nil {
		return true
	}
	return false
}

func rebootService(serviceName string) error {
	r := exec.Command("/usr/bin/systemctl", "restart", serviceName)
	if err := r.Run(); err != nil {
		return fmt.Errorf("failed to reboot %s: %w", serviceName, err)
	}
	return nil
}

func doReboot(cfg config) error {
	// reboot utopia
	if err := rebootService(cfg.ServiceName); err != nil {
		return err
	}

	if cfg.WaitAfterRebootSeconds > 0 {
		time.Sleep(time.Duration(cfg.WaitAfterRebootSeconds) * time.Second)
	}

	// reboot another service
	if cfg.AlsoRebootService != "" {
		if err := rebootService(cfg.AlsoRebootService); err != nil {
			return err
		}
	}
	return nil
}
