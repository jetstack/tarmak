package tarmak

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/go-homedir"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

func (t *Tarmak) SSH(argsAdditional []string) {

	sess, err := t.Context().Environment().AWS.Session()
	if err != nil {
		t.log.Fatal(err)
	}
	svc := ec2.New(sess)

	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running"), aws.String("pending")},
		},
		&ec2.Filter{
			Name:   aws.String("tag:Environment"),
			Values: []*string{aws.String(t.Context().Environment().Name)},
		},
	}

	instances, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{Filters: filters})
	if err != nil {
		t.log.Fatal(err)
	}

	hosts := []*config.Host{}

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			if instance.PrivateIpAddress == nil || instance.InstanceId == nil {
				continue
			}

			host := &config.Host{
				ID:             *instance.InstanceId,
				Hostname:       *instance.PrivateIpAddress,
				HostnamePublic: false,
				User:           "centos",
			}
			if instance.PublicIpAddress != nil {
				host.Hostname = *instance.PublicIpAddress
				host.HostnamePublic = true
			}

			for _, tag := range instance.Tags {
				if *tag.Key == "tarmak_role" {
					host.Roles = strings.Split(*tag.Value, ",")
				}
			}

			hosts = append(hosts, host)
		}
	}

	hostsByRole := map[string][]*config.Host{}
	for _, host := range hosts {
		for _, role := range host.Roles {
			if _, ok := hostsByRole[role]; !ok {
				hostsByRole[role] = []*config.Host{host}
			} else {
				hostsByRole[role] = append(hostsByRole[role], host)
			}
			host.Aliases = append(host.Aliases, fmt.Sprintf("%s-%d", role, len(hostsByRole[role])))
		}
	}

	var sshConfig bytes.Buffer

	sshConfig.WriteString(fmt.Sprintf("# ssh config for tarmak context %s\n", t.Context().GetName()))

	// remove role-1 for single instances
	for role, hosts := range hostsByRole {
		if len(hosts) != 1 {
			continue
		}
		for pos, _ := range hosts[0].Aliases {
			if hosts[0].Aliases[pos] == fmt.Sprintf("%s-1", role) {
				hosts[0].Aliases[pos] = role
			}
		}
	}

	homeDir, err := homedir.Dir()
	if err != nil {
		t.log.Fatal(err)
	}

	baseDir := filepath.Join(homeDir, ".tarmak", t.Context().GetName())
	if err := utils.EnsureDirectory(baseDir, 0700); err != nil {
		t.log.Fatal(err)
	}

	sshConfigFile := filepath.Join(baseDir, "ssh_config")

	for _, host := range hosts {
		sshConfig.WriteString(fmt.Sprintf(`host %s
    User %s
    UserKnownHostsFile %s
    Hostname %s
    StrictHostKeyChecking no
    ServerAliveInterval 60
    IdentitiesOnly yes
    IdentityFile %s
`,
			strings.Join(append(host.Aliases, host.ID), " "),
			host.User,
			filepath.Join(baseDir, "ssh_known_hosts"),
			host.Hostname,
			t.Context().Environment().SSHKeyPath,
		))
		if !host.HostnamePublic {
			sshConfig.WriteString(fmt.Sprintf(
				"    ProxyCommand ssh -F %s -W %%h:%%p bastion\n\n",
				sshConfigFile,
			))
		}
	}

	err = ioutil.WriteFile(sshConfigFile, sshConfig.Bytes(), 0600)
	if err != nil {
		t.log.Fatal(err)
	}

	args := []string{
		"ssh",
		"-F",
		sshConfigFile,
	}
	args = append(args, argsAdditional...)

	cmd := exec.Command(args[0], args[1:len(args)]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err = cmd.Start()
	if err != nil {
		t.log.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		t.log.Fatal(err)
	}

}
