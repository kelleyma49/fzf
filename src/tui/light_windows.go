//+build windows

package tui

import (
	"github.com/junegunn/fzf/src/util"
	"golang.org/x/sys/windows"
)

var (
	oldStateInput  uint32
	oldStateOutput uint32
)

func (r *LightRenderer) initPlatform() error {

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

	for _, con := range [2]windows.Handle{windows.Stderr, windows.Stdout} {
		// enable vt100 emulation (https://docs.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences)
		if err := windows.GetConsoleMode(con, &oldStateOutput); err != nil {
			return err
		}
		//var requestedOutModes uint32 = windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING | windows.DISABLE_NEWLINE_AUTO_RETURN
		var requestedOutModes uint32 = (oldStateOutput | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING | windows.DISABLE_NEWLINE_AUTO_RETURN) | windows.ENABLE_LINE_INPUT
		if err := windows.SetConsoleMode(con, requestedOutModes); err != nil {
			return err
		}
	}

	if err := windows.GetConsoleMode(windows.Stdin, &oldStateInput); err != nil {
		return err
	}
	var requestedInModes uint32 = oldStateInput | windows.ENABLE_VIRTUAL_TERMINAL_INPUT
	if err := windows.SetConsoleMode(windows.Stdin, requestedInModes); err != nil {
		return err
	}

	return nil
}

func (r *LightRenderer) closePlatform() {
	windows.SetConsoleMode(windows.Stderr, oldStateOutput)
	windows.SetConsoleMode(windows.Stdin, oldStateInput)
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
