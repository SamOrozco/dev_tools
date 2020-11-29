package main

import (
	"dev_tools/files"
	"fmt"
	"github.com/labstack/gommon/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Options struct {
	VerboseLogging bool
	ComponentsCSV  string
}

var (
	DefaultConfigFileName string = "bldr.yaml"
	VerboseLogging        bool   // search dirs only flag name
	ComponentsCSV         string // only build component with given names, csv on names
	rootCmd               = &cobra.Command{
		Use:   "bldr",
		Short: "project builder",
		Long:  `build your projects`,
		Run: func(cmd *cobra.Command, args []string) {
			var fileName string
			if len(args) < 1 {
				// check if bldr.yaml is in current file location
				if !files.FileExists("./bldr.yaml") {
					panic("must supply file location or have one in current dir")
				} else {
					fileName = "./bldr.yaml"
				}
			} else {
				fileName = args[0]
			}

			buildConfigFromFile(fileName, &Options{
				VerboseLogging: VerboseLogging,
				ComponentsCSV:  ComponentsCSV,
			},
				".")
		},
	}
)

func main() {
	rootCmd.PersistentFlags().BoolVarP(&VerboseLogging, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&ComponentsCSV, "ComponentsCSV", "c", "", "csv of the component to build if empty build all")
	Execute()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func Bldr(config *Config, options *Options, dir string) {
	rootDir := dir
	if len(config.RootDir) > 0 {
		rootDir = config.RootDir
	}

	if len(config.Components) < 1 {
		panic("no components configured to build")
	}

	// IMPORTANT where build will write artifacts
	resolvedArtifactPath := "."
	var err error
	if len(config.ArtifactDir) > 0 {
		if filepath.IsAbs(config.ArtifactDir) {
			resolvedArtifactPath = config.ArtifactDir
		} else {
			if resolvedArtifactPath, err = filepath.Abs(files.JoinSegmentsOfFilePath(resolvedArtifactPath, config.ArtifactDir)); err != nil {
				panic(err)
			}
		}
	}

	if err := files.CreateDirIfNotExists(resolvedArtifactPath); err != nil {
		panic(err)
	}
	componentNamePredicate := getComponentNamePredicate(options.ComponentsCSV)

	for i := range config.Components {
		curComp := config.Components[i]
		if componentNamePredicate(curComp) {
			println(fmt.Sprintf("Start [%d]%s", i, strings.Repeat(".", 20)))
			buildComponent(
				rootDir,
				resolvedArtifactPath,
				curComp,
				options,
			)
			println(fmt.Sprintf("Done [%d]%s", i, strings.Repeat(".", 20)))
		}
	}
}

/**

 */
func buildComponent(
	location,
	artifactLocation string,
	comp *Component,
	options *Options,
) []string {
	if comp == nil {
		return []string{}
	}

	componentLocation := files.JoinSegmentsOfFilePath(location, comp.Location)
	if !files.FileExists(componentLocation) {
		invalidComponentLocation(comp, componentLocation)
	}

	if comp.Build != nil {
		if options.VerboseLogging {
			LogStartBuilding(comp.Name)
		}
		execCommands(componentLocation, comp.Build.Commands, options)

		if comp.Build.Artifacts != nil {
			// copy artifacts to proper location
			copyArtifacts(comp.Build.Artifacts, componentLocation, artifactLocation, options)
		}
	} else {
		// if user did configure build command for component we expect there to be a build config in directory
		buildConfigFromFile(
			files.JoinSegmentsOfFilePath(componentLocation, DefaultConfigFileName),
			options,
			componentLocation)
	}

	return []string{}
}

func execCommands(
	location string,
	commands []*Command,
	options *Options,
) {
	if len(commands) < 1 {
		return
	}

	for i := range commands {
		currentCommand := commands[i]
		runCommand(location, currentCommand, options)
	}
}

func runCommand(
	location string,
	command *Command,
	options *Options,
) {
	if len(command.Windows) > 0 && isWindows() {
		execCommandString(location, command.Windows, options)
	} else if len(command.Linux) > 0 && isLinux() {
		execCommandString(location, command.Linux, options)
	} else if len(command.Mac) > 0 && isMac() {
		execCommandString(location, command.Mac, options)
	} else {
		execCommandString(location, command.Exec, options)
	}
}

func execCommandString(
	location string,
	cmd string,
	options *Options,
) {
	if len(cmd) < 1 {
		return
	}

	if options.VerboseLogging {
		LogExecutingCommand(location, cmd)
	}

	segments := strings.Fields(cmd)
	execCommand(location, segments...)
}

func execCommand(location string, commands ...string) {
	if len(commands) == 1 {
		cmd := exec.Command(commands[0])
		cmd.Dir = location
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	} else {
		cmd := exec.Command(commands[0], commands[1:]...)
		cmd.Dir = location
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}
}

func copyArtifacts(artifacts []*Artifact, curLocation, artifactLocation string, options *Options) {
	if artifacts != nil && len(artifacts) > 0 {
		for i := range artifacts {
			curArt := artifacts[i]
			copyArtifact(curArt, curLocation, artifactLocation, options)
		}
	}
}

func copyArtifact(artifact *Artifact, curLocation, artLocation string, options *Options) {
	if artifact != nil {

		name := artifact.Name
		artifactPath := files.JoinSegmentsOfFilePath(curLocation, name)
		if files.FileExists(artifactPath) {
			finalName := name
			// set final artifact name if alias
			if len(artifact.Alias) > 0 {
				finalName = artifact.Alias
			}

			// move to artifact dir
			dest := files.JoinSegmentsOfFilePath(artLocation, finalName)

			// if artifact is dir we expect to be copying dir
			// ensure destination is a fully created dir
			if !files.FileExists(dest) {
				// todo break out
				destDir, _ := filepath.Split(dest)
				if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
					panic(err)
				}
			} else {
				// removing dir and all contents to make room for new artifact
				if err := os.RemoveAll(dest); err != nil {
					panic(err)
				}
			}

			if err := os.Rename(artifactPath, dest); err != nil {
				panic(err)
			} else {
				if options.VerboseLogging {
					println(fmt.Sprintf("copying %s to %s", artifactPath, dest))
				}
			}
		}
	}
}

/**
 */
func buildConfigFromFile(
	loc string,
	options *Options,
	dir string,
) {
	data, err := files.ReadBytesFromFile(loc)
	if err != nil {
		panic(err)
	}
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		panic(err)
	}
	Bldr(config, options, dir)
}

/**

 */
func getComponentNamePredicate(comp string) func(component *Component) bool {
	if len(comp) < 1 {
		return func(component *Component) bool {
			return true
		}
	} else {
		names := strings.Split(comp, ",")
		varMap := make(map[string]bool, 0)
		for i := range names {
			varMap[names[i]] = true
		}

		return func(component *Component) bool {
			if val, exists := varMap[component.Name]; exists {
				return val
			}
			return false
		}
	}
}

/**
 */
func invalidComponentLocation(comp *Component, loc string) {
	if len(comp.Name) > 0 {
		panic(fmt.Sprintf("invalid component location(%s) for %s", loc, comp.Name))
	}
	panic(fmt.Sprintf("invalid location(%s) for unnamed component", loc))
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func isLinux() bool {
	return runtime.GOOS == "linux"
}

func isMac() bool {
	return runtime.GOOS == "darwin"
}

/**

 */
func LogStartBuilding(compName string) {
	println(color.Green(fmt.Sprintf("building component %s", compName)))
}

/**

 */
func LogExecutingCommand(location, cmd string) {
	println(color.Red(fmt.Sprintf("executing command[%s] in location[%s]", cmd, location)))
}
