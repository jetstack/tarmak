// Copyright Jetstack Ltd. See LICENSE for details.
package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"

	tarmakDocker "github.com/jetstack/tarmak/pkg/docker"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type TerraformContainer struct {
	*tarmakDocker.AppContainer
	t     *Terraform
	stack interfaces.Stack
	log   *log.Entry
}

type terraformOutputValue struct {
	Sensitive bool        `json:"sensitive,omitifempty"`
	Type      string      `json:"type,omitifempty"`
	Value     interface{} `value:"type,omitifempty"`
}

func MapToTerraformTfvars(input map[string]interface{}) (output string, err error) {
	var buf bytes.Buffer

	for key, value := range input {
		switch v := value.(type) {
		case map[string]string:
			_, err := buf.WriteString(fmt.Sprintf("%s = {\n", key))
			if err != nil {
				return "", err
			}

			keys := make([]string, len(v))
			pos := 0
			for key, _ := range v {
				keys[pos] = key
				pos++
			}
			sort.Strings(keys)
			for _, key := range keys {
				_, err := buf.WriteString(fmt.Sprintf("  %s = \"%s\"\n", key, v[key]))
				if err != nil {
					return "", err
				}
			}

			_, err = buf.WriteString("}\n")
			if err != nil {
				return "", err
			}
		case []string:
			values := make([]string, len(v))
			for pos, _ := range v {
				values[pos] = fmt.Sprintf(`"%s"`, v[pos])
			}
			_, err := buf.WriteString(fmt.Sprintf("%s = [%s]\n", key, strings.Join(values, ", ")))
			if err != nil {
				return "", err
			}
		case string:
			_, err := buf.WriteString(fmt.Sprintf("%s = \"%s\"\n", key, v))
			if err != nil {
				return "", err
			}
		case *net.IPNet:
			_, err := buf.WriteString(fmt.Sprintf("%s = \"%s\"\n", key, v.String()))
			if err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("ignoring unknown var key='%s' type='%#+v'", key, v)
		}
	}
	return buf.String(), nil
}

func (tc *TerraformContainer) Shell() (err error) {
	if err := tc.writeTfvars(); err != nil {
		return err
	}

	_, err = tc.Attach("/bin/sh", []string{})
	return err
}

func (tc *TerraformContainer) writeTfvars() (err error) {
	// adds parameters as CLI args
	terraformVars := utils.MergeMaps(
		tc.stack.Cluster().Environment().Tarmak().Variables(),
		tc.stack.Cluster().Environment().Variables(),
		tc.stack.Cluster().Variables(),
		tc.stack.Variables(),
	)
	tc.log.WithFields(terraformVars).Debug("terraform vars generated")

	terraformVarsFile, err := MapToTerraformTfvars(terraformVars)
	if err != nil {
		return err
	}

	remoteStateTar, err := tarmakDocker.TarStreamFromFile("terraform.tfvars", terraformVarsFile)
	if err != nil {
		return err
	}

	err = tc.UploadToContainer(remoteStateTar, "/terraform")
	if err != nil {
		return err
	}

	return nil
}

func (tc *TerraformContainer) Plan(additionalArgs []string, destroy bool) (changesNeeded bool, err error) {

	args := []string{"plan", "-out=terraform.plan", "-detailed-exitcode", "-input=false"}

	if destroy {
		args = append(args, "-destroy")
	}

	if err := tc.writeTfvars(); err != nil {
		return false, err
	}

	args = append(args, additionalArgs...)

	returnCode, err := tc.Execute("terraform", args)
	if err != nil {
		return false, err
	}

	if returnCode == 0 {
		return false, nil
	}
	if returnCode == 2 {
		return true, nil
	}
	return false, fmt.Errorf("unexpected return code: exp=0/2, act=%d", returnCode)
}

func (tc *TerraformContainer) Apply() error {
	returnCode, err := tc.Execute("terraform", []string{"apply", "-input=false", "terraform.plan"})
	if err != nil {
		return err
	}
	if exp, act := 0, returnCode; exp != act {
		return fmt.Errorf("unexpected return code: exp=%d, act=%d", exp, act)
	}
	return nil
}

func (tc *TerraformContainer) Init() error {
	returnCode, err := tc.Execute("terraform", []string{"init", "-input=false", "-get-plugins=false"})

	if err != nil {
		return err
	}
	if exp, act := 0, returnCode; exp != act {
		return fmt.Errorf("unexpected return code: exp=%d, act=%d", exp, act)
	}
	return nil
}

func (tc *TerraformContainer) InitForceCopy() error {
	returnCode, err := tc.Execute("terraform", []string{"init", "-force-copy", "-input=false", "-get-plugins=false"})
	if err != nil {
		return err
	}
	if exp, act := 0, returnCode; exp != act {
		return fmt.Errorf("unexpected return code: exp=%d, act=%d", exp, act)
	}
	return nil
}

func (tc *TerraformContainer) Output() (map[string]interface{}, error) {
	stdOut, stdErr, returnCode, err := tc.Capture("terraform", []string{"output", "-json"})
	if err != nil {
		return nil, err
	}
	if exp, act := 0, returnCode; exp != act {
		return nil, fmt.Errorf("unexpected return code: exp=%d, act=%d: %s", exp, act, stdErr)
	}

	var values map[string]terraformOutputValue
	if err := json.Unmarshal([]byte(stdOut), &values); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %s", err)
	}

	variables := make(map[string]interface{})
	for key, value := range values {
		variables[key] = value.Value
	}

	return variables, nil
}

func (tc *TerraformContainer) CopyRemoteState(content string) error {
	remoteStateTar, err := tarmakDocker.TarStreamFromFile("terraform_remote_state.tf", content)
	if err != nil {
		return err
	}

	err = tc.UploadToContainer(remoteStateTar, "/terraform")
	if err != nil {
		return err
	}
	tc.log.Debug("copied remote state config into container")

	return nil
}

func (tc *TerraformContainer) prepare() error {
	// validate environment
	err := tc.t.tarmak.Cluster().Environment().Validate()
	if err != nil {
		return fmt.Errorf("error validating environment: %s", err)
	}

	// rootDir
	rootPath, err := tc.t.tarmak.RootPath()
	if err != nil {
		return fmt.Errorf("error getting rootPath: %s", err)
	}

	// get aws secrets
	if environmentProvider, err := tc.t.tarmak.Cluster().Environment().Provider().Environment(); err != nil {
		return fmt.Errorf("error getting environment secrets from provider: %s", err)
	} else {
		tc.Env = append(tc.Env, environmentProvider...)
	}

	// set default commandpfals
	tc.Cmd = []string{"sleep", "3600"}
	tc.WorkingDir = "/terraform"
	tc.Keep = tc.t.tarmak.KeepContainers()

	// build terraform image if needed
	tc.log.Debug("prepare container")

	err = tc.AppContainer.Prepare()
	if err != nil {
		return err
	}

	// tar terraform manifests
	tarOpts := &archive.TarOptions{
		Compression:  archive.Uncompressed,
		NoLchown:     true,
		IncludeFiles: []string{"."},
	}

	terraformDir := filepath.Clean(filepath.Join(rootPath, "terraform", tc.t.tarmak.Cluster().Environment().Provider().Cloud(), tc.stack.Name()))
	tc.log = tc.log.WithField("terraform-dir", terraformDir)

	terraformDirInfo, err := os.Stat(terraformDir)
	if err != nil {
		return err
	}
	if !terraformDirInfo.IsDir() {
		return fmt.Errorf("path '%s' is not a directory", terraformDir)
	}

	terraformTar, err := archive.TarWithOptions(terraformDir, tarOpts)
	if err != nil {
		return err
	}

	err = tc.UploadToContainer(terraformTar, "/terraform")
	if err != nil {
		return err
	}
	tc.log.Debug("copied terraform manifests into container")

	// if instance pools exist, execute template
	instancePools := tc.stack.InstancePools()
	if len(instancePools) > 0 {
		tc.log.Debug("generating instance pools templates")
		templatesGlob := filepath.Clean(filepath.Join(rootPath, "terraform", tc.t.tarmak.Cluster().Environment().Provider().Cloud(), "templates/instance_pools/*.tf.template"))
		templates := template.Must(template.New("instance_pools").Funcs(sprig.TxtFuncMap()).ParseGlob(templatesGlob))

		baseTemplate := "instance_pool.tf.template"
		tpl := templates.Lookup(baseTemplate)

		buf := new(bytes.Buffer)

		if err := tpl.Execute(
			buf,
			map[string]interface{}{
				"InstancePools": instancePools,
				"Roles":         tc.stack.Roles(),
				"Stack":         tc.stack.Name(),
			},
		); err != nil {
			return err
		}

		templateTfTar, err := tarmakDocker.TarStreamFromFile("node_groups.tf", buf.String())
		if err != nil {
			return err
		}

		err = tc.UploadToContainer(templateTfTar, "/terraform")
		if err != nil {
			return err
		}
	}

	err = tc.Start()
	if err != nil {
		return fmt.Errorf("error starting container: %s", err)
	}

	/*err = rpc.Start()
	if err != nil {
		return fmt.Errorf("error starting RPC server: %s", err)
	}*/

	return nil
}
