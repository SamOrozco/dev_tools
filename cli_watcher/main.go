package main

import (
	"dev_tools/files"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/spf13/cobra"
	"io"
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

	// cmd write to standard out and a file
	logFile, fileName, err := GetAndCreateLogFile(commandString)
	defer logFile.Close()
	var stdOutWriter io.Writer
	var stdErrWriter io.Writer
	if err == nil {
		println(fmt.Sprintf("writing output to %s", fileName))
		stdOutWriter = io.MultiWriter(os.Stdout, logFile)
		stdErrWriter = io.MultiWriter(os.Stderr, logFile)
	} else {
		stdOutWriter = os.Stdout
		stdErrWriter = os.Stderr
	}

	cmd.Stdout = stdOutWriter
	cmd.Stderr = stdErrWriter
	cmd.Stdin = os.Stdin

	err = cmd.Run()
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

func GetAndCreateLogFile(cmdString string) (*os.File, string, error) {
	fileString := strings.ReplaceAll(cmdString, " ", "_") + ".log"
	if files.FileExists(fileString) {
		if err := os.Remove(fileString); err != nil {
			panic(err)
		}
	}
	file, err := os.Create(fileString)
	return file, fileString, err
}
