package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kechako/piperun/cmdpipe"
)

var cmdSep = flag.String("s", "{}", "a string of command separator.")

func makeCmdPipes(args []string) []*cmdpipe.CmdPipe {
	cmdPipes := make([]*cmdpipe.CmdPipe, 0, 10)

	var cmd *cmdpipe.CmdPipe
	for _, arg := range args {
		if arg == *cmdSep {
			cmd = nil
			continue
		}

		if cmd == nil {
			cmd = cmdpipe.NewCmdPipe()
			cmdPipes = append(cmdPipes, cmd)
		}

		if cmd.Name == "" {
			cmd.Name = arg
		} else {
			cmd.Args = append(cmd.Args, arg)
		}
	}

	var prevCmd *cmdpipe.CmdPipe
	for _, cmd := range cmdPipes {
		if prevCmd != nil {
			r, w := io.Pipe()

			prevCmd.PipeWriter = w
			cmd.PipeReader = r
		}

		prevCmd = cmd
	}

	return cmdPipes
}

func _main() (int, error) {
	var err error

	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		return 1, errors.New("Invalid parameters.")
	}

	cmdPipes := makeCmdPipes(args)
	if len(cmdPipes) == 0 {
		return 1, errors.New("Invalid parameters.")
	}

	for _, cmd := range cmdPipes {
		err = cmd.Start()
		if err != nil {
			return 2, err
		}
	}

	for _, cmdPipe := range cmdPipes {
		cmdPipe.Wait()
	}

	for _, cmdPipe := range cmdPipes {
		if cmdPipe.Error != nil {
			return cmdPipe.ExitStatus, cmdPipe.Error
		}
	}

	return 0, nil
}

func main() {
	exitstatus, err := _main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] %v\n", err)
		os.Exit(exitstatus)
	}
}
