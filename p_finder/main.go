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
	"time"
)

type AppOptions struct {
	DirsOnly              bool
	FilesOnly             bool
	VerboseLogging        bool
	ExcludeFileExtensions map[string]bool
	IncludeFileExtensions map[string]bool
	MaxDepth              int
}

type RegexMatch struct {
	FilePath              string
	MatchValue            string
	MatchValueWithPadding string
	IsPathName            bool
}

var (
	DirsOnly              bool   // search dirs only flag name
	VerboseLogging        bool   // enable verbose logging
	FilesOnly             bool   // enable verbose logging
	ExcludedFileExtension string // csv of excluded file extensions
	IncludedFileExtension string // csv of only file extensions we want to search for
	MaxDepth              int
	rootCmd               = &cobra.Command{
		Use:   "pf",
		Short: "pattern finder",
		Long:  `A fast and configurable pattern finder in files contents.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				panic("must pass pattern and directory e.g. `pf ^.*test$ test_dir/`")
			}
			pattern := args[0]
			pattern = patternReplace(pattern)
			findDir := args[1]

			startTime := time.Now()
			pfDir(pattern, findDir, &AppOptions{
				DirsOnly:              DirsOnly,
				VerboseLogging:        VerboseLogging,
				ExcludeFileExtensions: convertCsvToFlagMap(ExcludedFileExtension),
				IncludeFileExtensions: convertCsvToFlagMap(IncludedFileExtension),
				MaxDepth:              MaxDepth,
			})
			LogTimeRan(startTime)
		},
	}
)

func main() {
	rootCmd.PersistentFlags().BoolVarP(&DirsOnly, "dirs", "d", false, "find in dir names only")
	rootCmd.PersistentFlags().BoolVarP(&VerboseLogging, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&FilesOnly, "files", "f", false, "find in files only")
	rootCmd.PersistentFlags().StringVarP(&ExcludedFileExtension, "excluded-file-extensions", "x", "", "comma separated list of excluded file extensions")
	rootCmd.PersistentFlags().StringVarP(&IncludedFileExtension, "included-file-extensions", "i", "", "comma separated list of included file extensions")
	rootCmd.PersistentFlags().IntVarP(&MaxDepth, "max-depth", "m", 10, "max directory depth we will search defaults 30")
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
			0,
		)
	} else if !options.DirsOnly {
		fileSearchWaitGroup.Add(1)
		go searchFileAsync(
			dir,
			regexPattern,
			fileChan,
			fileSearchWaitGroup,
			options,
			0,
		)
	}

	// wait for all to finish
	fileSearchWaitGroup.Wait()

	if len(printValues) < 1 {
		println(color.Green("no matches!"))
	} else {
		println(strings.Join(printValues, "\n"))
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
	depth int,
) {

	// max depth check
	if depth >= (options.MaxDepth - 1) {
		if options.VerboseLogging {
			LogHitMaxDepth(dirPath)
		}
		wg.Done()
		return
	}

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
			if !options.FilesOnly {
				sendDirValueIfApplicable(currentFilePath, regex, wg, fileChan)
			}
			// recurse
			wg.Add(1)
			go searchDirAsync(currentFilePath, regex, fileChan, wg, options, depth+1)
		} else if !options.DirsOnly {
			sendDirValueIfApplicable(currentFilePath, regex, wg, fileChan)
			wg.Add(1)
			go searchFileAsync(currentFilePath, regex, fileChan, wg, options, depth+1)
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
	depth int,
) {

	// max depth check
	if depth >= options.MaxDepth {
		LogHitMaxDepth(filePath)
		wg.Done()
		return
	}

	if options.VerboseLogging {
		LogSearchingFile(filePath)
	}

	// if we have file extensions check this file
	if fileIsExcludedBecauseOfExtensionParam(filePath, options.IncludeFileExtensions, options.ExcludeFileExtensions) {
		wg.Done() // this files extension has been excluded
		return
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
			IsPathName:            false,
		}
	} else {

		// only on error or not containing value will we close wait group here
		// otherwise printing go routine will close wait group after printing.
		wg.Done()
	}
}

func fileIsExcludedBecauseOfExtensionParam(filePath string, includedExtensions, excludedExtensions map[string]bool) bool {
	if (includedExtensions == nil || len(includedExtensions) < 1) &&
		(excludedExtensions == nil || len(excludedExtensions) < 1) {
		return false
	}

	// included overrides excluded
	if includedExtensions != nil && len(includedExtensions) > 0 {
		// handle included
		toLowerExtension := strings.ToLower(filepath.Ext(filePath))
		if _, exists := includedExtensions[toLowerExtension]; exists {
			// included
			return false
		} else {
			return true
		}
	} else {
		// handle excluded
		toLowerExtension := strings.ToLower(filepath.Ext(filePath))
		if _, exists := excludedExtensions[toLowerExtension]; exists {
			// included
			return true
		} else {
			return false
		}
	}
}

/**
send dir name to file chan if applicable
*/
func sendDirValueIfApplicable(
	currentPath string,
	reg *regexp.Regexp,
	wg *sync.WaitGroup,
	fileChan chan *RegexMatch) {
	if reg.MatchString(currentPath) {
		wg.Add(1) // need add because we remove in handling values
		fileChan <- &RegexMatch{
			FilePath:   currentPath,
			IsPathName: true,
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
		if file.IsPathName {
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

/**
exec command
*/
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

/**
log searching file
*/
func LogSearchingFile(name string) {
	println(color.BlueBg(fmt.Sprintf("searching [%s]", name)))
}

/**
log search dir
*/
func LogSearchingDir(name string) {
	println(color.RedBg(fmt.Sprintf("searching [%s]", name)))
}

func LogHitMaxDepth(filePath string) {
	println(color.RedBg(fmt.Sprintf("hit max depth at %s", filePath)))
}

/**
converts csv to key boolean map for searching
*/
func convertCsvToFlagMap(csvString string) map[string]bool {
	result := make(map[string]bool)
	if len(csvString) < 1 {
		return result
	}

	segments := strings.Split(csvString, ",")
	for i := range segments {
		result[strings.ToLower(strings.TrimSpace(segments[i]))] = true
	}
	return result
}

/**
If pattern matches presets replace
*/
func patternReplace(pattern string) string {
	if strings.ToLower(strings.TrimSpace(pattern)) == "email" {
		// email regex
		return `([-!#-'*+/-9=?A-Z^-~]+(\.[-!#-'*+/-9=?A-Z^-~]+)*|"([]!#-[^-~ \t]|(\\[\t -~]))+")@[0-9A-Za-z]([0-9A-Za-z-]{0,61}[0-9A-Za-z])?(\.[0-9A-Za-z]([0-9A-Za-z-]{0,61}[0-9A-Za-z])?)+`
	}

	if strings.ToLower(strings.TrimSpace(pattern)) == "phone" {
		// phone regex
		return `^(\+\d{1,2}\s?)?1?\-?\.?\s?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}$`
	}

	if strings.ToLower(strings.TrimSpace(pattern)) == "url" {
		// url regex
		return `https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`
	}

	return pattern
}

func LogTimeRan(startTime time.Time) {
	println()
	println()
	timeSince := time.Since(startTime)
	println(color.RedBg(fmt.Sprintf("ran in %dms", timeSince.Milliseconds())))
}
