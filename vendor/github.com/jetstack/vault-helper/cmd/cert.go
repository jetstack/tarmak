package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jetstack/vault-helper/pkg/cert"
)

// initCmd represents the init command
var certCmd = &cobra.Command{
	Use: "cert [cert role] [common name] [destination path]",
	// TODO: Make short better
	Short: "Create local key to generate a CSR. Call vault with CSR for specified cert role.",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		i, err := newInstanceToken(cmd)
		if err != nil {
			i.Log.Fatal(err)
		}

		if err := i.TokenRenewRun(); err != nil {
			i.Log.Fatal(err)
		}

		c := cert.New(log, i)
		if len(args) != 3 {
			i.Log.Fatal("wrong number of arguments given. Usage: vault-helper cert [cert role] [common name] [destination path]")
		}
		abs, err := filepath.Abs(args[2])
		if err != nil {
			i.Log.Fatalf("failed to generate absoute path from destination '%s': %v", args[2], err)
		}
		c.SetDestination(abs)

		c.SetRole(args[0])
		c.SetCommonName(args[1])

		if err := setFlagsCert(c, cmd); err != nil {
			log.Fatal(err)
		}

		if err := c.RunCert(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	certCmd.PersistentFlags().Int(cert.FlagKeyBitSize, 2048, "Bit size used for generating key. [int]")
	certCmd.Flag(cert.FlagKeyBitSize).Shorthand = "b"

	certCmd.PersistentFlags().String(cert.FlagKeyType, "RSA", "Type of key to generate. [string]")
	certCmd.Flag(cert.FlagKeyType).Shorthand = "t"

	certCmd.PersistentFlags().StringSlice(cert.FlagIpSans, []string{}, "IP sans. [[]string] (default none)")
	certCmd.Flag(cert.FlagIpSans).Shorthand = "i"

	certCmd.PersistentFlags().StringSlice(cert.FlagSanHosts, []string{}, "Host Sans. [[]string] (default none)")
	certCmd.Flag(cert.FlagSanHosts).Shorthand = "s"

	certCmd.PersistentFlags().String(cert.FlagOwner, "", "Owner of created file/directories. Uid value also accepted. [string] (default <current user>)")
	certCmd.Flag(cert.FlagOwner).Shorthand = "o"

	certCmd.PersistentFlags().String(cert.FlagGroup, "", "Group of created file/directories. Gid value also accepted. [string] (default <current user-group)")
	certCmd.Flag(cert.FlagGroup).Shorthand = "g"

	instanceTokenFlags(certCmd)

	RootCmd.AddCommand(certCmd)
}

func setFlagsCert(c *cert.Cert, cmd *cobra.Command) error {
	vInt, err := cmd.PersistentFlags().GetInt(cert.FlagKeyBitSize)
	if err != nil {
		return fmt.Errorf("error parsing %s [int] '%d': %v", cert.FlagKeyBitSize, vInt, err)
	}
	c.SetBitSize(vInt)

	vStr, err := cmd.PersistentFlags().GetString(cert.FlagKeyType)
	if err != nil {
		return fmt.Errorf("error parsing %s [string] '%s': %v", cert.FlagKeyType, vStr, err)
	}
	c.SetKeyType(vStr)

	vStr, err = cmd.PersistentFlags().GetString(cert.FlagOwner)
	if err != nil {
		return fmt.Errorf("error parsing %s [string] '%s': %v", cert.FlagOwner, vStr, err)
	}
	c.SetOwner(vStr)

	vStr, err = cmd.PersistentFlags().GetString(cert.FlagGroup)
	if err != nil {
		return fmt.Errorf("error parsing %s [string] '%s': %v", cert.FlagGroup, vStr, err)
	}
	c.SetGroup(vStr)

	vSli, err := cmd.PersistentFlags().GetStringSlice(cert.FlagIpSans)
	if err != nil {
		return fmt.Errorf("error parsing %s [[]string] '%s': %v", cert.FlagIpSans, vSli, err)
	}
	c.SetIPSans(vSli)

	vSli, err = cmd.PersistentFlags().GetStringSlice(cert.FlagSanHosts)
	if err != nil {
		return fmt.Errorf("error parsing %s [[]string] '%s': %v", cert.FlagSanHosts, vSli, err)
	}
	c.SetSanHosts(vSli)

	return nil
}
