package main

import (
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

// reads lines in from files and sorts them alphabetically and prints
func main() {
	if len(os.Args) < 2 {
		panic("must pass in lines file")
	}

	lineFile, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadAll(lineFile)
	if err != nil {
		panic(err)
	}
	dataString := string(data)
	lines := trimLines(strings.Split(dataString, "\n"))
	print("start line size ")
	println(len(lines))
	lowerToNonLowerMap, lowerList := createLowerToNonLowerMap(lines)

	// sort to lower list
	sort.Strings(lowerList)

	for i := range lowerList {
		// print original value
		println(lowerToNonLowerMap[lowerList[i]])
	}
}

func createLowerToNonLowerMap(lines []string) (map[string]string, []string) {
	res := make(map[string]string, 0)
	lowerList := make([]string, len(lines))
	for i := range lines {
		curVal := lines[i]
		lowerVal := strings.ToLower(curVal)
		lowerList[i] = lowerVal
		res[lowerVal] = curVal
	}
	return res, lowerList
}

func trimLines(lines []string) []string {
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return lines
}
