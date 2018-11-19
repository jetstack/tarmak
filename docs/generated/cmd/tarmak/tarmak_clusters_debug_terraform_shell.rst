.. _tarmak_clusters_debug_terraform_shell:

tarmak clusters debug terraform shell
-------------------------------------

Prepares a Terraform container and executes a shell in this container

Synopsis
~~~~~~~~


Prepares a Terraform container and executes a shell in this container

::

  tarmak clusters debug terraform shell [flags]

Options
~~~~~~~

::

  -h, --help   help for shell

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

* `tarmak clusters debug terraform <tarmak_clusters_debug_terraform.html>`_ 	 - Operations for debugging Terraform configuration

