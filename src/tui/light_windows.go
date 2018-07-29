//+build windows

package tui

import (
	"fmt"
	"os"
	"syscall"

	"github.com/junegunn/fzf/src/util"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/windows"
)

var (
	oldStateInput  uint32
	oldStateOutput uint32
)

func (r *LightRenderer) initPlatform() error {
	for _, con := range [2]windows.Handle{windows.Stderr, windows.Stdout} {
		// enable vt100 emulation (https://docs.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences)
		if err := windows.GetConsoleMode(con, &oldStateOutput); err != nil {
			return err
		}
		//var requestedOutModes uint32 = windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING | windows.DISABLE_NEWLINE_AUTO_RETURN
		//var requestedOutModes uint32 = oldStateOutput | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING | windows.ENABLE_LINE_INPUT | windows.ENABLE_PROCESSED_OUTPUT
		var requestedOutModes uint32 = windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING | windows.ENABLE_PROCESSED_OUTPUT /*| windows.DISABLE_NEWLINE_AUTO_RETURN*/
		if err := windows.SetConsoleMode(con, requestedOutModes); err != nil {
			return err
		}
	}

	inHandle := windows.Stdin
	//inHandle, _ := syscall.Open("CONIN$", syscall.O_RDWR, 0)
	if err := windows.GetConsoleMode(windows.Handle(inHandle), &oldStateInput); err != nil {
		return err
	}
	var requestedInModes uint32 = windows.ENABLE_VIRTUAL_TERMINAL_INPUT | windows.ENABLE_PROCESSED_INPUT | windows.ENABLE_WINDOW_INPUT | windows.ENABLE_MOUSE_INPUT | windows.ENABLE_EXTENDED_FLAGS
	if err := windows.SetConsoleMode(windows.Handle(inHandle), requestedInModes); err != nil {
		return err
	}

	// channel for non-blocking reads:
	r.ttyinChannel = make(chan byte)

	// the following allows for non-blocking IO.
	// syscall.SetNonblock() is a NOOP under Windows.
	go func() {
		for {
			fd := r.fd()
			b := make([]byte, 1)
			_, err := util.Read(fd, b)
			if err == nil {
				//	break
				r.ttyinChannel <- b[0]
			}
		}
	}()

	return nil
}

func (r *LightRenderer) closePlatform() {
	windows.SetConsoleMode(windows.Stderr, oldStateOutput)
	windows.SetConsoleMode(windows.Stdin, oldStateInput)
}

func openTtyIn() *os.File {
	in, err := os.OpenFile("CONIN$", syscall.O_RDWR, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open "+consoleDevice)
		os.Exit(2)
	}
	return in
}

func (r *LightRenderer) setupTerminal() {
	terminal.MakeRaw(r.fd())
}

func (r *LightRenderer) restoreTerminal() {
	terminal.Restore(r.fd(), r.origState)
}

func (r *LightRenderer) updateTerminalSize() {
	var bufferInfo windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Stdout, &bufferInfo); err != nil {
		r.width = getEnv("COLUMNS", defaultWidth)
		r.height = r.maxHeightFunc(getEnv("LINES", defaultHeight))

	} else {
		r.width = int(bufferInfo.Window.Right - bufferInfo.Window.Left)
		r.height = r.maxHeightFunc(int(bufferInfo.Window.Bottom - bufferInfo.Window.Top))
		fmt.Printf("w %d h %d", r.width, r.height)
	}
}

func (r *LightRenderer) findOffset() (row int, col int) {
	var bufferInfo windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Stdout, &bufferInfo); err != nil {
		return -1, -1
	}
	return int(bufferInfo.CursorPosition.X), int(bufferInfo.CursorPosition.Y)
}

func (r *LightRenderer) getch(nonblock bool) (int, bool) {
	if nonblock {
		select {
		case bc := <-r.ttyinChannel:
			return int(bc), true
		default:
			return 0, false
		}
	} else {
		bc := <-r.ttyinChannel
		return int(bc), true
	}
}
