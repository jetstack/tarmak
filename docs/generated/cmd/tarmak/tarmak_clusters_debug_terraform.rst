.. _tarmak_clusters_debug_terraform:

tarmak clusters debug terraform
-------------------------------

Operations for debugging Terraform configuration

Synopsis
~~~~~~~~


Operations for debugging Terraform configuration

Options
~~~~~~~

::

  -h, --help   help for terraform

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config-directory string        config directory for tarmak's configuration (default "~/.tarmak")
      --current-cluster string         override the current cluster set in the config
      --keep-containers                do not clean-up terraform/packer containers after running them
      --log-flush-frequency duration   Maximum number of seconds between log flushes (default 5s)
      --public-api-endpoint            Override kubeconfig to point to cluster's public API endpoint
  -v, --verbose                        enable verbose logging
      --wing-dev-mode                  use a bundled wing version rather than a tagged release from GitHub

SEE ALSO
~~~~~~~~

* `tarmak clusters debug <tarmak_clusters_debug.html>`_ 	 - Operations for debugging a cluster
* `tarmak clusters debug terraform shell <tarmak_clusters_debug_terraform_shell.html>`_ 	 - Prepares a Terraform container and executes a shell in this container

