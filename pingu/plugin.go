package pingu

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io/ioutil"
	"path/filepath"
	"plugin"
)

const SymbolName = "New"

type Author struct {
	Email string
	Name  string
}

type Plugin interface {
	Author() Author
	Commands() Commands
	Name() string
	Tasks() Tasks
	Version() string
}

type Plugins []Plugin

func LoadPlugins(dir string) ([]func(*viper.Viper) Plugin, error) {
	plugins := make([]func(*viper.Viper) Plugin, 0)
	files, err := ioutil.ReadDir(dir)

	if err != nil {
		return plugins, errors.WithMessage(err, "unable to read directory")
	}

	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".so" {
			continue
		}

		p, err := loadPlugin(filepath.Join(dir, f.Name()))

		if err != nil {
			return plugins, errors.WithMessage(err, "unable to load plugin")
		}

		plugins = append(plugins, p)
	}

	return plugins, nil
}

func loadPlugin(path string) (func(*viper.Viper) Plugin, error) {
	p, err := plugin.Open(path)

	if err != nil {
		return nil, errors.WithMessage(err, "unable to open plugin")
	}

	symbol, err := p.Lookup(SymbolName)

	if err != nil {
		return nil, errors.WithMessage(err, "symbol lookup failed")
	}

	factory, ok := symbol.(func(*viper.Viper) Plugin)

	if !ok {
		return nil, errors.Errorf(
			"symbol %s (from %s) is %T, not func(*viper.Viper) pingu.Plugin",
			SymbolName,
			filepath.Base(path),
			symbol,
		)
	}

	return factory, nil
}
