package files

import (
	"os"
	"strings"
)

const PATH_SEPARATOR = string(os.PathSeparator)

func CreateDirIfNotExists(filePath string) error {
	if !FileExists(filePath) {
		return CreateDir(filePath)
	}
	return nil
}

func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func CreateDir(filePath string) error {
	return os.Mkdir(filePath, os.ModePerm)
}

func GetHomeDirLocation() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return userHomeDir
}

func JoinSegmentsOfFilePath(segments ...string) string {
	return strings.Join(segments, PATH_SEPARATOR)
}
