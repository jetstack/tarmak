// Copyright Jetstack Ltd. See LICENSE for details.
package kubernetes

import (
	"fmt"
	"strings"
)

type Policy struct {
	Name     string
	Policies []*policyPath
	Role     string
}

type policyPath struct {
	path         string
	capabilities []string
}

func (pp *policyPath) String() string {
	capabilities := make([]string, len(pp.capabilities))
	for pos, cap := range pp.capabilities {
		capabilities[pos] = fmt.Sprintf(`"%s"`, cap)
	}

	return fmt.Sprintf(
		`path "%s" {
  capabilities = [%s]
}
`,
		pp.path,
		strings.Join(capabilities, ", "),
	)
}

func (p *Policy) Policy() string {
	policies := make([]string, len(p.Policies))
	for pos, pol := range p.Policies {
		policies[pos] = pol.String()
	}
	return strings.Join(policies, "\n")
}
