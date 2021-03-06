package consoleicon

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var shell32 = windows.NewLazySystemDLL("shell32.dll")
var procExtractIconExW = shell32.NewProc("ExtractIconExW")

var kernel32 = windows.NewLazySystemDLL("kernel32.dll")
var procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")

var user32 = windows.NewLazySystemDLL("user32.dll")
var procLoadIcon = user32.NewProc("LoadIconW")
var procDestroyIcon = user32.NewProc("DestroyIcon")
var procSendMessage = user32.NewProc("SendMessageA")

type _Icon uintptr

func (h _Icon) Close() {
	if h != 0 {
		procDestroyIcon.Call(uintptr(h))
	}
}

func readFrom(fname string) (_Icon, error) {
	_fname, err := windows.UTF16PtrFromString(fname)
	if err != nil {
		return _Icon(0), err
	}
	var handle uintptr

	rc, _, err := procExtractIconExW.Call(
		uintptr(unsafe.Pointer(_fname)),
		0,
		uintptr(unsafe.Pointer(&handle)),
		0, //small
		1)
	if rc <= 0 {
		return 0, err
	}
	return _Icon(handle), nil
}

func sendMessage(h, m, w, l uintptr) uintptr {
	rc, _, _ := procSendMessage.Call(h, m, w, l)
	return rc
}

func getConsoleWindow() uintptr {
	handle, _, _ := procGetConsoleWindow.Call()
	return handle
}

func setFrom(fname string) (func(bool), error) {
	h, err := readFrom(fname)
	if err != nil {
		return func(bool) {}, err
	}

	hConsole := getConsoleWindow()

	org_big := sendMessage(hConsole, WM_GETICON, ICON_BIG, uintptr(h))
	org_small := sendMessage(hConsole, WM_GETICON, ICON_SMALL, uintptr(h))

	sendMessage(hConsole, WM_SETICON, ICON_BIG, uintptr(h))
	sendMessage(hConsole, WM_SETICON, ICON_SMALL, uintptr(h))
	return func(restore bool) {
		if restore {
			sendMessage(hConsole, WM_SETICON, ICON_BIG, org_big)
			sendMessage(hConsole, WM_SETICON, ICON_SMALL, org_small)
		}
		h.Close()
	}, nil
}

func setFromExe() (func(bool), error) {
	fname, err := os.Executable()
	if err != nil {
		return func(bool) {}, err
	}
	return SetFrom(fname)
}
