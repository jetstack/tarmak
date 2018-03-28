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
  git clone git@github.com:jetstack/tarmak.git
  cd tarmak
  make build
  ln -s $PWD/tarmak_$(uname -s | tr '[:upper:]' '[:lower:]')_amd64 /usr/local/bin/tarmak

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


Or using docker:

::

  cd $GOPATH/src/github.com/jetstack/tarmak/docs
  make docker_html

You can now open ``_build/html/index.html`` in a browser or serve the site with
a `web server of your choice <https://gist.github.com/willurd/5720255>`_.


Updating puppet subtrees
************************

Puppet modules are maintained as separate repositories, which get bundled into
tarmak using git subtree. To pull the latest changes from the upstream repositories,
run ``make subtrees``.


Release Checklist
-----------------

This is a list to collect manual tasks/checks necessary for cutting a
release of Tarmak:

* Ensure release references are updated (don't forget to commit)

::

  make release VERSION=x.y.x

* Tag release commit with ``x.y.z`` and push to GitLab and GitHub
* Update the CHANGELOG using the release notes

::

  # relnotes is the golang tool from https://github.com/kubernetes/release/tree/master/toolbox/relnotes
  relnotes -repo tarmak -owner jetstack -doc-url=https://docs.tarmak.io -htmlize-md -markdown-file CHANGELOGX.md x.y(-1).z-1..x.y.z

* Branch out minor releases into ``release-x.y``

After release job has run:

* Make sure we update the generated `releases <https://github.com/jetstack/tarmak/releases>`_ page
