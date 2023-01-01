//go:build windows

package commandline

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func PowershellProfilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	documentDir := windowsDocumentDir(homeDir)
	return filepath.Join(documentDir, "WindowsPowerShell", "profile.ps1"), nil
}

func BashrcPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".bashrc"), nil
}

func windowsDocumentDir(homeDir string) string {
	defaultDocumentDir := filepath.Join(homeDir, "Documents")

	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\Shell Folders`, registry.QUERY_VALUE)
	if err != nil {
		return defaultDocumentDir
	}
	defer k.Close()
	value, _, err := k.GetStringValue("Personal")
	if value == "" || err != nil {
		return defaultDocumentDir
	}
	return value
}
