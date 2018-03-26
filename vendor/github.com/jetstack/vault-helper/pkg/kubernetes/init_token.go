package kubernetes

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
)

type InitToken struct {
	Role          string
	Policies      []string
	kubernetes    *Kubernetes
	token         *string
	ExpectedToken string
}

func (i *InitToken) Ensure() error {
	var result error

	// always ensure token role and init token policy is set (this is idempotent)
	if err := i.writeTokenRole(); err != nil {
		result = multierror.Append(result, fmt.Errorf("not able to write token role: %s", err))
	}
	if err := i.writeInitTokenPolicy(); err != nil {
		result = multierror.Append(result, fmt.Errorf("not able to write init token policy: %s", err))
	}
	if result != nil {
		return result
	}

	// make sure init token exists
	initToken, err := i.InitToken()
	if err != nil {
		return fmt.Errorf("not able to ensure init token: %s", err)
	}
	i.token = &initToken

	return result
}

// Get init token name
func (i *InitToken) Name() string {
	return fmt.Sprintf("%s-%s", i.kubernetes.clusterID, i.Role)
}

// Get name path suffix for token role
func (i *InitToken) namePath() string {
	return fmt.Sprintf("%s/%s", i.kubernetes.clusterID, i.Role)
}

// Construct file path for ../create
func (i *InitToken) createPath() string {
	return filepath.Join("auth/token/create", i.Name())
}

// Construct file path for ../auth
func (i *InitToken) Path() string {
	return filepath.Join("auth/token/roles", i.Name())
}

// Write token role to vault
func (i *InitToken) writeTokenRole() error {
	policies := i.Policies
	policies = append(policies, "default")

	writeData := map[string]interface{}{
		"period":           fmt.Sprintf("%ds", int(i.kubernetes.MaxValidityComponents.Seconds())),
		"orphan":           true,
		"allowed_policies": strings.Join(policies, ","),
		"path_suffix":      i.namePath(),
	}

	_, err := i.kubernetes.vaultClient.Logical().Write(i.Path(), writeData)
	if err != nil {
		return fmt.Errorf("error writing token role %s: %v", i.Path(), err)
	}

	return nil
}

// Construct policy and send to kubernetes to be written to vault
func (i *InitToken) writeInitTokenPolicy() error {
	p := &Policy{
		Name: fmt.Sprintf("%s-creator", i.namePath()),
		Policies: []*policyPath{
			&policyPath{
				path:         i.createPath(),
				capabilities: []string{"create", "read", "update"},
			},
		},
	}
	return i.kubernetes.WritePolicy(p)
}

// Return init token if token exists
// Retrieve from generic if !exists
func (i *InitToken) InitToken() (string, error) {
	if i.token != nil {
		return *i.token, nil
	}

	// get init token from generic
	token, err := i.secretsGeneric().InitToken(i.Name(), i.Role, []string{fmt.Sprintf("%s-creator", i.namePath())}, i.ExpectedToken)
	if err != nil {
		return "", err
	}

	i.token = &token
	return token, nil
}

func (i *InitToken) GetInitToken() (string, error) {
	if i.token != nil {
		return *i.token, nil
	}
	return "", fmt.Errorf("could not get init token")
}

func (i *InitToken) secretsGeneric() *Generic {
	return i.kubernetes.secretsGeneric
}
