// Copyright Jetstack Ltd. See LICENSE for details.
package terraform

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	multierror "github.com/hashicorp/go-multierror"
	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/cluster"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/provider/amazon"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

func (t *Terraform) GenerateCode(c interfaces.Cluster) (err error) {

	terraformCodePath := t.codePath(c)
	if err := utils.EnsureDirectory(
		terraformCodePath,
		0700,
	); err != nil {
		return err
	}
	if err := os.Chmod(terraformCodePath, 0700); err != nil {
		return err
	}

	// remove existing modules, create new directory and copy static files in
	rootPath, err := t.tarmak.RootPath()
	if err != nil {
		return err
	}
	sourceModulesPath := filepath.Join(
		rootPath,
		"terraform",
		c.Environment().Provider().Cloud(),
		"modules",
	)
	destModulesPath := filepath.Join(
		terraformCodePath,
		"modules",
	)
	if err := os.RemoveAll(destModulesPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := utils.CopyDir(sourceModulesPath, destModulesPath); err != nil {
		return err
	}

	// move in wing binary for terraform bucket object
	sourceWingBinary, err := os.Open(filepath.Join(rootPath, "wing_linux_amd64"))
	if err != nil {
		return err
	}
	defer sourceWingBinary.Close()

	destWingBinary, err := os.Create(filepath.Join(terraformCodePath, "wing_linux_amd64"))
	if err != nil {
		return err
	}
	defer destWingBinary.Close()

	_, err = io.Copy(destWingBinary, sourceWingBinary)
	if err != nil {
		return err
	}

	// create puppet.tar.gz
	puppetTarGzFilename := filepath.Clean(
		filepath.Join(
			terraformCodePath,
			"puppet.tar.gz",
		),
	)
	file, err := os.OpenFile(
		puppetTarGzFilename,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return fmt.Errorf("error creating %s: %s", puppetTarGzFilename, err)
	}
	if err = t.tarmak.Cluster().Environment().Tarmak().Puppet().TarGz(file); err != nil {
		return fmt.Errorf("error writing to %s: %s", puppetTarGzFilename, err)
	}

	// generate templates
	templ := &terraformTemplate{
		cluster:  c,
		destDir:  terraformCodePath,
		rootPath: rootPath,
	}
	if err := templ.Generate(); err != nil {
		return err
	}

	return nil

}

type terraformTemplate struct {
	cluster  interfaces.Cluster
	destDir  string
	rootPath string
}

func (t *terraformTemplate) Generate() error {

	var result error
	if err := t.generateRemoteStateConfig(); err != nil {
		result = multierror.Append(result, err)
	}
	for _, module := range []string{"state", "bastion", "network", "network-existing-vpc", "jenkins", "vault", "kubernetes"} {
		if err := t.generateModuleInstanceTemplates(module); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if err := t.generateTemplate("modules", "modules"); err != nil {
		result = multierror.Append(result, err)
	}
	if err := t.generateTemplate("inputs", "inputs"); err != nil {
		result = multierror.Append(result, err)
	}
	if err := t.generateTemplate("outputs", "outputs"); err != nil {
		result = multierror.Append(result, err)
	}
	if err := t.generateTemplate("providers", "providers"); err != nil {
		result = multierror.Append(result, err)
	}
	if err := t.generateTemplate("jenkins_elb", "modules/jenkins/jenkins_elb"); err != nil {
		result = multierror.Append(result, err)
	}
	if err := t.generateTerraformVariables(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (t *terraformTemplate) data(module string) map[string]interface{} {

	_, existingVPC := t.cluster.Config().Network.ObjectMeta.Annotations[clusterv1alpha1.ExistingVPCAnnotationKey]

	jenkinsCertificateARN := ""
	jenkinsInstall := false
	for _, instancePool := range t.cluster.InstancePools() {
		if instancePool.Role().Name() == clusterv1alpha1.InstancePoolTypeJenkins {
			jenkinsInstall = true
			jenkinsCertificateARN, _ = instancePool.Config().Annotations[cluster.JenkinsCertificateARNAnnotationKey]
			break
		}
	}

	return map[string]interface{}{
		"ClusterTypeClusterSingle": clusterv1alpha1.ClusterTypeClusterSingle,
		"ClusterTypeHub":           clusterv1alpha1.ClusterTypeHub,
		"ClusterTypeClusterMulti":  clusterv1alpha1.ClusterTypeClusterMulti,
		"ClusterType":              t.cluster.Type(),
		"InstancePools":            t.cluster.InstancePools(),
		"ExistingVPC":              existingVPC,
		// cluster.Roles() returns a list of roles based off of the types of instancePools in tarmak.yaml
		"Roles":                 t.cluster.Roles(),
		"SocketPath":            tarmakSocketPath(t.cluster.ConfigPath()),
		"JenkinsCertificateARN": jenkinsCertificateARN,
		"JenkinsInstall":        jenkinsInstall,
		"Module":                module,
	}
}

func (t *terraformTemplate) funcs() template.FuncMap {
	templatesFuncs := sprig.TxtFuncMap()
	templatesFuncs["CIDRToString"] = func(i *net.IPNet) string { return i.String() }
	templatesFuncs["stringFromPointer"] = func(i *string) string { return *i }
	templatesFuncs["dict"] = func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	}
	return templatesFuncs
}

// generate single file templates
func (t *terraformTemplate) generateTemplate(name string, target string) error {
	templateFile := filepath.Clean(
		filepath.Join(
			t.rootPath,
			"terraform",
			t.cluster.Environment().Provider().Cloud(),
			fmt.Sprintf("templates/%s.tf.template", name),
		),
	)

	templatesParsed, err := template.New(name).Funcs(t.funcs()).ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template '%s'", name)
	}

	mainTemplate := templatesParsed.Lookup(fmt.Sprintf("%s.tf.template", name))

	file, err := os.OpenFile(
		filepath.Join(
			t.destDir,
			fmt.Sprintf("%s.tf", target),
		),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := mainTemplate.Execute(
		file,
		// TODO: change behaviour of data function to not have to use module kubernetes below
		t.data("kubernetes"),
	); err != nil {
		return fmt.Errorf("failed to execute template '%s'", name)
	}

	return nil
}

func (t *terraformTemplate) generateModuleInstanceTemplates(module string) error {
	data := t.data(module)
	// generate instance pools security group rules
	if len(t.cluster.InstancePools()) > 0 {
		awsSGRules, err := t.generateAWSSecurityGroup()
		if err != nil {
			return err
		}

		data["AWSSGRules"] = awsSGRules
	}

	templateFile := filepath.Clean(
		filepath.Join(
			t.rootPath,
			"terraform",
			t.cluster.Environment().Provider().Cloud(),
			"templates/instance_pools/*.tf.template",
		),
	)
	name := "instance_pools"

	templatesParsed, err := template.New(name).Funcs(t.funcs()).ParseGlob(templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template '%s': %s", name, err)
	}

	mainTemplate := templatesParsed.Lookup(fmt.Sprintf("%s.tf.template", name))

	file, err := os.OpenFile(
		filepath.Join(
			t.destDir,
			fmt.Sprintf("modules/%s/%s.tf", module, name),
		),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := mainTemplate.Execute(
		file,
		data,
	); err != nil {
		return fmt.Errorf("failed to execute template '%s': %s", name, err)
	}

	return nil
}

// TODO: move this to the cloud provider
func (t *terraformTemplate) generateAWSSecurityGroup() (rules map[string][]*amazon.AWSSGRule, err error) {
	rules = make(map[string][]*amazon.AWSSGRule)
	for _, role := range t.cluster.Roles() {

		if role.Name() == "vault" || role.Name() == "bastion" {
			continue
		}

		roleRules, err := amazon.GenerateAWSRules(role)
		if err != nil {
			return nil, err
		}
		rules[role.Name()] = roleRules
	}

	return rules, nil
}

func (t *terraformTemplate) generateTerraformVariables() error {
	// merge maps overwrite less specific configs
	terraformVars := utils.MergeMaps(
		t.cluster.Environment().Tarmak().Variables(),
		t.cluster.Environment().Variables(),
		t.cluster.Variables(),
	)

	// generate a tfvar file from map
	terraformVarsFile, err := MapToTerraformTfvars(terraformVars)
	if err != nil {
		return err
	}

	// write to file
	file, err := os.OpenFile(
		filepath.Join(
			t.destDir,
			"terraform.tfvars",
		),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write([]byte(terraformVarsFile)); err != nil {
		return err
	}

	return nil
}

func (t *terraformTemplate) generateRemoteStateConfig() error {
	// write to file
	file, err := os.OpenFile(
		filepath.Join(
			t.destDir,
			"terraform_remote_state.tf",
		),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write([]byte(t.cluster.RemoteState())); err != nil {
		return err
	}

	return nil

}
