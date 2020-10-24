package main

import (
	"dev_tools/files"
	"fmt"
	"github.com/google/uuid"
	"os"
	"os/exec"
)

const TEMP_DIR_NAME = "temp_repos"

var TEMP_REPO_LOCATION = files.JoinSegmentsOfFilePath(files.GetHomeDirLocation(), TEMP_DIR_NAME)

func main() {
	// create temp repo if not exists
	if err := files.CreateDirIfNotExists(TEMP_REPO_LOCATION); err != nil {
		println("couldn't create temp repo dir")
		panic(err)
	}

	validateArgs(os.Args)
	handleTempGitClone(os.Args[1:])
}

func handleTempGitClone(args []string) {
	switch len(args) {
	case 1:
		handleCommand(args[0], "", "code")
		break
	case 2:
		handleCommand(args[0], args[1], "code")
		break
	case 3:
		handleCommand(args[0], args[1], args[2])
		break
	}
}

func handleCommand(repo, branch, openOption string) {
	tempDirLocation := CreateTempDir()
	if len(branch) == 0 {
		if err := simpleCloneToDir(repo, tempDirLocation); err != nil {
			panic(err)
		}
	} else {
		if err := branchSpecificCloneToDir(repo, branch, tempDirLocation); err != nil {
			panic(err)
		}
	}
	println(fmt.Sprintf("clone repo %s into dir [%s]", repo, tempDirLocation))

	// open option
	if openOption == "none" {
		return
	}

	// handle open
	if err := handleOpenCommand(tempDirLocation, openOption); err != nil {
		panic(err)
	}
}

func handleOpenCommand(tempDirLocation, openOption string) error {
	openCommand := exec.Command(openOption, tempDirLocation)
	openCommand.Stdout = os.Stdout
	openCommand.Stderr = os.Stderr
	return openCommand.Run()
}

func simpleCloneToDir(repo, location string) error {
	simpleCloneCmd := exec.Command("git", "-C", location, "clone", repo)
	simpleCloneCmd.Stdout = os.Stdout
	simpleCloneCmd.Stderr = os.Stderr
	return simpleCloneCmd.Run()
}

func branchSpecificCloneToDir(repo, branch, location string) error {
	branchCloneCmd := exec.Command("git", "clone", repo, "--branch", branch, "--single-branch", location)
	branchCloneCmd.Stdout = os.Stdout
	branchCloneCmd.Stderr = os.Stderr
	return branchCloneCmd.Run()
}

func validateArgs(args []string) {
	if len(args) < 2 {
		panic("must pass in repo location")
	}
}

func CreateTempDir() string {
	tempDirLocation := files.JoinSegmentsOfFilePath(TEMP_REPO_LOCATION, uuid.New().String())
	if err := files.CreateDirIfNotExists(tempDirLocation); err != nil {
		panic(err)
	}
	return tempDirLocation
}
