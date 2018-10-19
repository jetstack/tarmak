.. _tarmak:

tarmak
------

Tarmak is a toolkit for provisioning and managing Kubernetes clusters.

Synopsis
~~~~~~~~


Tarmak is a toolkit for provisioning and managing Kubernetes clusters.

Options
~~~~~~~

::

  -c, --config-directory string        config directory for tarmak's configuration (default "~/.tarmak")
      --current-cluster string         override the current cluster set in the config
  -h, --help                           help for tarmak
      --keep-containers                do not clean-up terraform/packer containers after running them
      --log-flush-frequency duration   Maximum number of seconds between log flushes (default 5s)
  -v, --verbose                        enable verbose logging
      --wing-dev-mode                  use a bundled wing version rather than a tagged release from GitHub

SEE ALSO
~~~~~~~~

* `tarmak clusters <tarmak_clusters.rst>`_ 	 - Operations on clusters
* `tarmak environments <tarmak_environments.rst>`_ 	 - Operations on environments
* `tarmak init <tarmak_init.rst>`_ 	 - Initialize a cluster
* `tarmak kubectl <tarmak_kubectl.rst>`_ 	 - Run kubectl on the current cluster
* `tarmak providers <tarmak_providers.rst>`_ 	 - Operations on providers
* `tarmak version <tarmak_version.rst>`_ 	 - Print the version number of tarmak

