package fileutil

import "io/ioutil"

// ForceWrite overwrite file if exist
func ForceWrite(path string, writeBytes []byte) error {
	return ioutil.WriteFile(path, writeBytes, 0777)
}
