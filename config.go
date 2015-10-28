package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Upload struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type Config struct {
	ClientId     string   `yaml:"client_id"`
	ClientEmail  string   `yaml:"client_email"`
	PrivateKeyId string   `yaml:"private_key_id"`
	PrivateKey   string   `yaml:"private_key"`
	Type         string   `yaml:"type"`
	Folder       string   `yaml:"folder"`
	Uploads      []Upload `yaml:"uploads"`
}

func ParseConfig(path string) (*Config, error) {
	config := &Config{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, err
}
