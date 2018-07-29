// +build !windows

package tui

import "github.com/junegunn/fzf/src/util"

func (r *LightRenderer) initPlatform() error {
	return nil
}

func (r *LightRenderer) closePlatform() {
	// NOOP
}

func (r *LightRenderer) getch(nonblock bool) (int, bool) {
	b := make([]byte, 1)
	fd := r.fd()
	util.SetNonblock(r.ttyin, nonblock)

	_, err := util.Read(fd, b)

	if err != nil {
		return 0, false
	}
	return int(b[0]), true
}
