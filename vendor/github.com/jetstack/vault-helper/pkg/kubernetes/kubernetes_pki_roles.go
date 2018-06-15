// Copyright Jetstack Ltd. See LICENSE for details.
package kubernetes

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
)

type pkiRole struct {
	Name string
	Data map[string]interface{}
}

func (k *Kubernetes) etcdClientRole() *pkiRole {
	return &pkiRole{
		Name: "client",
		Data: map[string]interface{}{
			"use_csr_common_name": false,
			"use_csr_sans":        false,
			"allow_any_name":      true,
			"max_ttl":             constructTimeString(k.MaxValidityComponents),
			"ttl":                 constructTimeString(k.MaxValidityComponents),
			"allow_ip_sans":       true,
			"server_flag":         false,
			"client_flag":         true,
		},
	}
}

func (k *Kubernetes) etcdServerRole() *pkiRole {
	return &pkiRole{
		Name: "server",
		Data: map[string]interface{}{
			"use_csr_common_name": false,
			"use_csr_sans":        false,
			"allow_any_name":      true,
			"max_ttl":             constructTimeString(k.MaxValidityComponents),
			"ttl":                 constructTimeString(k.MaxValidityComponents),
			"allow_ip_sans":       true,
			"server_flag":         true,
			"client_flag":         true,
		},
	}
}

func (k *Kubernetes) k8sAdminRole() *pkiRole {
	return &pkiRole{
		Name: "admin",
		Data: map[string]interface{}{
			"use_csr_common_name": false,
			"enforce_hostnames":   false,
			"organization":        []string{"system:masters"},
			"allowed_domains":     []string{"admin"},
			"allow_bare_domains":  true,
			"allow_localhost":     false,
			"allow_subdomains":    false,
			"allow_ip_sans":       false,
			"server_flag":         false,
			"client_flag":         true,
			"max_ttl":             constructTimeString(k.MaxValidityAdmin),
			"ttl":                 constructTimeString(k.MaxValidityAdmin),
		},
	}
}

func (k *Kubernetes) k8sAPIServerRole() *pkiRole {
	return &pkiRole{
		Name: "kube-apiserver",
		Data: map[string]interface{}{
			"use_csr_common_name": false,
			"use_csr_sans":        false,
			"enforce_hostnames":   false,
			"allow_localhost":     true,
			"allow_any_name":      true,
			"allow_bare_domains":  true,
			"allow_ip_sans":       true,
			"server_flag":         true,
			"client_flag":         true,
			"max_ttl":             constructTimeString(k.MaxValidityComponents),
			"ttl":                 constructTimeString(k.MaxValidityComponents),
		},
	}
}

func (k *Kubernetes) k8sAPIServerProxyRole() *pkiRole {
	return &pkiRole{
		Name: "kube-apiserver",
		Data: map[string]interface{}{
			"use_csr_common_name": false,
			"use_csr_sans":        false,
			"enforce_hostnames":   false,
			"allow_localhost":     false,
			"allow_any_name":      false,
			"allow_bare_domains":  true,
			"allow_ip_sans":       false,
			"server_flag":         false,
			"client_flag":         true,
			"allowed_domains":     []string{"kube-apiserver-proxy"},
			"max_ttl":             constructTimeString(k.MaxValidityComponents),
			"ttl":                 constructTimeString(k.MaxValidityComponents),
		},
	}
}

func (k *Kubernetes) k8sKubeletRole() *pkiRole {
	return &pkiRole{
		Name: "kubelet",
		Data: map[string]interface{}{
			"use_csr_common_name": false,
			"use_csr_sans":        false,
			"enforce_hostnames":   false,
			"organization":        []string{"system:nodes"},
			"allowed_domains":     []string{"kubelet", "system:node", "system:node:*", "*.compute.internal", "*.ec2.internal"},
			"allow_bare_domains":  true,
			"allow_glob_domains":  true,
			"allow_any_name":      false,
			"allow_localhost":     false,
			"allow_subdomains":    true,
			"allow_ip_sans":       false,
			"server_flag":         true,
			"client_flag":         true,
			"max_ttl":             constructTimeString(k.MaxValidityComponents),
			"ttl":                 constructTimeString(k.MaxValidityComponents),
		},
	}
}

func (k *Kubernetes) k8sComponentRole(roleName string) *pkiRole {
	return &pkiRole{
		Name: roleName,
		Data: map[string]interface{}{
			"use_csr_common_name": false,
			"enforce_hostnames":   false,
			"allowed_domains":     []string{roleName, fmt.Sprintf("system:%s", roleName)},
			"allow_bare_domains":  true,
			"allow_localhost":     false,
			"allow_subdomains":    false,
			"allow_ip_sans":       true,
			"server_flag":         false,
			"client_flag":         true,
			"max_ttl":             constructTimeString(k.MaxValidityComponents),
			"ttl":                 constructTimeString(k.MaxValidityComponents),
		},
	}
}

// this makes sure all etcd PKI roles are setup correctly
func (k *Kubernetes) ensurePKIRolesEtcd(p *PKIVaultBackend) error {
	var err error
	var result error

	if err = p.WriteRole(k.etcdClientRole()); err != nil {
		result = multierror.Append(result, err)
	}

	if err = p.WriteRole(k.etcdServerRole()); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (k *Kubernetes) deletePKIRolesEtcd(p *PKIVaultBackend) error {
	var result error

	if err := p.DeleteRole(k.etcdClientRole()); err != nil {
		result = multierror.Append(result, err)
	}

	if err := p.DeleteRole(k.etcdServerRole()); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (k *Kubernetes) ensureDryRunPKIRolesEtcd(p *PKIVaultBackend) (bool, error) {
	var result *multierror.Error

	secret, err := p.ReadRole(k.etcdClientRole())
	if err != nil {
		result = multierror.Append(result, err)
	} else if secret == nil || secret.Data == nil || len(secret.Data) == 0 {
		return true, nil
	}

	if !secretDataMatch(secret.Data, k.etcdClientRole().Data) {
		return true, result.ErrorOrNil()
	}

	secret, err = p.ReadRole(k.etcdServerRole())
	if err != nil {
		result = multierror.Append(result, err)
	} else if len(secret.Data) == 0 {
		return true, nil
	}

	if !secretDataMatch(secret.Data, k.etcdServerRole().Data) {
		return true, result.ErrorOrNil()
	}

	return false, result.ErrorOrNil()
}

func (k *Kubernetes) pkiRoleK8s() []*pkiRole {
	return []*pkiRole{
		k.k8sAdminRole(),
		k.k8sAPIServerRole(),
		k.k8sComponentRole("kube-scheduler"),
		k.k8sComponentRole("kube-controller-manager"),
		k.k8sComponentRole("kube-proxy"),
		k.k8sKubeletRole(),
	}
}

// this makes sure all kubernetes PKI roles are setup correctly
func (k *Kubernetes) ensurePKIRolesK8S(p *PKIVaultBackend) error {
	var result error

	for _, role := range k.pkiRoleK8s() {
		if err := p.WriteRole(role); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (k *Kubernetes) deletePKIRolesK8S(p *PKIVaultBackend) error {
	var result error

	for _, role := range k.pkiRoleK8s() {
		if err := p.DeleteRole(role); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (k *Kubernetes) ensureDryRunPKIRolesK8S(p *PKIVaultBackend) (bool, error) {
	var result *multierror.Error

	for _, role := range k.pkiRoleK8s() {
		secret, err := p.ReadRole(role)
		if err != nil {
			result = multierror.Append(result, err)
		} else if len(secret.Data) == 0 {
			return true, nil
		}

		if !secretDataMatch(secret.Data, role.Data) {
			return true, result.ErrorOrNil()
		}
	}

	return false, result.ErrorOrNil()
}

// this makes sure all kubernetes API Proxy PKI roles are setup correctly
func (k *Kubernetes) ensurePKIRolesK8SAPIProxy(p *PKIVaultBackend) error {
	var result error

	roles := []*pkiRole{
		k.k8sAPIServerProxyRole(),
	}

	for _, role := range roles {
		if err := p.WriteRole(role); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (k *Kubernetes) deletePKIRolesK8SAPIProxy(p *PKIVaultBackend) error {
	var result error

	roles := []*pkiRole{
		k.k8sAPIServerProxyRole(),
	}

	for _, role := range roles {
		if err := p.DeleteRole(role); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (k *Kubernetes) ensureDryRunPKIRolesK8SAPIProxy(p *PKIVaultBackend) (bool, error) {
	secret, err := p.ReadRole(k.k8sAPIServerProxyRole())
	if err != nil {
		return false, err
	}

	if len(secret.Data) == 0 {
		return true, nil
	}

	if !secretDataMatch(secret.Data, k.k8sAPIServerProxyRole().Data) {
		return true, nil
	}

	return false, nil
}

func constructTimeString(t time.Duration) string {
	h := int(t / time.Hour)
	t = t % time.Hour
	m := int(t / time.Minute)
	t = t % time.Minute
	s := int(t / time.Second)
	return fmt.Sprintf("%dh%dm%ds", h, m, s)
}

func secretDataMatch(secretData, roleData map[string]interface{}) bool {
	for key, data := range roleData {
		d, ok := secretData[key]
		if !ok || fmt.Sprintf("%v", data) != fmt.Sprintf("%v", d) {
			return false
		}
	}

	return true
}
