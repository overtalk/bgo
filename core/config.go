package core

import "encoding/xml"

type Config struct {
	XMLName xml.Name `xml:"xml"`
	Modules struct {
		ModuleConfBaseDir string `xml:"module_conf_base_dir,attr"`
		Module            []struct {
			Name string `xml:"name,attr"`
			Conf string `xml:"conf,attr"`
		} `xml:"module"`
	} `xml:"modules"`
}
