// Copyright Jetstack Ltd. See LICENSE for details.
package puppet

import (
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

	if len(d.variables) != 7 {
		t.Fatalf("unexpected number of variables: %d", len(d.variables))
	}

}
