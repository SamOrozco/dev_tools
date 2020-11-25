package main

import "strings"

// MATCHER FUNCS
var (
	matcherMap = map[MatchType]func(string, string) bool{
		Equals: func(configuredVal, testVal string) bool {
			return configuredVal == testVal
		},
		Contains: func(configuredVal, testVal string) bool {
			return strings.Contains(testVal, configuredVal)
		},
		StartsWith: func(configuredVal, testVal string) bool {
			return strings.HasPrefix(testVal, configuredVal)
		},
		EndsWith: func(configuredVal, testVal string) bool {
			return strings.HasSuffix(testVal, configuredVal)
		},
	}
)

type Matcher interface {
	Path(path string) bool
}

type simplePathMatcher struct {
	matchValue string
	matchFunc  func(val, testVal string) bool
}

func NewSimplePathMatcher(val string, matchType MatchType) Matcher {
	return &simplePathMatcher{
		matchValue: val,
		matchFunc:  matcherMap[matchType],
	}
}

func (s simplePathMatcher) Path(path string) bool {
	return s.matchFunc(s.matchValue, path)
}
