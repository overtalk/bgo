package ctcp

import (
	"encoding/xml"

	"github.com/overtalk/bgo/utils/xml"
)

type Config struct {
	XMLName    xml.Name `xml:"xml"`
	Name       string   `xml:"name"`
	Network    string   `xml:"network"`
	Host       string   `xml:"host"`
	Port       int      `xml:"port"`
	MaxConn    int      `xml:"maxConn"`
	ReadSynced bool     `xml:"readSynced"`
}

func (tcp *CTcpModule) LoadConfig(path string) error {
	cfg := &Config{}
	if err := xmlutil.ParseXml(path, cfg); err != nil {
		return err
	}

	tcp.cfg = cfg
	return nil
}
