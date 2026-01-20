//go:build windows
// +build windows

package main

import (
	"os"
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

	return service.Control(s, "install")
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

	return nil
}
