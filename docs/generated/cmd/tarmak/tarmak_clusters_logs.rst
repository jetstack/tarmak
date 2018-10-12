.. _tarmak_clusters_logs:

tarmak clusters logs
--------------------

Gather logs from an instance pool

Synopsis
~~~~~~~~


Gather logs from an instance pool

::

  tarmak clusters logs [flags]

Options
~~~~~~~

::

  -h, --help          help for logs
      --path string   location to store tar ball of bundled systemd unit logs (default "${TARMAK_CONFIG}/${CURRENT_CLUSTER}/${INSTANCE_POOL}.tar.gz")

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

