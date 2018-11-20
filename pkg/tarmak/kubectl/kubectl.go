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
	tunnel interfaces.Tunnel
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

	if authInfo == nil {
		authInfo = api.NewAuthInfo()
	}

	authInfo.ClientKeyData = privateKeyPem
	authInfo.ClientCertificateData = []byte(certPem)
	cluster.CertificateAuthorityData = []byte(caPem)

	return nil
}

func (k *Kubectl) ensureWorkingKubeconfig(configPath string, publicAPIEndpoint bool) error {

	// attempt to load an existing config to use
	var c *api.Config
	if _, err := os.Stat(configPath); err == nil {
		conf, err := clientcmd.LoadFromFile(configPath)
		if err != nil {
			return err
		}
		c = conf
	}

	// If we are using a public endpoint then we don't need to set up a tunnel
	// but we need to keep a tunnel var around (in struct) so we can close it
	// later on if it is being used. Use k.stopTunnel() to ensure no panics.
	if !publicAPIEndpoint {
		k.tunnel = k.tarmak.Cluster().APITunnel()
		if err := k.tunnel.Start(); err != nil {
			k.stopTunnel()
			return err
		}
	}

	c, cluster, err := k.setupConfig(c, publicAPIEndpoint)
	if err != nil {
		return err
	}

	retries := 5
	for {
		k.log.Debugf("trying to connect to %s", cluster.Server)

		var version string
		version, err = k.verifyAPIVersion(*c)
		if err == nil {
			k.log.Debugf("connected to Kubernetes API %s", version)
			// break with err == nil with successful connection
			break
		}

		if strings.Contains(err.Error(), "certificate signed by unknown authority") {
			// TODO: this not really clean, if CA mismatched request new certificate
			err = k.requestNewAdminCert(cluster, c.AuthInfos[k.tarmak.Cluster().ClusterName()])
			if err != nil {
				break
			}

		} else {
			k.log.Warnf("error connecting to cluster: %s", err)
		}

		retries -= 1
		if retries == 0 {
			err = errors.New("unable to connect to kubernetes after 5 tries")
			break
		}

		if !publicAPIEndpoint {
			k.stopTunnel()
			k.tunnel = k.tarmak.Cluster().APITunnel()
			err = k.tunnel.Start()
			if err != nil {
				break
			}
		}

		// force a new config
		c, cluster, err = k.setupConfig(nil, publicAPIEndpoint)
		if err != nil {
			break
		}
	}

	// ensure we close the tunnel on error
	if err != nil {
		k.stopTunnel()
		return err
	}

	if err := utils.EnsureDirectory(filepath.Dir(configPath), 0700); err != nil {
		k.stopTunnel()
		return err
	}

	if err := clientcmd.WriteToFile(*c, configPath); err != nil {
		k.stopTunnel()
		return err
	}

	return nil
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

func (k *Kubectl) Kubectl(args []string, publicEndpoint bool) error {
	if k.tarmak.Cluster().Type() == clusterv1alpha1.ClusterTypeHub {
		return fmt.Errorf(
			"current cluster is of type %s so has no Kubernetes cluster: %s",
			clusterv1alpha1.ClusterTypeHub, k.tarmak.Cluster().Name())
	}

	err := k.ensureWorkingKubeconfig(k.ConfigPath(), publicEndpoint)
	if err != nil {
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

	err := k.ensureWorkingKubeconfig(path, publicAPIEndpoint)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("KUBECONFIG=%s", path), nil
}

func (k *Kubectl) setupConfig(c *api.Config, publicAPIEndpoint bool) (*api.Config, *api.Cluster, error) {
	if c == nil {
		c = api.NewConfig()
	}

	// cluster name in tarmak is cluster name in kubeconfig
	key := k.tarmak.Cluster().ClusterName()
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
		cluster.Server = ""
		cluster.CertificateAuthorityData = []byte{}
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
	if len(authInfo.ClientCertificateData) == 0 || len(authInfo.ClientKeyData) == 0 ||
		len(cluster.CertificateAuthorityData) == 0 {

		if err := k.tarmak.Terraform().Prepare(k.tarmak.Environment().Hub()); err != nil {
			return nil, nil, fmt.Errorf("failed to prepare terraform: %s", err)
		}

		if err := k.requestNewAdminCert(cluster, authInfo); err != nil {
			return nil, nil, err
		}
	}

	if publicAPIEndpoint {
		cluster.Server = fmt.Sprintf("https://api.%s-%s.%s",
			k.tarmak.Environment().Name(),
			k.tarmak.Cluster().Name(),
			k.tarmak.Provider().PublicZone())

	} else {
		if k.tunnel == nil {
			return nil, nil, fmt.Errorf("failed to get tunnel information at it is nil: %v", k.tunnel)
		}

		cluster.Server = fmt.Sprintf("https://%s:%d",
			k.tunnel.BindAddress(), k.tunnel.Port())
		k.log.Warnf("ssh tunnel connecting to Kubernetes API server will close after 10 minutes: %s",
			cluster.Server)
	}

	return c, cluster, nil
}

func (k *Kubectl) stopTunnel() {
	if k.tunnel != nil {
		k.tunnel.Stop()
	}
}
