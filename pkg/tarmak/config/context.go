package config

import ()

type Context struct {
	Name   string  `yaml:"name,omitempty"` // only alphanumeric lowercase
	Stacks []Stack `yaml:"stacks,omitempty"`

	Contact string `yaml:"contact,omitempty"`
	Project string `yaml:"project,omitempty"`

	BaseImage string `yaml:"baseImage,omitempty"`
}
