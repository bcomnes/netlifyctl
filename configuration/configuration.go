package configuration

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
)

const DefaultConfigFileName = "netlify.toml"

type Settings struct {
	ID   string
	Path string `toml:"path,omitempty"`
}

type Context struct {
	Publish   string
	Functions string
}

type Configuration struct {
	Settings Settings
	Build    Context
	root     string
}

func (c Configuration) Root() string {
	return c.root
}

func Exist(configFile string) bool {
	pwd, err := os.Getwd()
	if err != nil {
		return false
	}

	single := filepath.Join(pwd, configFile)
	_, err = os.Stat(single)
	return err == nil
}

func Load(configFile string) (*Configuration, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var c Configuration
	c.root = pwd

	f, err := os.Open(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &c, nil
		}
		return nil, err
	}
	defer f.Close()

	if err := toml.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}

	logrus.Debugf("Parsed configuration: %+v", c)

	return &c, nil
}

func Save(configFile string, conf *Configuration) error {
	single := filepath.Join(conf.root, configFile)
	f, err := os.OpenFile(single, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(conf)
}
