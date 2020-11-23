package main

import (
	"dev_tools/files"
	"github.com/fatih/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type RegexMatch struct {
	FilePath              string
	MatchValue            string
	MatchValueWithPadding string
}

func main() {
	if len(os.Args) < 3 {
		panic("must pass pattern and directory e.g. `pf ^.*test$ test_dir/`")
	}

	pattern := os.Args[1]
	findDir := os.Args[2]

	regexPattern, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	fileSearchWaitGroup := &sync.WaitGroup{}
	fileChan := make(chan *RegexMatch, 0)

	// start listening in file chan
	go printFoundValues(fileChan, fileSearchWaitGroup)

	if files.IsDir(findDir) {
		fileSearchWaitGroup.Add(1)
		go searchDirAsync(findDir, regexPattern, fileChan, fileSearchWaitGroup)
	} else {
		fileSearchWaitGroup.Add(1)
		go searchFileAsync(findDir, regexPattern, fileChan, fileSearchWaitGroup)
	}

	fileSearchWaitGroup.Wait()
}

/**
Recursively search the directory for the regex pattern in files
*/
func searchDirAsync(dirPath string, regex *regexp.Regexp, fileChan chan *RegexMatch, wg *sync.WaitGroup) {
	dirFiles, err := ioutil.ReadDir(dirPath)
	if err != nil {
		wg.Done()
		return
	}

	for i := range dirFiles {
		currentFile := dirFiles[i]
		currentFilePath := filepath.Join(dirPath, currentFile.Name())

		if files.IsDir(currentFilePath) {
			wg.Add(1)
			go searchDirAsync(currentFilePath, regex, fileChan, wg)
		} else {
			wg.Add(1)
			go searchFileAsync(currentFilePath, regex, fileChan, wg)
		}
	}
	wg.Done()
}

/**
searches the given file as for the given pattern
if we find pattern value in file data we will pass file location to fileChan
*/
func searchFileAsync(filePath string, pattern *regexp.Regexp, fileChan chan *RegexMatch, wg *sync.WaitGroup) {
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
		}
	} else {

		// only on error or not containing value will we close wait group here
		// otherwise printing go routine will close wait group after printing.
		wg.Done()
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

func printFoundValues(fileChan chan *RegexMatch, fileSearchWaitGroup *sync.WaitGroup) {
	for file := range fileChan {
		path := file.FilePath
		matchValue := file.MatchValue
		matchValueWithPadding := file.MatchValueWithPadding
		color.Blue(path)
		println(strings.ReplaceAll(matchValueWithPadding, matchValue, color.HiMagentaString(matchValue)))

		// this will dec wait group so we can be sure to print
		// file not containing value will close wait group
		fileSearchWaitGroup.Done()
	}
}
