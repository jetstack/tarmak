package config

import (
	"github.com/Sirupsen/logrus"
)

type Tarmak interface {
	Log() *logrus.Entry
	RootPath() string
	Context() *Context
	Terraform() Terraform
	Packer() Packer
}

type Packer interface {
}

type Terraform interface {
}
