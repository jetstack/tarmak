package provider

import (
	"fmt"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/provider/amazon"
)

func NewFromConfig(tarmak interfaces.Tarmak, conf *tarmakv1alpha1.Provider) (interfaces.Provider, error) {
	var provider interfaces.Provider
	var err error

	if conf.Amazon != nil {
		if provider != nil {
			return nil, fmt.Errorf("provider '%s' has configuration options for to different clouds", conf.Name)
		}
		provider, err = amazon.NewFromConfig(tarmak, conf)
	}

	if provider == nil {
		return nil, fmt.Errorf("Unknown provider '%s'", conf.Name)
	}

	return provider, err
}
