package core

import "time"

type Action uint8

const (
	Continue = iota // continue
	Stop            // stop ticker
	Shutdown        // shutdown the server
)

// IModule defines the module core
type IModule interface {
	LoadConfig(path string) error
	Init() error
	LoadRelatedModules() error
	PreTicker() error
	Ticker() (time.Duration, Action)
	PreShut() error
	Shut() error
	GetName() string
	SetName(name string)
}

type Module struct{ name string }

func (module *Module) LoadConfig(path string) error    { return nil }
func (module *Module) Init() error                     { return nil }
func (module *Module) LoadRelatedModules() error       { return nil }
func (module *Module) PreTicker() error                { return nil }
func (module *Module) Ticker() (time.Duration, Action) { return 0, Stop }
func (module *Module) PreShut() error                  { return nil }
func (module *Module) Shut() error                     { return nil }
func (module *Module) GetName() string                 { return module.name }
func (module *Module) SetName(name string)             { module.name = name }
