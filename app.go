package gobay

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

// A Key represents a key for a Extention.
type Key string

// Extention like db, cache
type Extention interface {
	Object() interface{}
	Application() *Application
	Init(app *Application) error
	Close() error
}

// Application struct
type Application struct {
	rootPath    string
	env         string
	config      *viper.Viper
	extentions  map[Key]Extention
	initialized bool
	closed      bool
	mu          sync.Mutex
}

// Get the extention at the specified key, return nil when the component doesn't exist
func (d *Application) Get(key Key) Extention {
	ext, _ := d.GetOK(key)
	return ext
}

// GetOK the extention at the specified key, return false when the component doesn't exist
func (d *Application) GetOK(key Key) (Extention, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	ext, ok := d.extentions[key]
	if !ok {
		return nil, ok
	}
	return ext, ok
}

// Config returns the viper config for this application
func (d *Application) Config() *viper.Viper {
	return d.config
}

// Init the application and its extentions with the config.
func (d *Application) Init() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.initialized {
		return nil
	}

	if err := d.initConfig(); err != nil {
		return err
	}
	if err := d.initExtentions(); err != nil {
		return err
	}
	d.initialized = true
	return nil
}

func (d *Application) initConfig() error {
	configfile := filepath.Join(d.rootPath, "config.yaml")
	config := viper.New()
	config.SetConfigFile(configfile)
	if err := config.ReadInConfig(); err != nil {
		return err
	}
	config = config.Sub(d.env)

	// add default config
	config.SetDefault("debug", false)
	config.SetDefault("testing", false)
	config.SetDefault("timezone", "UTC")
	config.SetDefault("grpc_host", "localhost")
	config.SetDefault("grpc_port", 6000)
	config.SetDefault("openapi_host", "localhost")
	config.SetDefault("openapi_port", 3000)

	// read env
	config.AutomaticEnv()

	d.config = config

	return nil
}

func (d *Application) initExtentions() error {
	for _, ext := range d.extentions {
		if err := ext.Init(d); err != nil {
			return err
		}
	}
	return nil
}

// Close close app when exit
func (d *Application) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	if err := d.closeExtentions(); err != nil {
		return err
	}
	d.closed = true
	return nil
}

func (d *Application) closeExtentions() error {
	for _, ext := range d.extentions {
		if err := ext.Close(); err != nil {
			return err
		}
	}
	return nil
}

// CreateApp create an gobay Application
func CreateApp(rootPath string, env string, exts map[Key]Extention) (*Application, error) {
	if rootPath == "" || env == "" {
		return nil, fmt.Errorf("lack of rootPath or env")
	}

	app := &Application{rootPath: rootPath, env: env, extentions: exts}

	if err := app.Init(); err != nil {
		return nil, err
	}
	return app, nil
}
