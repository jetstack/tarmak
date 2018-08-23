// Copyright Jetstack Ltd. See LICENSE for details.
package puppet

import (
	"strings"
	"testing"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
)

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

func TestIsNVMeInstances(t *testing.T) {
	nVMeInstances := []string{
		"c5.large",
		"C5.large",
		"c5",
		"i3.metal",
		"I3.metal",
		"M5.foo",
	}

	noneNVMeInstances := []string{
		"foo",
		"t2.bla",
		"i3.foo",
		"I3.foo",
		"m5e.foo",
	}

	for _, i := range nVMeInstances {
		if !isAWSNVMeInstance(i) {
			t.Errorf("expected '%s' to be NVMe instance, got false", i)
		}
	}

	for _, i := range noneNVMeInstances {
		if isAWSNVMeInstance(i) {
			t.Errorf("expected '%s' to not be NVMe instance, got true", i)
		}
	}
}
