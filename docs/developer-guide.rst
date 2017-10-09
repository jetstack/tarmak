.. dev-guide:

Developer guide
===============

Here we will walk through how to compile the Tarmak CLI and documentation from source.

Building Tarmak
---------------

Prerequisites
*************

* Go (for the CLI)
* Python 2.x (for documentation)
* `virtualenv <https://pypi.python.org/pypi/virtualenv>`_ and `virtualenvwrapper <https://virtualenvwrapper.readthedocs.io>`_ (for documentation)

Building Tarmak binary
**********************

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

Building Tarmak documentation
*****************************

To build the documentation run the following.

::

  cd $GOPATH/src/github.com/jetstack/tarmak/docs
  mkvirtualenv -p $(which python2) tarmak-docs
  pip install -r requirements.txt
  make html

You can now open `_build/html/index.html` in a browser.
