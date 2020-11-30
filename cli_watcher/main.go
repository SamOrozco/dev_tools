package main

import (
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

var (
	rootCmd = &cobra.Command{
		Use:   "cliw",
		Short: "Cli Watcher",
		Long:  `Watch your cli commands`,
		Run: func(cmd *cobra.Command, args []string) {
			osArgsLen := len(os.Args)
			if osArgsLen < 2 {
				println("no arguments passed")
				os.Exit(0)
			}

			newArgs := os.Args[1:]
			RunWatcher(newArgs)
		},
	}
)

func main() {
	Execute()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func RunWatcher(args []string) {
	commandString := strings.Join(args, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		err = beeep.Alert("Error", fmt.Sprintf("There was an error running (%s)", commandString), "assets/warning.png")
		if err != nil {
			panic(err)
		}
		return
	}

	// notify
	err = beeep.Alert("Finished", fmt.Sprintf("(%s) has finished running", commandString), "assets/information.png")
	if err != nil {
		panic(err)
	}
}
