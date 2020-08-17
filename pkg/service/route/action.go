package route

// IAction a handler for certain protocol id
type IAction interface {
	GetAID() uint8
	Handle(IRequest) IOutProtocol
}

type noneAction struct{}

func (*noneAction) GetAID() uint8 { return 0 }
func (*noneAction) Handle(_ IRequest) IOutProtocol {
	return BytesOutProtocol(nil)
}

// NoneAction an action doing nothing
var NoneAction IAction = &noneAction{}
