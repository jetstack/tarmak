.. _tarmak_clusters_destroy:

tarmak clusters destroy
-----------------------

Destroy the current cluster

Synopsis
~~~~~~~~


Destroy the current cluster

::

  tarmak clusters destroy [flags]

Options
~~~~~~~

::

      --dry-run                             don't actually change anything, just show changes that would occur
      --force-destroy-state-stack           force destroy the state stack, this is unreversible (!!!)
  -h, --help                                help for destroy
  -S, --infrastructure-stacks stringArray   run operation on these stacks only, valid stacks are: state, network, tools, bastion, vault, kubernetes

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

