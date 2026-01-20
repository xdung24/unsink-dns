//go:build linux
// +build linux

package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kardianos/service"
)

//go:embed unsinkdns.service
var systemdUnit string

func installService(s service.Service) error {
	// Copy current executable to /usr/local/bin/unsinkdns so ExecStart matches
	dest := "/usr/local/bin/unsinkdns"
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to locate current executable: %w", err)
	}
	exePath, err := filepath.Abs(exe)
	if err != nil {
		exePath = exe
	}

	if exePath != dest {
		data, err := os.ReadFile(exePath)
		if err != nil {
			return fmt.Errorf("failed to read executable: %w", err)
		}
		if err := os.WriteFile(dest, data, 0755); err != nil {
			return fmt.Errorf("failed to copy binary to %s: %w", dest, err)
		}
	}

	unitPath := "/etc/systemd/system/unsinkdns.service"
	if err := os.WriteFile(unitPath, []byte(systemdUnit), 0644); err != nil {
		return fmt.Errorf("failed to write unit file: %w", err)
	}

	if out, err := exec.Command("systemctl", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl daemon-reload failed: %v: %s", err, string(out))
	}

	if out, err := exec.Command("systemctl", "enable", "--now", "unsinkdns.service").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl enable/start failed: %v: %s", err, string(out))
	}

	return nil
}

func removeService(s service.Service) error {
	// stop and disable the service (ignore errors)
	_ = exec.Command("systemctl", "stop", "unsinkdns.service").Run()
	_ = exec.Command("systemctl", "disable", "unsinkdns.service").Run()

	unitPath := "/etc/systemd/system/unsinkdns.service"
	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unit file: %w", err)
	}

	if out, err := exec.Command("systemctl", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl daemon-reload failed: %v: %s", err, string(out))
	}

	// Remove the copied executable
	dest := "/usr/local/bin/unsinkdns"
	if removeErr := os.Remove(dest); removeErr != nil && !os.IsNotExist(removeErr) {
		return fmt.Errorf("failed to remove binary: %w", removeErr)
	}

	return nil
}
