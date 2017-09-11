package provider

import (
	"fmt"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/provider/aws"
)

func NewProviderFromConfig(tarmak interfaces.Tarmak, conf *tarmakv1alpha1.Provider) (interfaces.Provider, error) {
	if conf.AWS != nil {
		return aws.NewFromConfig(tarmak, conf)
	}
	return nil, fmt.Errorf("Unknown provider '%s'", conf.Name)
}
