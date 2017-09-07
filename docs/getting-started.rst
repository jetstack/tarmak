.. getting-started:

Getting Started
================

Here we will walk through how to use Tarmak to spin up a Kubernetes cluster in AWS. This will deploy the Kubernetes master nodes, an etcd cluster, worker nodes, vault and a bastion node with a public IP address.

Prerequisites
-------------

* An AWS account
* Go 
* Python 2.x 
* Docker 
* A domain delegated to AWS
* Vault with an AWS secret backend configured (optional)

Steps
-----

First we will clone the Tarmak repository and build the `tarmak` binary. Make sure you have your `$GOPATH` set correctly. The last line may change depending on your architecture.

:: 

  mkdir -p $GOPATH/src/github.com/jetstack
  cd $GOPATH/src/github.com/jetstack
  git clone git@gitlab.jetstack.net:tarmak/tarmak.git
  cd tarmak
  make build
  ln -s $PWD/tarmak_darwin_amd64 /usr/local/bin/tarmak

You should now be able to run `tarmak` to view the available commands.

::

  % tarmak
  Tarmak is a Toolkit to spin up kubernetes clusters

  Usage:
    tarmak [command]

  Available Commands:
    help              Help about any command
    image-build       This builds an image for an environment using packer
    init              init a cluster configuration
    kubectl           kubectl against the current cluster
    list              list nodes of the context
    puppet-dist       Build a puppet.tar.gz
    ssh               ssh into instance
    terraform-apply   This applies the set of stacks in the current context
    terraform-destroy This applies the set of stacks in the current context
    terraform-shell   This prepare a terraform container and executes a shell in this context
    version           Print the version number of tarmak

  Flags:
        --config string   config file (default is $HOME/.tarmak.yaml)
    -h, --help            help for tarmak
    -t, --toggle          Help message for toggle

  Use "tarmak [command] --help" for more information about a command.

The plan will be to first initialise our cluster configuration for our environment, then build an image for our environment and then finally apply our context to AWS. Run `tarmak init` to initialise out configuration.

:: 

  % tarmak init
  What should be the name of the cluster?

  The name consists of two parts seperated by a dash. First part is the environment name, second part the cluster name. Both names should be matching [a-z0-9]+

  Enter a value: dev-cluster

  Do you want to use vault to get credentials for AWS? [Y/N] 
  Enter a value (Default is N): Y 

  Which path should be used for AWS credentials?
  Enter a value (Default is jetstack/aws/jetstack-dev/sts/admin): jetstack/aws/jetstack-dev/sts/admin

  Which region should be used?
  Enter a value (Default is eu-west-1): eu-west-1

  What bucket prefix should be used?
  Enter a value (Default is tarmak-): tarmak-

  What public zone should be used?

  Please make sure you can delegate this zone to AWS!

  Enter a value: jetstack.io

  What private zone should be used?
  Enter a value (Default is tarmak.local): tarmak.local

  What is the mail address of someone responsible?
  Enter a value: luke.addison@jetstack.io

  What is the project name?
  Enter a value (Default is k8s-playground): k8s-playground

  %

By default the configuration will be created at $HOME/.tarmak/tarmak.yaml. Now we create an image for our environment by running `tarmak image-build` (this is the step that requires docker to be installed locally).

::

  % tarmak image-build 
  <long output omitted>

Finally, we can apply our context to AWS using `tarmak terraform-apply` which will spin up our cluster. If we want to tear down the cluster we can run `tarmak terraform-destroy`.