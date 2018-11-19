// Copyright Jetstack Ltd. See LICENSE for details.
package kubectl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var _ interfaces.Kubectl = &Kubectl{}

type Kubectl struct {
	tarmak interfaces.Tarmak
	log    *logrus.Entry
}

func New(tarmak interfaces.Tarmak) *Kubectl {
	k := &Kubectl{
		tarmak: tarmak,
		log:    tarmak.Log(),
	}

	return k
}

func (k *Kubectl) ConfigPath() string {
	return filepath.Join(k.tarmak.Cluster().ConfigPath(), "kubeconfig")
}

func (k *Kubectl) requestNewAdminCert(cluster *api.Cluster, authInfo *api.AuthInfo) error {
	path := fmt.Sprintf("%s/pki/k8s/sign/admin", k.tarmak.Cluster().ClusterName())

	k.log.Infof("request new certificate from vault (%s)", path)

	if err := k.tarmak.Cluster().Environment().Validate(); err != nil {
		k.log.Fatal("could not validate config: ", err)
	}

	vault := k.tarmak.Environment().Vault()

	// read vault root token
	vaultRootToken, err := vault.RootToken()
	if err != nil {
		return err
	}

	// get kubernetes outputs
	outputs, err := k.tarmak.Environment().Hub().TerraformOutput()
	if err != nil {
		return err
	}

	interfaceInstanceFQDNs := outputs["instance_fqdns"].([]interface{})
	instanceFQDNs := make([]string, len(interfaceInstanceFQDNs))
	for i := range interfaceInstanceFQDNs {
		instanceFQDNs[i] = interfaceInstanceFQDNs[i].(string)
	}

	vaultTunnel, err := vault.TunnelFromFQDNs(instanceFQDNs, outputs["vault_ca"].(string))
	if err != nil {
		return err
	}
	defer vaultTunnel.Stop()

	v := vaultTunnel.VaultClient()
	v.SetToken(vaultRootToken)

	// generate new RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("unable to generate private key: %s", err)
	}
	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// define CSR template
	var csrTemplate = x509.CertificateRequest{
		Subject:            pkix.Name{CommonName: "admin"},
		SignatureAlgorithm: x509.SHA512WithRSA,
	}

	// generate the CSR request
	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		return err
	}

	// pem encode CSR
	csrPem := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csr,
	})

	inputData := map[string]interface{}{
		"csr":         string(csrPem),
		"common_name": "admin",
	}

	output, err := v.Logical().Write(path, inputData)
	if err != nil {
		return err
	}

	certPemIntf, ok := output.Data["certificate"]
	if !ok {
		return errors.New("key certificate not found")
	}

	certPem, ok := certPemIntf.(string)
	if !ok {
		return fmt.Errorf("certificate has unexpected type %s", certPemIntf)
	}

	caPemIntf, ok := output.Data["issuing_ca"]
	if !ok {
		return errors.New("issuing_ca not found")
	}

	caPem, ok := caPemIntf.(string)
	if !ok {
		return fmt.Errorf("issuing_ca has unexpected type %s", caPemIntf)
	}

	authInfo.ClientKeyData = privateKeyPem
	authInfo.ClientCertificateData = []byte(certPem)
	cluster.CertificateAuthorityData = []byte(caPem)

	return nil
}

func (k *Kubectl) ensureWorkingKubeconfig(configPath string, publicAPIEndpoint bool) (interfaces.Tunnel, error) {
	c := api.NewConfig()

	// cluster name in tarmak is cluster name in kubeconfig
	key := k.tarmak.Cluster().ClusterName()

	// load an existing config
	if _, err := os.Stat(configPath); err == nil {
		conf, err := clientcmd.LoadFromFile(configPath)
		if err != nil {
			return nil, err
		}
		c = conf
	}

	c.CurrentContext = key

	ctx, ok := c.Contexts[key]
	if !ok {
		ctx = api.NewContext()
		ctx.Namespace = "kube-system"
		ctx.Cluster = key
		ctx.AuthInfo = key
		c.Contexts[key] = ctx
	}

	cluster, ok := c.Clusters[key]
	if !ok {
		cluster = api.NewCluster()
		cluster.CertificateAuthorityData = []byte{}
		cluster.Server = ""
		c.Clusters[key] = cluster
	}

	authInfo, ok := c.AuthInfos[key]
	if !ok {
		authInfo = api.NewAuthInfo()
		authInfo.ClientCertificateData = []byte{}
		authInfo.ClientKeyData = []byte{}
		c.AuthInfos[key] = authInfo
	}

	// check if certificates are set
	if len(authInfo.ClientCertificateData) == 0 || len(authInfo.ClientKeyData) == 0 || len(cluster.CertificateAuthorityData) == 0 {

		if err := k.tarmak.Terraform().Prepare(k.tarmak.Environment().Hub()); err != nil {
			return nil, fmt.Errorf("failed to prepare terraform: %s", err)
		}

		if err := k.requestNewAdminCert(cluster, authInfo); err != nil {
			return nil, err
		}
	}

	retries := 5
	tunnel := k.tarmak.Cluster().APITunnel()
	if err := tunnel.Start(); err != nil {
		return tunnel, err
	}

	if publicAPIEndpoint {
		cluster.Server = fmt.Sprintf("https://api.%s-%s.%s",
			k.tarmak.Environment().Name(),
			k.tarmak.Cluster().Name(),
			k.tarmak.Provider().PublicZone())
	} else {
		cluster.Server = fmt.Sprintf("https://%s:%d",
			tunnel.BindAddress(), tunnel.Port())
		k.log.Warnf("ssh tunnel connecting to Kubernetes API server will close after 10 minutes: %s",
			cluster.Server)
	}

	for {

		k.log.Debugf("trying to connect to %+v", cluster.Server)

		version, err := k.verifyAPIVersion(*c)
		if err == nil {
			k.log.Debugf("connected to Kubernetes API %s", version)
			break
		} else if strings.Contains(err.Error(), "certificate signed by unknown authority") {
			// TODO: this not really clean, if CA mismatched request new certificate
			if err := k.requestNewAdminCert(cluster, authInfo); err != nil {
				return tunnel, err
			}
		} else {
			k.log.Warnf("error connecting to cluster: %s", err)
		}

		retries -= 1
		if retries == 0 {
			return tunnel, errors.New("unable to connect to kubernetes after 5 tries")
		}

		tunnel.Stop()
		tunnel = k.tarmak.Cluster().APITunnel()
		if err := tunnel.Start(); err != nil {
			return tunnel, err
		}
	}

	if err := utils.EnsureDirectory(filepath.Dir(configPath), 0700); err != nil {
		return tunnel, err
	}

	if err := clientcmd.WriteToFile(*c, configPath); err != nil {
		return tunnel, err
	}

	return tunnel, nil

}

func (k *Kubectl) verifyAPIVersion(c api.Config) (version string, err error) {
	clientConfig := clientcmd.NewDefaultClientConfig(c, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return "", err
	}

	// test connectivity
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return "", err
	}

	versionInfo, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}

	return versionInfo.String(), nil
}

func (k *Kubectl) Kubectl(args []string) error {
	if k.tarmak.Cluster().Type() == clusterv1alpha1.ClusterTypeHub {
		currentCluster, err := k.tarmak.Config().CurrentCluster()
		if err != nil {
			return fmt.Errorf("error retrieving current cluster: %s", err)
		}
		return fmt.Errorf("the current cluster '%s' is a hub and therefore does not contain a Kubernetes cluster", currentCluster)
	}

	tunnel, err := k.ensureWorkingKubeconfig(k.ConfigPath(), false)
	if err != nil {
		if tunnel != nil {
			tunnel.Stop()
		}
		return err
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("KUBECONFIG=%s", k.ConfigPath()),
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err = cmd.Start()
	if err != nil {
		return err
	}

	cmd.Wait()

	return nil
}

func (k *Kubectl) Kubeconfig(path string, publicAPIEndpoint bool) (string, error) {
	if k.tarmak.Cluster().Type() == clusterv1alpha1.ClusterTypeHub {
		return "", fmt.Errorf(
			"current cluster is of type %s so has no Kubernetes cluster: %s",
			clusterv1alpha1.ClusterTypeHub, k.tarmak.Cluster().Name())
	}

	tunnel, err := k.ensureWorkingKubeconfig(path, publicAPIEndpoint)
	if err != nil {
		if tunnel != nil {
			tunnel.Stop()
		}

		return "", err
	}

	return fmt.Sprintf("KUBECONFIG=%s", path), nil
}
