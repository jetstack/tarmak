// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"context"
	"fmt"
	"strings"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

const FlagTerraformStacks = "terraform-stacks"
const FlagForceDestroyStateStack = "force-destroy-state"

func (t *Tarmak) Terraform() interfaces.Terraform {
	return t.terraform
}

func (t *Tarmak) CmdTerraformApply(args []string, ctx context.Context) error {
	selectStacks := t.flags.Cluster.Apply.InfrastructureStacks

	stacks := t.Cluster().Stacks()
	for _, stack := range stacks {

		if len(selectStacks) > 0 {
			found := false
			for _, selectStack := range selectStacks {
				if selectStack == stack.Name() {
					found = true
				}
			}
			if !found {
				continue
			}
		}

		stack.Log().Info("running apply")
		err := t.terraform.Apply(stack, args, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Tarmak) CmdTerraformDestroy(args []string, ctx context.Context) error {
	selectStacks := t.flags.Cluster.Destroy.InfrastructureStacks
	forceDestroyStateStack := t.flags.Cluster.Destroy.ForceDestroyStateStack

	stacks := t.Cluster().Stacks()
	for posStack, _ := range stacks {
		stack := stacks[len(stacks)-posStack-1]
		if !forceDestroyStateStack && stack.Name() == tarmakv1alpha1.StackNameState {
			t.log.Debugf("ignoring stack '%s'", stack.Name())
			continue
		}

		if len(selectStacks) > 0 {
			found := false
			for _, selectStack := range selectStacks {
				if selectStack == stack.Name() {
					found = true
				}
			}
			if !found {
				continue
			}
		}

		stack.Log().Info("running destroy")
		err := t.terraform.Destroy(stack, args, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Tarmak) CmdTerraformShell(args []string) error {
	paramStackName := ""
	if len(args) > 0 {
		paramStackName = strings.ToLower(args[0])
	}

	// find matching stacks
	stacks := t.Cluster().Stacks()
	stackNames := make([]string, len(stacks))
	for pos, stack := range stacks {
		stackNames[pos] = stack.Name()
		if stack.Name() == paramStackName {
			return t.terraform.Shell(stack, args)
		}
	}

	return fmt.Errorf("you have to provide exactly one parameter that contains one of the stack names %s", strings.Join(stackNames, ", "))
}
