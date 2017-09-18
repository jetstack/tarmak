package context

import (
	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
)

func defineToolsRoles(roleMap map[string]*role.Role) {
	jenkinsRole := &role.Role{
		Stateful: true,
		AWS:      &role.RoleAWS{},
	}
	jenkinsRole.WithName("jenkins")
	roleMap[clusterv1alpha1.ServerPoolTypeJenkins] = jenkinsRole

	bastionRole := &role.Role{
		Stateful: true,
		AWS:      &role.RoleAWS{},
	}
	bastionRole.WithName("bastion")
	roleMap[clusterv1alpha1.ServerPoolTypeBastion] = bastionRole
}

func defineVaultRoles(roleMap map[string]*role.Role) {
	vaultRole := &role.Role{
		Stateful: true,
		AWS:      &role.RoleAWS{},
	}
	vaultRole.WithName("vault")
	roleMap[clusterv1alpha1.ServerPoolTypeVault] = vaultRole
}

func defineKubernetesRoles(roleMap map[string]*role.Role) {
	masterRole := &role.Role{
		Stateful: false,
		AWS: &role.RoleAWS{
			ELBAPI:     true,
			IAMEC2Full: true,
			IAMELBFull: true,
		},
	}
	masterRole.WithName("master").WithPrefix("kubernetes")
	roleMap[clusterv1alpha1.ServerPoolTypeMaster] = masterRole

	nodeRole := &role.Role{
		Stateful: false,
		AWS: &role.RoleAWS{
			ELBIngress:                     true,
			IAMEC2Read:                     true,
			IAMEC2ModifyInstanceAttributes: true,
		},
	}
	nodeRole.WithName("worker").WithPrefix("kubernetes")
	roleMap[clusterv1alpha1.ServerPoolTypeNode] = nodeRole

	etcdRole := &role.Role{
		Stateful: true,
		AWS:      &role.RoleAWS{},
	}
	etcdRole.WithName("etcd").WithPrefix("kubernetes")
	roleMap[clusterv1alpha1.ServerPoolTypeEtcd] = etcdRole

	masterEtcdRole := &role.Role{
		Stateful: false,
		AWS: &role.RoleAWS{
			ELBAPI:     true,
			IAMEC2Full: true,
			IAMELBFull: true,
		},
	}
	masterEtcdRole.WithName("etcd-master").WithPrefix("kubernetes")
	roleMap[clusterv1alpha1.ServerPoolTypeMasterEtcd] = masterEtcdRole
}
