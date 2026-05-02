package main

import (
	"os"

	"github.com/jcrussell/byob-go-cli/internal/byobcmd"
	"github.com/jcrussell/byob-go-cli/pkg/cmd/root"
	"github.com/jcrussell/byob-go-cli/pkg/cmdutil"
)

func main() {
	f := cmdutil.New()
	os.Exit(byobcmd.Run(root.NewCmdRoot(f), os.Args[1:], f.IOStreams))
}
