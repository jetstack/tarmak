.. _tarmak_environments_list:

tarmak environments list
------------------------

Print a list of environments

Synopsis
~~~~~~~~


Print a list of environments

::

  tarmak environments list [flags]

Options
~~~~~~~

::

  -h, --help   help for list

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config-directory string   config directory for tarmak's configuration (default "~/.tarmak")
      --current-cluster string    override the current cluster set in the config
      --keep-containers           do not clean-up terraform/packer containers after running them
      --public-api-endpoint       Override kubeconfig to point to cluster's public API endpoint
  -v, --verbose                   enable verbose logging
      --wing-dev-mode             use a bundled wing version rather than a tagged release from GitHub

SEE ALSO
~~~~~~~~

* `tarmak environments <tarmak_environments.html>`_ 	 - Operations on environments

