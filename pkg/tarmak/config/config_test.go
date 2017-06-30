package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestDefaultConfigOmitEmpty(t *testing.T) {
	c := DefaultConfigHub()

	y, err := yaml.Marshal(c)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if strings.Contains(string(y), "null") || strings.Contains(string(y), "\"\"") {
		t.Error("yaml contains empty values, probably forgot omitempty")
	}

	c = DefaultConfigSingle()

	y, err = yaml.Marshal(c)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if strings.Contains(string(y), "null") || strings.Contains(string(y), "\"\"") {
		t.Error("yaml contains empty values, probably forgot omitempty")
	}
}

func TestDefaultConfig_Validate(t *testing.T) {
	c := DefaultConfigHub()

	err := c.Validate()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	c = DefaultConfigSingle()
	err = c.Validate()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

}
