.. _tarmak_clusters_apply:

tarmak clusters apply
---------------------

Create or update the currently configured cluster

Synopsis
~~~~~~~~


Create or update the currently configured cluster

::

  tarmak clusters apply [flags]

Options
~~~~~~~

::

  -C, --configuration-only                  apply changes to configuration only, by running only puppet
      --dry-run                             don't actually change anything, just show changes that would occur
  -h, --help                                help for apply
  -I, --infrastructure-only                 apply changes to infrastructure only, by running only terraform
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

* `tarmak clusters <tarmak_clusters.html>`_ 	 - Operations on clusters

