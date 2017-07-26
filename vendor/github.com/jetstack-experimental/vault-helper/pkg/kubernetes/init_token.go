package kubernetes

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
)

type InitToken struct {
	Role       string
	Policies   []string
	kubernetes *Kubernetes
	token      *string
}

func (i *InitToken) Ensure() (bool, error) {
	var result error

	ensureInitToken := func() (bool, error) {
		_, written, err := i.InitToken()
		return written, err
	}

	change := false
	for _, f := range []func() error{
		i.writeTokenRole,
		i.writeInitTokenPolicy,
	} {
		if err := f(); err != nil {
			result = multierror.Append(result, err)
		}
		if written, err := ensureInitToken(); err != nil {
			result = multierror.Append(result, err)
		} else if written {
			change = true
		}
	}

	return change, result
}

func (i *InitToken) Name() string {
	return fmt.Sprintf("%s-%s", i.kubernetes.clusterID, i.Role)
}

func (i *InitToken) NamePath() string {
	return fmt.Sprintf("%s/%s", i.kubernetes.clusterID, i.Role)
}

func (i *InitToken) CreatePath() string {
	return filepath.Join("auth/token/create", i.Name())
}

func (i *InitToken) Path() string {
	return filepath.Join("auth/token/roles", i.Name())
}

func (i *InitToken) writeTokenRole() error {
	policies := i.Policies
	policies = append(policies, "default")

	writeData := map[string]interface{}{
		"period":           fmt.Sprintf("%ds", int(i.kubernetes.MaxValidityComponents.Seconds())),
		"orphan":           true,
		"allowed_policies": strings.Join(policies, ","),
		"path_suffix":      i.NamePath(),
	}

	_, err := i.kubernetes.vaultClient.Logical().Write(i.Path(), writeData)
	if err != nil {
		return fmt.Errorf("error writing token role %s: %s", i.Path(), err)
	}

	return nil
}

func (i *InitToken) writeInitTokenPolicy() error {
	p := &Policy{
		Name: fmt.Sprintf("%s-creator", i.NamePath()),
		Policies: []*policyPath{
			&policyPath{
				path:         i.CreatePath(),
				capabilities: []string{"create", "read", "update"},
			},
		},
	}
	return i.kubernetes.WritePolicy(p)
}

func (i *InitToken) InitToken() (string, bool, error) {
	if i.token != nil {
		return *i.token, false, nil
	}

	// get init token from generic
	token, written, err := i.kubernetes.secretsGeneric.InitToken(i.Name(), i.Role, []string{fmt.Sprintf("%s-creator", i.NamePath())})
	if err != nil {
		return "", false, err
	}

	i.token = &token
	return token, written, nil
}
