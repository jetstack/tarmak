package main

import (
	"fmt"
	"os"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := renderTemplates(); err != nil {
		logrus.Fatalf("error templating: %s", err)
	}
}

type terraformBase struct {
	// ([a-z0-9]*)
	prefix string
	// ([a-z0-9]*)
	name string
}

func (b terraformBase) Prefix(sep string) string {
	if b.prefix == "" {
		return ""
	}
	return b.prefix + sep
}

type Role struct {
	terraformBase

	// If node groups should be created as ASG or static Instances
	ASG bool

	// enable ELB API internal
	ELBAPI bool
	// enable ELB ingress external
	ELBIngress bool

	IAMELBFull               bool
	IAMEC2EBSFull            bool
	IAMEC2InstanceAttributes bool
}

func (r *Role) Name() string {
	return r.name
}

func (r *Role) TFName() string {
	return r.Prefix("_") + r.name
}

func (r *Role) DNSName() string {
	return r.Prefix("-") + r.name
}

func (r *Role) HasELB() bool {
	return r.ELBIngress || r.ELBAPI
}

func (r *Role) HasMaster() bool {
	return (r.Name() == "master" || r.Name() == "etcd-master")
}

func (r *Role) ELBIngressExternalName() string {
	return r.name + "-ingress"
}

func (r *Role) ELBAPIName() string {
	return r.name + "-api"
}

// TODO define vault
// TODO define all (== etcd/master/worker/etcd-master)
var defaultRoles = map[string]*Role{
	"etcd": &Role{
		terraformBase: terraformBase{
			name:   "etcd",
			prefix: "kubernetes",
		},
		ASG: false, // etcd is stateful so no static instances
	},
	"master": &Role{
		terraformBase: terraformBase{
			name:   "master",
			prefix: "kubernetes",
		},
		IAMEC2EBSFull:            true,
		IAMEC2InstanceAttributes: true,
		ASG:    true,
		ELBAPI: true,
	},
	"worker": &Role{
		terraformBase: terraformBase{
			name:   "worker",
			prefix: "kubernetes",
		},
		IAMEC2InstanceAttributes: true,
		ASG:        true,
		ELBIngress: true,
	},
}

// This represents an additional (EBS/GCPPD) volume

type Volume struct {
	Name   string
	Size   int // Size in GB
	aws    *VolumeAWS
	device string
}

func (v *Volume) Device() string {
	// TODO: Support multiple device names
	return "/dev/sdd"

	if v.aws != nil {
		return v.aws.Type
	}
	panic("no type specified")
}

func (v *Volume) Type() string {
	if v.aws != nil {
		return v.aws.Type
	}
	panic("no type specified")
}

type VolumeAWS struct {
	Type string // gp2/st1
	// TODO: io1 (*needs more arguments) but would be good for at least etcd data dir
}

// This represents a seperate node group
type nodeGroup struct {
	terraformBase

	name string
	role *Role

	volumes []Volume

	AWS *nodeGroupAWS
}

type nodeGroupAWS struct {
}

func (n *nodeGroup) Role() *Role {
	return n.role
}

func (n *nodeGroup) Volumes() []Volume {
	letters := "defghijklmnop"
	if len(n.volumes) >= 10 {
		panic("Max 10 volumes are currently supported")
	}
	for pos, _ := range n.volumes {
		n.volumes[pos].device = fmt.Sprintf("/dev/sd%r", letters[pos])
	}
	return n.volumes
}

// This returns the unprefixed name
func (n *nodeGroup) bareName() string {
	if n.name == "" {
		return n.Role().Name()
	}
	return n.name
}

// This returns a DNS compatible name
func (n *nodeGroup) DNSName() string {
	return n.role.Prefix("-") + n.bareName()
}

// This returns a TF compatible name
func (n *nodeGroup) TFName() string {
	return n.role.Prefix("_") + n.bareName()
}

func getRoles(nodeGroups []nodeGroup) (roles []*Role) {
	exists := make(map[*Role]bool)
	for _, ng := range nodeGroups {
		if _, ok := exists[ng.role]; !ok {
			exists[ng.role] = true
			roles = append(roles, ng.role)
		}
	}
	return roles
}

func validateNodeGroups(nodeGroup []nodeGroup) error {
	return nil
}

func renderTemplates() error {

	//contents, err := ioutil.ReadFile("node_group.tf.template")
	//if err != nil {
	//	return err
	//}

	templates := template.Must(template.New("test").Funcs(sprig.TxtFuncMap()).ParseGlob("*.tf.template"))

	baseTemplate := "node_group.tf.template"
	tpl := templates.Lookup(baseTemplate)

	f, err := os.Create("result.tf")
	if err != nil {
		return err
	}
	defer f.Close()

	nodeGroups := []nodeGroup{
		nodeGroup{
			role: defaultRoles["master"],
			volumes: []Volume{
				Volume{
					Name: "docker",
					Size: 20,
					aws: &VolumeAWS{
						Type: "gp2",
					},
				},
			},
		},
		nodeGroup{
			role: defaultRoles["worker"],
			volumes: []Volume{
				Volume{
					Name: "docker",
					Size: 50,
					aws: &VolumeAWS{
						Type: "gp2",
					},
				},
			},
		},
		nodeGroup{
			role: defaultRoles["worker"],
			name: "workercheap",
		},
		nodeGroup{
			role: defaultRoles["etcd"],
			volumes: []Volume{
				Volume{
					Name: "data",
					Size: 20,
					aws: &VolumeAWS{
						Type: "gp2",
					},
				},
			},
		},
	}

	if err := tpl.Execute(
		f,
		map[string]interface{}{
			"NodeGroups": nodeGroups,
			"Roles":      getRoles(nodeGroups),
		},
	); err != nil {
		return err
	}

	return nil

}
