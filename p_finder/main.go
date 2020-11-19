package main

import (
	"dev_tools/files"
	"github.com/labstack/gommon/color"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

type RegexMatch struct {
	FilePath   string
	MatchValue string
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
	go func() {
		for file := range fileChan {
			path := file.FilePath
			matchValue := file.MatchValue
			println(color.Blue(path))
			println(color.Yellow(matchValue))

			// this will dec wait group so we can be sure to print
			// file not containing value will close wait group
			fileSearchWaitGroup.Done()
		}
	}()

	// walk file path and for every file search for value
	err = filepath.Walk(findDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// if file not dir look for value
		if !info.IsDir() {
			fileSearchWaitGroup.Add(1)
			go searchFile(path, regexPattern, fileChan, fileSearchWaitGroup)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}
	fileSearchWaitGroup.Wait()
}

// if file contains regex pattern write file path to chan
func searchFile(filePath string, pattern *regexp.Regexp, fileChan chan *RegexMatch, wg *sync.WaitGroup) {
	dataBytes, err := files.ReadBytesFromFile(filePath)
	if err != nil {
		wg.Done()
		return
	}
	if pattern.Match(dataBytes) {
		fileChan <- &RegexMatch{
			FilePath:   filePath,
			MatchValue: getPrintValue(dataBytes, pattern),
		}
	} else {

		// only on error or not containing value will we close wait group here
		// otherwise printing go routine will close wait group after printing.
		wg.Done()
	}
}

func getPrintValue(dataBytes []byte, pattern *regexp.Regexp) string {
	return string(pattern.Find(dataBytes))
}
