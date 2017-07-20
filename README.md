# Tarmak

```
Tarmak is a Toolkit to spin up kubernetes clusters

Usage:
  tarmak [command]

Available Commands:
  help              Help about any command
  image-build       This builds an image for an environment using packer
  init              init a cluster configuration
  list              list nodes of the context
  puppet-dist       Build a puppet.tar.gz
  ssh               ssh into instance
  terraform-apply   This applies the set of stacks in the current context
  terraform-destroy This applies the set of stacks in the current context
  version           Print the version number of tarmak

Flags:
      --config string   config file (default is $HOME/.tarmak.yaml)
  -h, --help            help for tarmak
  -t, --toggle          Help message for toggle

Use "tarmak [command] --help" for more information about a command.
```
