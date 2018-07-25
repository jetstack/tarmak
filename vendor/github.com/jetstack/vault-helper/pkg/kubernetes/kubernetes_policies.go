// Copyright Jetstack Ltd. See LICENSE for details.
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

func (k *Kubernetes) DeletePolicy(p *Policy) error {
	err := k.vaultClient.Sys().DeletePolicy(p.Name)
	if err != nil {
		return fmt.Errorf("error deleting policy '%s': %v", p.Name, err)
	}

	return nil
}

func (k *Kubernetes) ReadPolicy(p *Policy) (string, error) {
	policy, err := k.vaultClient.Sys().GetPolicy(p.Name)
	if err != nil {
		return "", fmt.Errorf("error reading policy '%s': %v", p.Name, err)
	}

	return policy, nil
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

func (k *Kubernetes) deletePolicies() error {
	var result *multierror.Error

	for _, p := range []*Policy{
		k.etcdPolicy(),
		k.masterPolicy(),
		k.workerPolicy(),
	} {
		if err := k.DeletePolicy(p); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (k *Kubernetes) ensureDryRunPolicies() (bool, error) {
	var result *multierror.Error

	for _, p := range []*Policy{k.etcdPolicy(), k.masterPolicy(), k.workerPolicy()} {
		policy, err := k.ReadPolicy(p)
		if err != nil {
			result = multierror.Append(result, err)
		} else if policy != p.Policy() {
			return true, result.ErrorOrNil()
		}

	}

	return false, result.ErrorOrNil()
}

func (k *Kubernetes) etcdPolicy() *Policy {
	role := "etcd"
	return &Policy{
		Name: fmt.Sprintf("%s/%s", k.clusterID, role),
		Role: role,
		Policies: []*policyPath{
			&policyPath{
				path:         filepath.Join(k.etcdKubernetesBackend.Path(), "sign/server"),
				capabilities: []string{"create", "read", "update"},
			},
			&policyPath{
				path:         filepath.Join(k.etcdOverlayBackend.Path(), "sign/server"),
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
				path:         filepath.Join(k.etcdKubernetesBackend.Path(), "sign/client"),
				capabilities: []string{"create", "read", "update"},
			},
			&policyPath{
				path:         k.secretsBackend.ServiceAccountsPath(),
				capabilities: []string{"read"},
			},
			&policyPath{
				path:         k.secretsBackend.EncryptionConfigPath(),
				capabilities: []string{"read"},
			},
		},
	}

	// add master roles
	for _, k8sRole := range []string{"kube-apiserver", "kube-scheduler", "kube-controller-manager", "admin"} {
		p.Policies = append(
			p.Policies,
			&policyPath{
				path:         filepath.Join(k.kubernetesBackend.Path(), "sign", k8sRole),
				capabilities: []string{"create", "read", "update"},
			},
		)
	}

	// allow to get a api proxy certificate
	p.Policies = append(
		p.Policies,
		&policyPath{
			path:         filepath.Join(k.kubernetesAPIProxyBackend.Path(), "sign", "kube-apiserver"),
			capabilities: []string{"create", "read", "update"},
		},
	)

	// adds the roles from the worker
	p.Policies = append(p.Policies, k.workerPolicyPaths()...)

	return p
}

func (k *Kubernetes) workerPolicyPaths() []*policyPath {
	return []*policyPath{
		&policyPath{
			path:         filepath.Join(k.kubernetesBackend.Path(), "sign/kubelet"),
			capabilities: []string{"create", "read", "update"},
		},
		&policyPath{
			path:         filepath.Join(k.kubernetesBackend.Path(), "sign/kube-proxy"),
			capabilities: []string{"create", "read", "update"},
		},
		&policyPath{
			path:         filepath.Join(k.etcdOverlayBackend.Path(), "sign/client"),
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
