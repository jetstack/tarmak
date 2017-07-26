package kubernetes

import (
	"fmt"
	"strings"

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
			"max_ttl":             fmt.Sprintf("%ds", int(k.MaxValidityComponents.Seconds())),
			"ttl":                 fmt.Sprintf("%ds", int(k.MaxValidityComponents.Seconds())),
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
			"max_ttl":             fmt.Sprintf("%ds", int(k.MaxValidityComponents.Seconds())),
			"ttl":                 fmt.Sprintf("%ds", int(k.MaxValidityComponents.Seconds())),
			"allow_ip_sans":       true,
			"server_flag":         true,
			"client_flag":         true,
		},
	}
}

// this makes sure all etcd PKI roles are setup correctly
func (k *Kubernetes) ensurePKIRolesEtcd(p *PKI) error {
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

func (k *Kubernetes) k8sAdminRole() *pkiRole {
	return &pkiRole{
		Name: "admin",
		Data: map[string]interface{}{
			"use_csr_common_name": false,
			"enforce_hostnames":   false,
			"organization":        "system:masters",
			"allowed_domains":     "admin",
			"allow_bare_domains":  true,
			"allow_localhost":     false,
			"allow_subdomains":    false,
			"allow_ip_sans":       false,
			"server_flag":         false,
			"client_flag":         true,
			"max_ttl":             fmt.Sprintf("%ds", int(k.MaxValidityAdmin.Seconds())),
			"ttl":                 fmt.Sprintf("%ds", int(k.MaxValidityAdmin.Seconds())),
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
			"client_flag":         false,
			"max_ttl":             fmt.Sprintf("%ds", int(k.MaxValidityComponents.Seconds())),
			"ttl":                 fmt.Sprintf("%ds", int(k.MaxValidityComponents.Seconds())),
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
			"organization":        "system:nodes",
			"allowed_domains":     strings.Join([]string{"kubelet", "system:node", "system:node:*"}, ","),
			"allow_bare_domains":  true,
			"allow_glob_domains":  true,
			"allow_localhost":     false,
			"allow_subdomains":    false,
			"server_flag":         true,
			"client_flag":         true,
			"max_ttl":             fmt.Sprintf("%ds", int(k.MaxValidityComponents.Seconds())),
			"ttl":                 fmt.Sprintf("%ds", int(k.MaxValidityComponents.Seconds())),
		},
	}
}

func (k *Kubernetes) k8sComponentRole(roleName string) *pkiRole {
	return &pkiRole{
		Name: roleName,
		Data: map[string]interface{}{
			"use_csr_common_name": false,
			"enforce_hostnames":   false,
			"allowed_domains":     strings.Join([]string{roleName, fmt.Sprintf("system:%s", roleName)}, ","),
			"allow_bare_domains":  true,
			"allow_localhost":     false,
			"allow_subdomains":    false,
			"allow_ip_sans":       false,
			"server_flag":         false,
			"client_flag":         true,
			"max_ttl":             fmt.Sprintf("%ds", int(k.MaxValidityComponents.Seconds())),
			"ttl":                 fmt.Sprintf("%ds", int(k.MaxValidityComponents.Seconds())),
		},
	}
}

// this makes sure all kubernetes PKI roles are setup correctly
func (k *Kubernetes) ensurePKIRolesK8S(p *PKI) error {
	var result error

	roles := []*pkiRole{
		k.k8sAdminRole(),
		k.k8sAPIServerRole(),
		k.k8sComponentRole("kube-scheduler"),
		k.k8sComponentRole("kube-controller-manager"),
		k.k8sComponentRole("kube-proxy"),
		k.k8sKubeletRole(),
	}

	for _, role := range roles {
		if err := p.WriteRole(role); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}
