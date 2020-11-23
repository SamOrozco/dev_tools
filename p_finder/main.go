package main

import (
	"dev_tools/files"
	"fmt"
	"github.com/labstack/gommon/color"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type AppOptions struct {
	DirsOnly       bool
	VerboseLogging bool
}

type RegexMatch struct {
	FilePath              string
	MatchValue            string
	MatchValueWithPadding string
	IsDir                 bool
}

var (
	DirsOnly       bool // search dirs only flag name
	VerboseLogging bool // enable verbose logging
	rootCmd        = &cobra.Command{
		Use:   "hugo",
		Short: "Hugo is a very fast static site generator",
		Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at http://hugo.spf13.com`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				panic("must pass pattern and directory e.g. `pf ^.*test$ test_dir/`")
			}
			pattern := args[0]
			findDir := args[1]
			pfDir(pattern, findDir, &AppOptions{
				DirsOnly:       DirsOnly,
				VerboseLogging: VerboseLogging,
			})
		},
	}
)

func main() {
	rootCmd.PersistentFlags().BoolVarP(&DirsOnly, "dirs", "d", false, "find in dir names only")
	rootCmd.PersistentFlags().BoolVarP(&VerboseLogging, "verbose", "v", false, "enable verbose logging")
	Execute()
}

/**
PF the given dir
*/
func pfDir(pattern, dir string, options *AppOptions) {
	println(fmt.Sprintf(color.Green("searching for \"%s\""), pattern))

	regexPattern, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	fileSearchWaitGroup := &sync.WaitGroup{}
	fileChan := make(chan *RegexMatch, 0)

	// start listening in file chan
	printValues := make([]string, 0)
	go handleFoundValues(fileChan, fileSearchWaitGroup, &printValues)

	if files.IsDir(dir) {
		fileSearchWaitGroup.Add(1)
		go searchDirAsync(
			dir,
			regexPattern,
			fileChan,
			fileSearchWaitGroup,
			options,
		)
	} else if !options.DirsOnly {
		fileSearchWaitGroup.Add(1)
		go searchFileAsync(
			dir,
			regexPattern,
			fileChan,
			fileSearchWaitGroup,
			options,
		)
	}

	// wait for all to finish
	fileSearchWaitGroup.Wait()

	if len(printValues) < 1 {
		println(color.Green("no matches!"))
	} else {
		for i := range printValues {
			println(printValues[i])
		}
	}
}

/**
Recursively search the directory for the regex pattern in files
*/
func searchDirAsync(
	dirPath string,
	regex *regexp.Regexp,
	fileChan chan *RegexMatch,
	wg *sync.WaitGroup,
	options *AppOptions,
) {

	if options.VerboseLogging {
		LogSearchingDir(dirPath)
	}

	dirFiles, err := ioutil.ReadDir(dirPath)
	if err != nil {
		wg.Done()
		return
	}
	for i := range dirFiles {
		currentFile := dirFiles[i]
		currentFilePath := filepath.Join(dirPath, currentFile.Name())
		if files.IsDir(currentFilePath) {
			// send dir name if it matches
			sendDirValueIfApplicable(currentFilePath, regex, wg, fileChan)
			// recurse
			wg.Add(1)
			go searchDirAsync(currentFilePath, regex, fileChan, wg, options)
		} else if !options.DirsOnly {
			wg.Add(1)
			go searchFileAsync(currentFilePath, regex, fileChan, wg, options)
		}
	}
	wg.Done()
}

/**
searches the given file as for the given pattern
if we find pattern value in file data we will pass file location to fileChan
*/
func searchFileAsync(
	filePath string,
	pattern *regexp.Regexp,
	fileChan chan *RegexMatch,
	wg *sync.WaitGroup,
	options *AppOptions,
) {
	if options.VerboseLogging {
		LogSearchingFile(filePath)
	}

	dataBytes, err := files.ReadBytesFromFile(filePath)
	if err != nil {
		wg.Done()
		return
	}
	if pattern.Match(dataBytes) {
		matchVal, matchValueWithPadding := getPrintValue(dataBytes, pattern)
		fileChan <- &RegexMatch{
			FilePath:              filePath,
			MatchValue:            matchVal,
			MatchValueWithPadding: matchValueWithPadding,
			IsDir:                 false,
		}
	} else {

		// only on error or not containing value will we close wait group here
		// otherwise printing go routine will close wait group after printing.
		wg.Done()
	}
}

func sendDirValueIfApplicable(
	currentPath string,
	reg *regexp.Regexp,
	wg *sync.WaitGroup,
	fileChan chan *RegexMatch) {
	if reg.MatchString(currentPath) {
		wg.Add(1) // need add because we remove in handling values
		fileChan <- &RegexMatch{
			FilePath: currentPath,
			IsDir:    true,
		}
	}
}

/**
Gets print values for the current data
*/
func getPrintValue(dataBytes []byte, pattern *regexp.Regexp) (string, string) {
	stringVal := string(dataBytes)
	indexes := pattern.FindStringIndex(stringVal)
	startIdx := indexes[0]
	endIdx := indexes[1]
	frontPadding := startIdx
	endPadding := len(stringVal) - endIdx

	valueString := stringVal[startIdx:endIdx]
	paddingValues := []int{1000, 500, 250}
	for i := range paddingValues {
		padding := paddingValues[i]
		if frontPadding > padding && endPadding > padding {
			return valueString, stringVal[startIdx-padding : endIdx+padding]
		}
	}
	return valueString, valueString
}

func handleFoundValues(
	fileChan chan *RegexMatch,
	fileSearchWaitGroup *sync.WaitGroup,
	printValues *[]string,
) {
	for file := range fileChan {
		path := file.FilePath
		matchValue := file.MatchValue
		matchValueWithPadding := file.MatchValueWithPadding

		// directory is a special case
		if file.IsDir {
			*printValues = append(*printValues, color.Red(path))
		} else { // if file
			*printValues = append(*printValues, color.Blue(path))
			*printValues = append(*printValues, strings.ReplaceAll(matchValueWithPadding, matchValue, color.MagentaBg(matchValue)))
		}

		// this will dec wait group so we can be sure to print
		// file not containing value will close wait group
		fileSearchWaitGroup.Done()
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func LogSearchingFile(name string) {
	println(color.BlueBg(fmt.Sprintf("searching [%s]", name)))
}

func LogSearchingDir(name string) {
	println(color.RedBg(fmt.Sprintf("searching [%s]", name)))
}
