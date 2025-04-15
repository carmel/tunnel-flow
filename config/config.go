package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

func LoadClient(path string) (*Client, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Client
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

type Client struct {
	ClientID  string `yaml:"client_id"`
	AuthToken string `yaml:"auth_token"`
	Server    string `yaml:"server"`

	Services []struct {
		Name  string `yaml:"name"`
		Local string `yaml:"local"`
	} `yaml:"services"`
}
