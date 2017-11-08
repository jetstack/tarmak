package kubernetes

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
)

func (k *Kubernetes) WritePolicy(p *Policy) error {
	err := k.vaultClient.Sys().PutPolicy(p.Name, p.Policy())
	if err != nil {
		return fmt.Errorf("error writting policy '%s': %v", p.Name, err)
	}

	return nil
}

func (k *Kubernetes) ensurePolicies() error {
	var result error

	str := "Policies written for: "
	for _, p := range []*Policy{
		k.etcdPolicy(),
		k.masterPolicy(),
		k.workerPolicy(),
	} {
		if err := k.WritePolicy(p); err != nil {
			result = multierror.Append(result, err)
		} else {
			str += "'" + p.Role + "'  "
		}
	}
	k.Log.Infof(str)

	return result
}

func (k *Kubernetes) etcdPolicy() *Policy {
	role := "etcd"
	return &Policy{
		Name: fmt.Sprintf("%s/%s", k.clusterID, role),
		Role: role,
		Policies: []*policyPath{
			&policyPath{
				path:         filepath.Join(k.etcdKubernetesPKI.Path(), "sign/server"),
				capabilities: []string{"create", "read", "update"},
			},
			&policyPath{
				path:         filepath.Join(k.etcdOverlayPKI.Path(), "sign/server"),
				capabilities: []string{"create", "read", "update"},
			},
		},
	}
}

func (k *Kubernetes) masterPolicy() *Policy {
	role := "master"
	p := &Policy{
		Name: fmt.Sprintf("%s/%s", k.clusterID, role),
		Role: role,
		Policies: []*policyPath{
			&policyPath{
				path:         filepath.Join(k.etcdKubernetesPKI.Path(), "sign/client"),
				capabilities: []string{"create", "read", "update"},
			},
			&policyPath{
				path:         filepath.Join(k.secretsGeneric.Path(), "service-accounts"),
				capabilities: []string{"read"},
			},
		},
	}

	// add master roles
	for _, k8sRole := range []string{"kube-apiserver", "kube-scheduler", "kube-controller-manager", "admin"} {
		p.Policies = append(
			p.Policies,
			&policyPath{
				path:         filepath.Join(k.kubernetesPKI.Path(), "sign", k8sRole),
				capabilities: []string{"create", "read", "update"},
			},
		)
	}

	// adds the roles from the worker
	p.Policies = append(p.Policies, k.workerPolicyPaths()...)

	return p
}

func (k *Kubernetes) workerPolicyPaths() []*policyPath {
	return []*policyPath{
		&policyPath{
			path:         filepath.Join(k.kubernetesPKI.Path(), "sign/kubelet"),
			capabilities: []string{"create", "read", "update"},
		},
		&policyPath{
			path:         filepath.Join(k.kubernetesPKI.Path(), "sign/kube-proxy"),
			capabilities: []string{"create", "read", "update"},
		},
		&policyPath{
			path:         filepath.Join(k.etcdOverlayPKI.Path(), "sign/client"),
			capabilities: []string{"create", "read", "update"},
		},
	}
}

func (k *Kubernetes) workerPolicy() *Policy {
	role := "worker"
	return &Policy{
		Name:     fmt.Sprintf("%s/%s", k.clusterID, role),
		Role:     role,
		Policies: k.workerPolicyPaths(),
	}
}
