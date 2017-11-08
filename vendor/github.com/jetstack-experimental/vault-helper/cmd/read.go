package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/vault-helper/pkg/read"
)

// initCmd represents the init command
var readCmd = &cobra.Command{
	Use:   "read [vault path]",
	Short: "Read arbitrary vault path. If no output file specified, output to console.",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		i, err := newInstanceToken(cmd)
		if err != nil {
			i.Log.Fatal(err)
		}

		if err := i.TokenRenewRun(); err != nil {
			i.Log.Fatal(err)
		}

		r := read.New(log, i)
		if len(args) != 1 {
			log.Fatal("incorrect number of arguments given. Usage: vault-helper read [vault path] [flags]")
		}
		r.SetVaultPath(args[0])

		if err := setFlagsRead(r, cmd); err != nil {
			log.Fatal(err)
		}

		if err := r.RunRead(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	instanceTokenFlags(readCmd)

	readCmd.PersistentFlags().String(read.FlagOutputPath, "", "Set destination file path of read responce. Output to console if no filepath given (default <console>)")
	readCmd.Flag(read.FlagOutputPath).Shorthand = "d"
	readCmd.PersistentFlags().String(read.FlagField, "", "If included, the raw value of the specified field will be output. If not, output entire responce in JSON (default <all>)")
	readCmd.Flag(read.FlagField).Shorthand = "f"
	readCmd.PersistentFlags().String(read.FlagOwner, "", "Set owner of output file. Uid value also accepted. (default <current user>)")
	readCmd.Flag(read.FlagOwner).Shorthand = "o"
	readCmd.PersistentFlags().String(read.FlagGroup, "", "Set group of output file. Gid value also accepted. (default <current user-group>)")
	readCmd.Flag(read.FlagGroup).Shorthand = "g"

	RootCmd.AddCommand(readCmd)
}

func setFlagsRead(r *read.Read, cmd *cobra.Command) error {
	value, err := cmd.PersistentFlags().GetString(read.FlagOutputPath)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %v", read.FlagOutputPath, value, err)
	}
	if value != "" {
		abs, err := filepath.Abs(value)
		if err != nil {
			return fmt.Errorf("error generating absoute path from destination '%s': %v", value, err)
		}
		r.SetFilePath(abs)
	}

	value, err = cmd.PersistentFlags().GetString(read.FlagField)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %v", read.FlagField, value, err)
	}
	if value != "" {
		r.SetFieldName(value)
	}

	value, err = cmd.PersistentFlags().GetString(read.FlagOwner)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %v", read.FlagOwner, value, err)
	}
	if value != "" {
		r.SetOwner(value)
	}

	value, err = cmd.PersistentFlags().GetString(read.FlagGroup)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %v", read.FlagGroup, value, err)
	}
	if value != "" {
		r.SetGroup(value)
	}

	return nil
}
