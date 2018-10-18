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

  -h, --help           help for logs
      --path string    target tar ball path (default "${TARMAK_CONFIG}/${CURRENT_CLUSTER}/${INSTANCE_POOL}-logs.tar.gz")
      --since string   gather logs since date (default "2018-11-27 10:10:08")
      --until string   gather logs until date (default "2018-11-28 10:10:08")

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

* `tarmak clusters <tarmak_clusters.html>`_ 	 - Operations on clusters

