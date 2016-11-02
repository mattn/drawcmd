package drawcmd

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"syscall"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	user32                         = syscall.NewLazyDLL("user32.dll")
	gdi32                          = syscall.NewLazyDLL("gdi32.dll")
	procGetConsoleWindow           = kernel32.NewProc("GetConsoleWindow")
	procGetDC                      = user32.NewProc("GetDC")
	procReleaseDC                  = user32.NewProc("ReleaseDC")
	procGetCurrentConsoleFont      = kernel32.NewProc("GetCurrentConsoleFont")
	procGetConsoleFontSize         = kernel32.NewProc("GetConsoleFontSize")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetPixel                   = gdi32.NewProc("SetPixel")
)

type (
	short int16
	dword uint32

	coord struct {
		X short
		Y short
	}

	smallRect struct {
		Left   short
		Top    short
		Right  short
		Bottom short
	}

	consoleFontInfo struct {
		Font     dword
		FontSize coord
	}

	consoleScreenBufferInfo struct {
		Size              coord
		CursorPosition    coord
		Attributes        uint16
		Window            smallRect
		MaximumWindowSize coord
	}
)

func syserr(err error) bool {
	if errno, ok := err.(syscall.Errno); ok {
		return errno != 0
	}
	return err != nil
}

func Render(img image.Image) error {
	hwnd, _, err := procGetConsoleWindow.Call()
	if syserr(err) {
		return err
	}
	hdc, _, err := procGetDC.Call(hwnd)
	if syserr(err) {
		return err
	}
	defer procReleaseDC.Call(hdc)

	stdout, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if syserr(err) {
		return err
	}

	var font consoleFontInfo
	var t int32 = 1
	_, _, err = procGetCurrentConsoleFont.Call(uintptr(stdout), uintptr(t), uintptr(unsafe.Pointer(&font)))
	if syserr(err) {
		return err
	}

	fontSize, _, err := procGetConsoleFontSize.Call(uintptr(stdout), uintptr(font.Font))
	if syserr(err) {
		return err
	}
	fs := *(*coord)(unsafe.Pointer(&fontSize))

	bounds := img.Bounds()
	dx, dy := bounds.Dx(), bounds.Dy()

	skip := int(dy / int(fs.Y))
	for i := 0; i < skip; i++ {
		fmt.Println()
	}

	var csbi consoleScreenBufferInfo
	_, _, err = procGetConsoleScreenBufferInfo.Call(uintptr(stdout), uintptr(unsafe.Pointer(&csbi)))
	if syserr(err) {
		return err
	}

	starty := int(csbi.CursorPosition.Y-csbi.Window.Top) * int(fs.Y)
	if int(csbi.CursorPosition.Y)+skip > int(csbi.Window.Top) {
		starty -= skip * int(fs.Y)
	}

	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a < 128 {
				continue
			}
			c := uint32(uint8(b))<<16 | uint32(uint8(g))<<8 | uint32(uint8(r))
			procSetPixel.Call(hdc, uintptr(x), uintptr(y+starty), uintptr(c))
		}
	}
	return nil
}
