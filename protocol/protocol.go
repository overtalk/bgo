// auto-generate
// generate by "go run tools/generate-proto/main.go"
package gamepb

import "reflect"

var (
	protoIDTypes = map[Protocol_Id]reflect.Type{
		0:   reflect.TypeOf(new(PingReq)).Elem(),
		2:   reflect.TypeOf(new(CreateRoomReq)).Elem(),
		3:   reflect.TypeOf(new(ShutdownRoomReq)).Elem(),
		5:   reflect.TypeOf(new(SaveRoomReq)).Elem(),
		100: reflect.TypeOf(new(PingResp)).Elem(),
		102: reflect.TypeOf(new(CreateRoomResp)).Elem(),
		103: reflect.TypeOf(new(ShutdownRoomResp)).Elem(),
		105: reflect.TypeOf(new(SaveRoomResp)).Elem(),
	}
)

func GetProtoReflectType(protoID Protocol_Id) reflect.Type {
	if typ, ok := protoIDTypes[protoID]; ok {
		return typ
	}
	return nil
}
