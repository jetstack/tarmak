package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/vault-helper/pkg/cert"
	"github.com/jetstack-experimental/vault-helper/pkg/kubeconfig"
)

// initCmd represents the init command
var kubeconfCmd = &cobra.Command{
	Use: "kubeconfig [cert role] [common name] [cert path] [kubeconfig path]",
	// TODO: Make short better
	Short: "Create local key to generate a CSR. Call vault with CSR for specified cert role. Write kubeconfig to yaml file.",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		if len(args) != 4 {
			log.Fatal("Wrong number of arguments given.\nUsage: vault-helper kubeconfig [cert role] [common name] [cert path] [kubeconfig path]")
		}

		abs, err := filepath.Abs(args[3])
		if err != nil {
			log.Fatalf("error generating absoute path from destination '%s': %v", args[3], err)
		}
		args = args[:3]

		i, err := newInstanceToken(cmd)
		if err != nil {
			i.Log.Fatal(err)
		}

		if err := i.TokenRenewRun(); err != nil {
			i.Log.Fatal(err)
		}
		c := cert.New(i.Log, i)
		if err := c.RunCert(); err != nil {
			c.Log.Fatal(err)
		}

		u := kubeconfig.New(log, c)
		u.SetFilePath(abs)

		if err := u.RunKube(); err != nil {
			u.Log.Fatal(err)
		}
	},
}

func init() {
	kubeconfCmd.PersistentFlags().Int(cert.FlagKeyBitSize, 2048, "Bit size used for generating key. [int]")
	kubeconfCmd.Flag(cert.FlagKeyBitSize).Shorthand = "b"

	kubeconfCmd.PersistentFlags().String(cert.FlagKeyType, "RSA", "Type of key to generate. [string]")
	kubeconfCmd.Flag(cert.FlagKeyType).Shorthand = "t"

	kubeconfCmd.PersistentFlags().StringSlice(cert.FlagIpSans, []string{}, "IP sans. [[]string] (default none)")
	kubeconfCmd.Flag(cert.FlagIpSans).Shorthand = "i"

	kubeconfCmd.PersistentFlags().StringSlice(cert.FlagSanHosts, []string{}, "Host Sans. [[]string] (default none)")
	kubeconfCmd.Flag(cert.FlagSanHosts).Shorthand = "s"

	kubeconfCmd.PersistentFlags().String(cert.FlagOwner, "", "Owner of created file/directories. Uid value also accepted. [string (default <current user>)")
	kubeconfCmd.Flag(cert.FlagOwner).Shorthand = "o"

	kubeconfCmd.PersistentFlags().String(cert.FlagGroup, "", "Group of created file/directories. Gid value also accepted. [string] (default <current user-group>)")
	kubeconfCmd.Flag(cert.FlagGroup).Shorthand = "g"

	instanceTokenFlags(kubeconfCmd)

	RootCmd.AddCommand(kubeconfCmd)
}
