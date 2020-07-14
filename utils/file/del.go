package fileutil

import "os"

func Del(file string) error {
	return os.Remove(file)
}
