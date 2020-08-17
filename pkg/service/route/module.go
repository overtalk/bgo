package route

// IModule module handler
type IModule interface {
	GetMID() uint8
	Handle(IRequest) IOutProtocol
}

// -----------------------------------------------
// noneModule
// -----------------------------------------------
type noneModule struct{}

func (*noneModule) GetMID() uint8 { return 0 }
func (*noneModule) Handle(_ IRequest) IOutProtocol {
	return BytesOutProtocol(nil)
}

// NoneModule a module doing nothing
var NoneModule IModule = &noneModule{}

// -----------------------------------------------
// baseModule
// -----------------------------------------------
type baseModule struct {
	mid     uint8
	actions map[uint8]IAction
}

// judge baseModule is a implication of IModule
var _ IModule = (*baseModule)(nil)

// NewModule create a IModule instance
func NewModule(mid uint8, acts ...IAction) IModule {
	modActions := make(map[uint8]IAction, len(acts))
	for _, v := range acts {
		modActions[v.GetAID()] = v
	}
	return &baseModule{mid: mid, actions: modActions}
}

func (m *baseModule) GetMID() uint8 {
	return m.mid
}

func (m *baseModule) Handle(r IRequest) IOutProtocol {
	actionID := r.GetAID()
	act, ok := m.actions[actionID]
	if !ok {
		act = NoneAction
		//zaplog.S.Errorf("module %d: action(%d) not found", m.mid, actionID)
	}
	return act.Handle(r)
}
