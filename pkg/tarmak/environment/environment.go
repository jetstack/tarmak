// Copyright Jetstack Ltd. See LICENSE for details.
package environment

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/rest"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/cluster"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/provider"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
	wingclient "github.com/jetstack/tarmak/pkg/wing/client"
)

type Environment struct {
	conf *tarmakv1alpha1.Environment

	clusters []interfaces.Cluster

	sshKeyPrivate interface{}

	hubCluster interfaces.Cluster // this is the cluster that contains state/vault/tools
	provider   interfaces.Provider
	tarmak     interfaces.Tarmak

	log *logrus.Entry
}

var _ interfaces.Environment = &Environment{}

func NewFromConfig(tarmak interfaces.Tarmak, conf *tarmakv1alpha1.Environment, clusters []*clusterv1alpha1.Cluster) (*Environment, error) {
	e := &Environment{
		conf:   conf,
		tarmak: tarmak,
		log:    tarmak.Log().WithField("environment", conf.Name),
	}

	var result error

	providerConf, err := tarmak.Config().Provider(conf.Provider)
	if err != nil {
		return nil, fmt.Errorf("error finding provider '%s'", conf.Provider)
	}

	// init provider
	e.provider, err = provider.NewFromConfig(tarmak, providerConf)
	if err != nil {
		return nil, fmt.Errorf("error initializing provider '%s'", conf.Provider)
	}

	// TODO RENABLE
	//networkCIDRs := []*net.IPNet{}

	for posCluster, _ := range clusters {
		clusterConf := clusters[posCluster]
		clusterIntf, err := cluster.NewFromConfig(e, clusterConf)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		e.clusters = append(e.clusters, clusterIntf)
		if len(clusters) == 1 || clusterConf.Name == "hub" {
			e.hubCluster = clusterIntf
		}
	}
	if result != nil {
		return nil, result
	}

	return e, nil
}

func (e *Environment) Name() string {
	return e.conf.Name
}

func (c *Environment) HubName() string {
	return fmt.Sprintf("%s-%s", c.Name(), clusterv1alpha1.ClusterTypeHub)
}

func (e *Environment) Config() *tarmakv1alpha1.Environment {
	return e.conf.DeepCopy()
}

func (e *Environment) Provider() interfaces.Provider {
	return e.provider
}

func (e *Environment) Tarmak() interfaces.Tarmak {
	return e.tarmak
}

func (e *Environment) Cluster(name string) (interfaces.Cluster, error) {
	for pos, _ := range e.clusters {
		cluster := e.clusters[pos]
		if cluster.Name() == name {
			return cluster, nil
		}
	}
	return nil, fmt.Errorf("cluster '%s' in environment '%s' not found", name, e.Name())
}

func (e *Environment) validateSSHKey() error {
	bytes, err := ioutil.ReadFile(e.SSHPrivateKeyPath())
	if err != nil {
		return fmt.Errorf("unable to read ssh private key: %s", err)
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return errors.New("failed to parse PEM block containing the ssh private key")
	}

	e.sshKeyPrivate, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %s", err)
	}

	return fmt.Errorf("please implement me !!!")

}

func (e *Environment) Variables() map[string]interface{} {
	output := map[string]interface{}{}
	output["environment"] = e.Name()
	if e.conf.Contact != "" {
		output["contact"] = e.conf.Contact
	}
	if e.conf.Project != "" {
		output["project"] = e.conf.Project
	}

	for key, value := range e.provider.Variables() {
		output[key] = value
	}

	output["state_bucket"] = e.Provider().RemoteStateBucketName()
	output["state_cluster_name"] = e.hubCluster.Name()
	output["tools_cluster_name"] = e.hubCluster.Name()
	output["vault_cluster_name"] = e.hubCluster.Name()
	return output
}

func (e *Environment) ConfigPath() string {
	return filepath.Join(e.tarmak.ConfigPath(), e.Name())
}

func generateRSAKey(bitSize int, filePath string) (*rsa.PrivateKey, error) {
	reader := rand.Reader

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return nil, err
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer outFile.Close()

	if err := os.Chmod(filePath, 0600); err != nil {
		return nil, err
	}

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	if err := pem.Encode(outFile, privateKey); err != nil {
		return nil, err
	}

	return key, nil

}

func (e *Environment) SSHPrivateKey() interface{} {
	if e.sshKeyPrivate == nil {
		key, err := e.getSSHPrivateKey()
		if err != nil {
			e.log.Fatal(err)
		}
		e.sshKeyPrivate = key
	}
	return e.sshKeyPrivate
}

func (e *Environment) getSSHPrivateKey() (interface{}, error) {
	path := e.SSHPrivateKeyPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := utils.EnsureDirectory(filepath.Dir(path), 0700); err != nil {
			return nil, fmt.Errorf("error creating directory: %s", err)
		}

		sshKey, err := generateRSAKey(4096, path)
		if err != nil {
			return nil, fmt.Errorf("error generating ssh key: %s", err)
		}
		return sshKey, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to find ssh key in %s: %s", path, err)
	}

	sshKeyBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read ssh key %s: %s", path, err)
	}

	sshKey, err := ssh.ParseRawPrivateKey(sshKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse ssh key %s: %s", path, err)
	}

	return sshKey, nil
}

func (e *Environment) SSHPrivateKeyPath() string {
	if e.conf.SSH == nil || e.conf.SSH.PrivateKeyPath == "" {
		return filepath.Join(e.ConfigPath(), "id_rsa")
	}

	dir, err := e.Tarmak().HomeDirExpand(e.conf.SSH.PrivateKeyPath)
	if err != nil {
		return e.conf.SSH.PrivateKeyPath
	}
	return dir
}

func (e *Environment) Location() string {
	return e.conf.Location
}

func (e *Environment) Clusters() []interfaces.Cluster {
	return e.clusters
}

func (e *Environment) Type() string {
	clusterConfigs := e.tarmak.Config().Clusters(e.Name())

	if len(clusterConfigs) == 0 {
		return tarmakv1alpha1.EnvironmentTypeEmpty
	}

	for _, clusterConfig := range clusterConfigs {
		if clusterConfig.Name == "hub" {
			return tarmakv1alpha1.EnvironmentTypeMulti
		}
	}
	return tarmakv1alpha1.EnvironmentTypeSingle
}

func (e *Environment) Log() *logrus.Entry {
	return e.log
}

func (e *Environment) Validate() error {
	var result error

	err := e.Provider().Validate()
	if err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (e *Environment) BucketPrefix() string {
	stackState := e.hubCluster.Stack(tarmakv1alpha1.StackNameState)
	if stackState == nil {
		return ""
	}
	bucketPrefix, ok := stackState.Variables()["bucket_prefix"]
	if !ok {
		return ""
	}
	bucketPrefixString, ok := bucketPrefix.(string)
	if !ok {
		return ""
	}
	return bucketPrefixString
}

func (e *Environment) StateStack() interfaces.Stack {
	return e.hubCluster.Stack(tarmakv1alpha1.StackNameState)
}

func (e *Environment) VaultStack() interfaces.Stack {
	return e.hubCluster.Stack(tarmakv1alpha1.StackNameVault)
}

func (e *Environment) vaultRootTokenPath() string {
	return filepath.Join(e.ConfigPath(), "vault_root_token")
}

func (e *Environment) VaultRootToken() (string, error) {
	path := e.vaultRootTokenPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := utils.EnsureDirectory(filepath.Dir(path), 0700); err != nil {
			return "", fmt.Errorf("error creating directory: %s", err)
		}

		uuidValue := uuid.New()

		err := ioutil.WriteFile(path, []byte(fmt.Sprintf("%s\n", uuidValue.String())), 0600)
		if err != nil {
			return "", err
		}

		return uuidValue.String(), nil
	}

	uuidBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("unable to read vault root token %s: %s", path, err)
	}

	return strings.TrimSpace(string(uuidBytes)), nil
}

func (e *Environment) VaultTunnel() (interfaces.VaultTunnel, error) {
	stackVault := e.hubCluster.Stack(tarmakv1alpha1.StackNameVault)
	if stackVault == nil {
		return nil, errors.New("could not find vault stack")
	}
	vaultStack, ok := stackVault.(*stack.VaultStack)
	if !ok {
		return nil, fmt.Errorf("could not convert stack to VaultStack: %T", stackVault)
	}

	return vaultStack.VaultTunnel()
}

func (e *Environment) WingTunnel() interfaces.Tunnel {
	return e.Tarmak().SSH().Tunnel(
		"bastion",
		"localhost",
		9443,
	)
}

func (e *Environment) WingClientset() (*wingclient.Clientset, interfaces.Tunnel, error) {
	tunnel := e.WingTunnel()
	if err := tunnel.Start(); err != nil {
		return nil, nil, err
	}

	// TODO: Do proper TLS here
	restConfig := &rest.Config{
		Host: fmt.Sprintf("https://127.0.0.1:%d", tunnel.Port()),
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}

	clientset, err := wingclient.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}
	return clientset, tunnel, nil
}

func (e *Environment) Parameters() map[string]string {
	return map[string]string{
		"name":     e.Name(),
		"location": e.Location(),
		"provider": e.Provider().String(),
	}
}
