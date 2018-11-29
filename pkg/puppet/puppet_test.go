// Copyright Jetstack Ltd. See LICENSE for details.
package puppet

import (
	"strings"
	"testing"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
)

type featureGateMap struct {
	*testing.T
	conf           *clusterv1alpha1.ClusterKubernetesClusterAutoscaler
	globalGates    map[string]bool
	componentGates map[string]bool
	usePodPriority bool
}

func TestOIDCFields(t *testing.T) {
	c := clusterv1alpha1.ClusterKubernetes{
		APIServer: &clusterv1alpha1.ClusterKubernetesAPIServer{
			OIDC: &clusterv1alpha1.ClusterKubernetesAPIServerOIDC{
				IssuerURL:      "http://123",
				ClientID:       "client_id",
				SigningAlgs:    []string{"alg1", "alg2"},
				GroupsPrefix:   "groups-prefix",
				GroupsClaim:    "groups-claim",
				UsernamePrefix: "username-prefix",
				UsernameClaim:  "username-claim",
			},
		},
	}

	d := hieraData{}

	kubernetesClusterConfig(&c, &d)

	count := 0
	for _, v := range d.variables {
		if strings.Contains(v, "oidc") {
			count += 1
		}
	}

	if act, exp := count, 7; act != exp {
		t.Fatalf("unexpected number of variables: exp:%d act:%d", exp, act)
	}

}

func TestFeatureGatesString(t *testing.T) {
	f := &featureGateMap{
		T: t,
	}

	f.testFeatureMap("")

	f.componentGates = map[string]bool{
		"A": true,
		"B": false,
	}
	f.testFeatureMap(`
  A: true
  B: false`)

	f.componentGates = map[string]bool{}
	f.globalGates = map[string]bool{
		"B": true,
		"C": false,
	}
	f.testFeatureMap(`
  B: true
  C: false`)

	// component gate overrides global gate
	f.componentGates = map[string]bool{
		"A": true,
		"B": false,
	}
	f.testFeatureMap(`
  A: true
  B: false
  C: false`)

	f.usePodPriority = true
	f.testFeatureMap(`
  A: true
  B: false
  C: false`)

	f.conf = &clusterv1alpha1.ClusterKubernetesClusterAutoscaler{}
	f.testFeatureMap(`
  A: true
  B: false
  C: false`)

	f.conf.Overprovisioning = &clusterv1alpha1.ClusterKubernetesClusterAutoscalerOverprovisioning{
		Enabled: false,
	}
	f.testFeatureMap(`
  A: true
  B: false
  C: false`)

	f.conf.Overprovisioning = &clusterv1alpha1.ClusterKubernetesClusterAutoscalerOverprovisioning{
		Enabled: true,
	}
	f.testFeatureMap(`
  A: true
  B: false
  C: false
  PodPriority: true`)

	f.usePodPriority = false
	f.testFeatureMap(`
  A: true
  B: false
  C: false`)
}

func (f *featureGateMap) testFeatureMap(exp string) {
	got := featureGatesString(f.globalGates, f.componentGates, f.usePodPriority, f.conf)
	if got != exp {
		f.Errorf("feature flags strings do not match\nexp=%s\ngot=%s", exp, got)
	}
}
