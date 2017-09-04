package config

import (
	"io/ioutil"
	"testing"
)

func TestNewAWSConfigClusterSingle(t *testing.T) {
	c := NewAWSConfigClusterSingle()

	tmpfile, err := ioutil.TempFile("", "tarmak.yaml")
	if err != nil {
		t.Fatal("unexpected error creating temp file: ", err)
	}

	err = writeYAML(c, tmpfile)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	log.Infof("wrote configuration to: %s\n", tmpfile.Name())
}
