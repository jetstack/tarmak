.. _tarmak_clusters_logs:

tarmak clusters logs
--------------------

Gather logs from a list of instances or target groups

Synopsis
~~~~~~~~


Gather logs from a list of instances or target groups [bastion vault etcd worker control-plane]

::

  tarmak clusters logs [target groups] [flags]

Options
~~~~~~~

::

  -h, --help           help for logs
      --path string    target tar ball path (default "./[target group]-logs.tar.gz")
      --since string   gather logs since date (default "$(date --date='24 hours ago')")
      --until string   gather logs until date (default "$(date --date='now')")

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

