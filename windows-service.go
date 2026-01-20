//go:build windows
// +build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kardianos/service"
)

func installService(s service.Service) error {
	// Copy current executable to C:\Program Files\UnsinkDNS\unsinkdns.exe
	destDir := `C:\Program Files\UnsinkDNS`
	dest := filepath.Join(destDir, "unsinkdns.exe")
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exePath, err := filepath.Abs(exe)
	if err != nil {
		exePath = exe
	}

	if exePath != dest {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}
		data, err := os.ReadFile(exePath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0755); err != nil {
			return err
		}
	}

	err = service.Control(s, "install")
	if err != nil {
		return err
	}

	// Add firewall rule to allow UDP port 53
	_ = exec.Command("netsh", "advfirewall", "firewall", "add", "rule", "name=UnsinkDNS", "dir=in", "action=allow", "protocol=UDP", "localport=53").Run()

	return nil
}

func removeService(s service.Service) error {
	err := service.Control(s, "uninstall")
	if err != nil {
		return err
	}

	// Remove the copied executable
	dest := `C:\Program Files\UnsinkDNS\unsinkdns.exe`
	if removeErr := os.Remove(dest); removeErr != nil && !os.IsNotExist(removeErr) {
		println("Error deleting " + dest)
	}

	// Remove firewall rule
	_ = exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name=UnsinkDNS").Run()

	return nil
}
