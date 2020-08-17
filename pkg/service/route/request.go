package route

// IRequest client request
type IRequest interface {
	GetMID() uint8
	GetAID() uint8
	GetProtoVer() uint8
	GetData() []byte
	GetSign() []byte
}
