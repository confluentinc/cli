package shared

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl"
	jsonParser "github.com/hashicorp/hcl/json/parser"
	"github.com/hashicorp/hcl/hcl/printer"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/confluentinc/cli/log"
	"path"
)

const (
	defaultConfigFile = "~/.confluent/config.hcl"
)

var ErrNoConfig = fmt.Errorf("no config file exists")

type Config struct {
	MetricSink MetricSink  `json:"-" hcl:"-"`
	Logger     *log.Logger `json:"-" hcl:"-"`
	Filename   string      `json:"-" hcl:"-"`
	AuthURL    string      `json:"auth_url" hcl:"auth_url"`
	AuthToken  string      `json:"auth_token" hcl:"auth_token"`
	Auth       *AuthConfig `json:"auth" hcl:"auth"`
}

func (c *Config) Load() error {
	filename, err := c.getFilename()
	if err != nil {
		return err
	}
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNoConfig
		}
		return errors.Wrapf(err, "unable to read config file: %s", filename)
	}
	err = hcl.Unmarshal(input, c)
	if err != nil {
		return errors.Wrapf(err, "unable to parse config file: %s", filename)
	}
	return nil
}

func (c *Config) Save() error {
	cfg, err := json.Marshal(c)
	if err != nil {
		return errors.Wrapf(err, "unable to marshal config")
	}
	ast, err := jsonParser.Parse(cfg)
	if err != nil {
		return errors.Wrapf(err, "unable to parse config")
	}
	filename, err := c.getFilename()
	if err != nil {
		return err
	}
	err = os.MkdirAll(path.Dir(filename), 0700)
	if err != nil {
		return errors.Wrapf(err, "unable to create config directory: %s", filename)
	}
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrapf(err, "unable to create config file: %s", filename)
	}
	defer f.Close()
	err = printer.Fprint(f, ast)
	if err != nil {
		return errors.Wrapf(err, "unable to write config to file: %s", filename)
	}
	return nil
}

func (c *Config) getFilename() (string, error) {
	if c.Filename == "" {
		c.Filename = defaultConfigFile
	}
	filename, err := homedir.Expand(c.Filename)
	if err != nil {
		return "", err
	}
	return filename, nil
}
