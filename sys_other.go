//go:build !windows

package main

import (
	"os"
	"path/filepath"
)

// Заглушки для сборки/тестов на Linux и macOS.
func initConsole()                 {}
func writeOut(s string)            { os.Stdout.WriteString(s) }
func isHiddenOrSystem(string) bool { return false }
func hide(string)                  {}
func unhide(string)                {}
func refreshDesktop()              {}
func desktopPath() string {
	if d := os.Getenv("DESKTOP_SORTER_DIR"); d != "" {
		return d
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, "Desktop")
}
