package cluster

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
	roleMap[clusterv1alpha1.InstancePoolTypeJenkins] = jenkinsRole

	bastionRole := &role.Role{
		Stateful: true,
		AWS:      &role.RoleAWS{},
	}
	bastionRole.WithName("bastion")
	roleMap[clusterv1alpha1.InstancePoolTypeBastion] = bastionRole
}

func defineVaultRoles(roleMap map[string]*role.Role) {
	vaultRole := &role.Role{
		Stateful: true,
		AWS:      &role.RoleAWS{},
	}
	vaultRole.WithName("vault")
	roleMap[clusterv1alpha1.InstancePoolTypeVault] = vaultRole
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
	roleMap[clusterv1alpha1.InstancePoolTypeMaster] = masterRole

	workerRole := &role.Role{
		Stateful: false,
		AWS: &role.RoleAWS{
			ELBIngress:                     true,
			IAMEC2Read:                     true,
			IAMEC2ModifyInstanceAttributes: true,
		},
	}
	workerRole.WithName("worker").WithPrefix("kubernetes")
	roleMap[clusterv1alpha1.InstancePoolTypeWorker] = workerRole

	etcdRole := &role.Role{
		Stateful: true,
		AWS:      &role.RoleAWS{},
	}
	etcdRole.WithName("etcd").WithPrefix("kubernetes")
	roleMap[clusterv1alpha1.InstancePoolTypeEtcd] = etcdRole

	masterEtcdRole := &role.Role{
		Stateful: false,
		AWS: &role.RoleAWS{
			ELBAPI:     true,
			IAMEC2Full: true,
			IAMELBFull: true,
		},
	}
	masterEtcdRole.WithName("etcd-master").WithPrefix("kubernetes")
	roleMap[clusterv1alpha1.InstancePoolTypeMasterEtcd] = masterEtcdRole
}
