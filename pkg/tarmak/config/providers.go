package config

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
)

func NewAWSProfileProvider(name string, profile string) *tarmakv1alpha1.Provider {
	return &tarmakv1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		AWS: &tarmakv1alpha1.ProviderAWS{
			Profile: profile,
		},
	}
}
