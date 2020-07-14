package fileutil

import "os"

// PathExists : if the file/dir exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func IsDir(path string) bool {
	i, err := os.Stat(path)
	if err != nil {
		return false
	}
	return i.IsDir()
}
