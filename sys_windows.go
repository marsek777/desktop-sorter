//go:build windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	shell32               = syscall.NewLazyDLL("shell32.dll")
	ole32                 = syscall.NewLazyDLL("ole32.dll")
	procSetConsoleOutCP   = kernel32.NewProc("SetConsoleOutputCP")
	procSetConsoleCP      = kernel32.NewProc("SetConsoleCP")
	procWriteConsoleW     = kernel32.NewProc("WriteConsoleW")
	procGetFileAttributes = kernel32.NewProc("GetFileAttributesW")
	procSetFileAttributes = kernel32.NewProc("SetFileAttributesW")
	procSetConsoleTitle   = kernel32.NewProc("SetConsoleTitleW")
	procSHGetKnownFolder  = shell32.NewProc("SHGetKnownFolderPath")
	procSHChangeNotify    = shell32.NewProc("SHChangeNotify")
	procCoTaskMemFree     = ole32.NewProc("CoTaskMemFree")
)

const (
	fileAttrHidden = 0x2
	fileAttrSystem = 0x4
	invalidAttrs   = 0xFFFFFFFF
)

// FOLDERID_Desktop {B4BFCC3A-DB2C-424C-B029-7FE99A87C641}
var folderIDDesktop = syscall.GUID{
	Data1: 0xB4BFCC3A, Data2: 0xDB2C, Data3: 0x424C,
	Data4: [8]byte{0xB0, 0x29, 0x7F, 0xE9, 0x9A, 0x87, 0xC6, 0x41},
}

func initConsole() {
	procSetConsoleOutCP.Call(65001)
	procSetConsoleCP.Call(65001)
	t, _ := syscall.UTF16PtrFromString("Desktop Sorter — сортировка рабочего стола")
	procSetConsoleTitle.Call(uintptr(unsafe.Pointer(t)))
}

// writeOut печатает через WriteConsoleW, чтобы кириллица корректно
// отображалась даже в старых консолях Windows 7/8.
func writeOut(s string) {
	h := syscall.Handle(os.Stdout.Fd())
	u := syscall.StringToUTF16(s)
	u = u[:len(u)-1]
	if len(u) == 0 {
		return
	}
	var written uint32
	r, _, _ := procWriteConsoleW.Call(uintptr(h), uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)), uintptr(unsafe.Pointer(&written)), 0)
	if r == 0 { // вывод перенаправлен в файл — пишем как UTF-8
		os.Stdout.WriteString(s)
	}
}

// desktopPath — настоящий путь рабочего стола (учитывает перенос в OneDrive
// и на другой диск). Только папка текущего пользователя.
func desktopPath() string {
	if procSHGetKnownFolder.Find() == nil {
		var p *uint16
		r, _, _ := procSHGetKnownFolder.Call(uintptr(unsafe.Pointer(&folderIDDesktop)), 0, 0, uintptr(unsafe.Pointer(&p)))
		if r == 0 && p != nil {
			s := syscall.UTF16ToString((*[1 << 15]uint16)(unsafe.Pointer(p))[:])
			procCoTaskMemFree.Call(uintptr(unsafe.Pointer(p)))
			if s != "" {
				return s
			}
		}
	}
	return filepath.Join(os.Getenv("USERPROFILE"), "Desktop")
}

func attrs(p string) uint32 {
	u, err := syscall.UTF16PtrFromString(p)
	if err != nil {
		return invalidAttrs
	}
	r, _, _ := procGetFileAttributes.Call(uintptr(unsafe.Pointer(u)))
	return uint32(r)
}

func isHiddenOrSystem(p string) bool {
	a := attrs(p)
	return a != invalidAttrs && a&(fileAttrHidden|fileAttrSystem) != 0
}

func setAttr(p string, a uint32) {
	u, err := syscall.UTF16PtrFromString(p)
	if err == nil {
		procSetFileAttributes.Call(uintptr(unsafe.Pointer(u)), uintptr(a))
	}
}

func hide(p string) {
	if a := attrs(p); a != invalidAttrs {
		setAttr(p, a|fileAttrHidden)
	}
}

func unhide(p string) {
	if a := attrs(p); a != invalidAttrs {
		setAttr(p, a&^fileAttrHidden)
	}
}

// refreshDesktop просит Проводник перерисовать рабочий стол.
func refreshDesktop() {
	const shcneAssocChanged = 0x08000000
	procSHChangeNotify.Call(shcneAssocChanged, 0, 0, 0)
}
