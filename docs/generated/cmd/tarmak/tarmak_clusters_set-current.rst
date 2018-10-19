.. _tarmak_clusters_set-current:

tarmak clusters set-current
---------------------------

Set current cluster in config

Synopsis
~~~~~~~~


Set current cluster in config

::

  tarmak clusters set-current [flags]

Options
~~~~~~~

::

  -h, --help   help for set-current

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

