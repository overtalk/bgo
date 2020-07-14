package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"

	gamepb "github.com/overtalk/bgo/protocol"
	"github.com/overtalk/bgo/utils/file"
)

const (
	wStringFormat = `
// auto-generate
// generate by "go run tools/generate-proto/main.go"
package gamepb

import "reflect"

var (
	protoIDTypes = map[Protocol_Id]reflect.Type {
%s
}
)

func GetProtoReflectType(protoID Protocol_Id) reflect.Type {
	if typ, ok := protoIDTypes[protoID]; ok {
		return typ
	}
	return nil
}
`
)

func main() {
	var (
		typeFields string
		wString    string
		maxMsgType int32
	)

	for msgType := range gamepb.Protocol_Id_name {
		if msgType > maxMsgType {
			maxMsgType = msgType
		}
	}

	for i := int32(0); i <= maxMsgType; i++ {
		if name, ok := gamepb.Protocol_Id_name[i]; ok {
			typeFields += fmt.Sprintf("        %d: reflect.TypeOf(new(%s)).Elem(),\n", i, name)
		}
	}

	wString = fmt.Sprintf(wStringFormat, typeFields)

	var (
		filename = "protocol/protocol.go"
	)
	fileutil.Del(filename)
	err := ioutil.WriteFile(filename, []byte(wString), 0666)
	if err != nil {
		log.Fatal("Write File Error:", err)
	}

	cmd := exec.Command("go", "fmt", filename)
	if _, err = cmd.Output(); err != nil {
		log.Fatal(err)
	}
}
