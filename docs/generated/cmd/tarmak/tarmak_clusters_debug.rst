.. _tarmak_clusters_debug:

tarmak clusters debug
---------------------

Operations for debugging a cluster

Synopsis
~~~~~~~~


Operations for debugging a cluster

Options
~~~~~~~

::

  -h, --help   help for debug

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config-directory string        config directory for tarmak's configuration (default "~/.tarmak")
      --current-cluster string         override the current cluster set in the config
      --keep-containers                do not clean-up terraform/packer containers after running them
      --log-flush-frequency duration   Maximum number of seconds between log flushes (default 5s)
  -v, --verbose                        enable verbose logging
      --wing-dev-mode                  use a bundled wing version rather than a tagged release from GitHub

SEE ALSO
~~~~~~~~

* `tarmak clusters <tarmak_clusters.rst>`_ 	 - Operations on clusters
* `tarmak clusters debug puppet <tarmak_clusters_debug_puppet.rst>`_ 	 - Operations for debugging Puppet configuration
* `tarmak clusters debug terraform <tarmak_clusters_debug_terraform.rst>`_ 	 - Operations for debugging Terraform configuration

