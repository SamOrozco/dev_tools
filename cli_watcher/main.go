package main

import (
	"fmt"
	"github.com/gen2brain/beeep"
	"os"
	"os/exec"
	"strings"
)

func main() {

	osArgsLen := len(os.Args)
	if osArgsLen < 2 {
		println("no arguments passed")
		os.Exit(0)
	}

	newArgs := os.Args[1:]
	commandString := strings.Join(newArgs, " ")
	cmd := exec.Command(newArgs[0], newArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		err = beeep.Notify("Error", fmt.Sprintf("There was an error running (%s)", commandString), "assets/warning.png")
		if err != nil {
			panic(err)
		}
		return
	}

	// notify
	err = beeep.Notify("Finished", fmt.Sprintf("(%s) has finished running", commandString), "assets/information.png")
	if err != nil {
		panic(err)
	}
}
