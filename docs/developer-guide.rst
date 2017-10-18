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

  $ tarmak
  Tarmak is a toolkit for provisioning and managing Kubernetes clusters.

  Usage:
    tarmak [command]

  Available Commands:
    clusters     Operations on clusters
    environments Operations on environments
    help         Help about any command
    init         Initialize a cluster
    kubectl      Run kubectl on the current cluster
    providers    Operations on providers
    version      Print the version number of tarmak

  Flags:
    -c, --config-directory string   config directory for tarmak's configuration (default "~/.tarmak")
    -h, --help                      help for tarmak
    -v, --verbose                   enable verbose logging

  Use "tarmak [command] --help" for more information about a command.

Building Tarmak documentation
*****************************

To build the documentation run the following.

::

  cd $GOPATH/src/github.com/jetstack/tarmak/docs
  make html

You can now open `_build/html/index.html` in a browser.
