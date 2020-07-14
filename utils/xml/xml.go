package xmlutil

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

func ParseXml(path string, v interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// parse
	return xml.Unmarshal(data, v)
}
