package files

import (
	"io/ioutil"
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

func RemoveDir(dirLocation string) error {
	return os.RemoveAll(dirLocation)
}

func ReadBytesFromFile(fileLocation string) ([]byte, error) {
	return ioutil.ReadFile(fileLocation)
}

func ReadStringFromFile(fileLocation string) (string, error) {
	data, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return "", err
	}
	return string(data), err
}

func IsDir(filePath string) bool {
	stat, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	switch mode := stat.Mode(); {
	case mode.IsDir():
		return true
	default:
		return false
	}
}
