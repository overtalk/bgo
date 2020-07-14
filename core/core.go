package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/overtalk/bgo/pkg/log"
	"github.com/overtalk/bgo/utils/xml"
)

var (
	once          sync.Once
	pluginManager *ModuleManager
)

type ModuleManager struct {
	notifyChan chan os.Signal
	cfg        *Config
	configPath string             // basic config path
	moduleList map[string]IModule // module instances
}

func GetCore() *ModuleManager {
	once.Do(func() {
		pluginManager = &ModuleManager{
			moduleList: make(map[string]IModule),
		}
	})

	return pluginManager
}

func (this *ModuleManager) RegisterModule(moduleName string, module IModule) {
	module.SetName(moduleName)

	if this.FindModule(moduleName) != nil {
		log.Fatalf("repeated module name : %s\n", moduleName)

	}

	this.moduleList[moduleName] = module
}

func (this *ModuleManager) DeregisterModule(name string) {
	moduleToDel := this.FindModule(name)
	if moduleToDel != nil {
		if err := moduleToDel.PreShut(); err != nil {
			logpkg.Error("preShut error", zap.Error(err), zap.String("module", name))
		}

		if err := moduleToDel.Shut(); err != nil {
			logpkg.Error("shut error", zap.Error(err), zap.String("module", name))
		}

		delete(this.moduleList, name)
	}

	logpkg.Debug("deregister module", zap.String("module", name))
}

func (this *ModuleManager) FindModule(name string) IModule {
	return this.moduleList[name]
}

func (this *ModuleManager) Start() error {
	for moduleName, _ := range this.moduleList {
		logpkg.Debug("register module", zap.String("module", moduleName))
	}

	funcMap := []func() error{
		this.loadConfig,
		this.init,
		this.loadRelatedModules,
		this.preTicker,
	}

	for _, function := range funcMap {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

func (this *ModuleManager) Ticker() {
	for _, module := range this.moduleList {
		go func(notifyChan chan os.Signal, module IModule) {
			for {
				duration, flag := module.Ticker()
				switch flag {
				case Continue:
					time.Sleep(duration)
				case Stop:
					logpkg.Debug("module stop ticker", zap.String("module", module.GetName()))
					return
				default:
					logpkg.Debug("module shutdown server", zap.String("module", module.GetName()))
					notifyChan <- syscall.SIGINT
				}
			}
		}(this.notifyChan, module)
	}
}

func (this *ModuleManager) Stop() error {
	funcMap := []func() error{
		this.preShut,
		this.shut,
	}

	for _, function := range funcMap {
		if err := function(); err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func (this *ModuleManager) SetNotifyChan(notifyChan chan os.Signal) { this.notifyChan = notifyChan }

func (this *ModuleManager) SetConfigPath(path string) {
	this.configPath = path
	cfg := &Config{}
	if err := xmlutil.ParseXml(path, cfg); err != nil {
		logpkg.Fatal("load core config error", zap.Error(err), zap.Any("path", path))
	}
	this.cfg = cfg
}

// ------------------- private func -------------------
func (this *ModuleManager) loadConfig() error {
	for _, module := range this.cfg.Modules.Module {
		path := filepath.Join(this.cfg.Modules.ModuleConfBaseDir, module.Conf)
		m := this.FindModule(module.Name)
		if m == nil {
			logpkg.Fatal("load config for module", zap.String("module", module.Name))
		}
		if err := m.LoadConfig(path); err != nil {
			logpkg.Error("load config", zap.String("module", module.Name), zap.Error(err))
			return err
		}
	}

	return nil
}

func (this *ModuleManager) init() error {
	// initialize all modules
	for _, module := range this.moduleList {
		if module == nil {
			continue
		}

		if err := module.Init(); err != nil {
			return err
		}
	}

	return nil
}

func (this *ModuleManager) loadRelatedModules() error {
	for _, module := range this.moduleList {
		if module == nil {
			continue
		}

		if err := module.LoadRelatedModules(); err != nil {
			return err
		}
	}

	return nil
}

func (this *ModuleManager) preTicker() error {
	for name, module := range this.moduleList {
		if module == nil {
			continue
		}

		if err := module.PreTicker(); err != nil {
			logpkg.Error("pre-ticker error", zap.Any("module", name))
			return err
		}
	}

	return nil
}

func (this *ModuleManager) preShut() error {
	for _, module := range this.moduleList {
		if module == nil {
			continue
		}

		if err := module.PreShut(); err != nil {
			return err
		}
	}

	return nil
}

func (this *ModuleManager) shut() error {
	for _, module := range this.moduleList {
		if module == nil {
			continue
		}

		if err := module.Shut(); err != nil {
			return err
		}
	}

	return nil
}
