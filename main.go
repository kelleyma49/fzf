package main

import (
	"github.com/junegunn/fzf/src"
	"github.com/junegunn/fzf/src/util"
)

var revision string

func main() {
	fzf.Run(fzf.ParseOptions(), revision)
	util.RunExitHandlers()
}
