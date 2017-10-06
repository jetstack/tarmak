// Copyright Jetstack Ltd. See LICENSE for details.
package role

type Role struct {
	// name of role
	name string

	// name prefix to use
	prefix string

	// should we use scaling groups or individual ordered instances
	Stateful bool

	AWS *RoleAWS
}

type RoleAWS struct {
	ELBIngress bool // enable ELB API internal
	ELBAPI     bool // enable ELB ingress external

	// IAM Permissions
	IAMELBFull                     bool // Full access to ELB loadbalancer config
	IAMEC2Full                     bool // Full access to all EC2 resources
	IAMEC2Read                     bool // Read access to all EC2 resources
	IAMEC2ModifyInstanceAttributes bool // Allow Instance to modify all instances parameters, TODO: This should only be allowed on the masters
}

func (r *Role) WithName(name string) *Role {
	r.name = name
	return r
}

func (r *Role) Name() string {
	return r.name
}

func (r *Role) WithPrefix(prefix string) *Role {
	r.prefix = prefix
	return r
}

func (r *Role) Prefix(delimiter string) string {
	if r.prefix == "" {
		return ""
	}
	return r.prefix + delimiter
}

func (r *Role) TFName() string {
	return r.Prefix("_") + r.Name()
}

func (r *Role) DNSName() string {
	return r.Prefix("-") + r.Name()
}

func (r *Role) HasELB() bool {
	return r.AWS.ELBIngress || r.AWS.ELBAPI
}

func (r *Role) HasEtcd() bool {
	return (r.Name() == "etcd" || r.Name() == "etcd-master")
}

func (r *Role) HasMaster() bool {
	return (r.Name() == "master" || r.Name() == "etcd-master")
}

func (r *Role) HasWorker() bool {
	return (r.Name() == "worker")
}

func (r *Role) ELBIngressExternalName() string {
	return r.Name() + "-ingress"
}

func (r *Role) ELBAPIName() string {
	return r.Name() + "-api"
}
