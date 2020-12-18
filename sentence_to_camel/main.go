package main

import (
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		panic("no string passed")
	}
	sentence := os.Args[1]
	println(convertSentenceToCamelCase(sentence))
}

func convertSentenceToCamelCase(sent string) string {
	segments := strings.Fields(sent)
	bldr := strings.Builder{}
	for i := range segments {
		currentSeg := segments[i]
		if i != 0 {
			currentSeg = strToUpper(currentSeg)
		}
		bldr.WriteString(currentSeg)
	}
	return bldr.String()
}

func strToUpper(val string) string {
	if len(val) == 1 {
		return strings.ToUpper(val)
	}
	return strings.ToUpper(val[:1]) + val[1:]
}
