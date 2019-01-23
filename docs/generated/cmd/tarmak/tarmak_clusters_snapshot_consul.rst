.. _tarmak_clusters_snapshot_consul:

tarmak clusters snapshot consul
-------------------------------

Manage snapshots on remote consul clusters

Synopsis
~~~~~~~~


Manage snapshots on remote consul clusters

Options
~~~~~~~

::

  -h, --help   help for consul

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config-directory string                          config directory for tarmak's configuration (default "~/.tarmak")
      --current-cluster string                           override the current cluster set in the config
      --ignore-missing-public-key-tags ssh_known_hosts   ignore missing public key tags on instances, by falling back to populating ssh_known_hosts with the first connection (default true)
      --keep-containers                                  do not clean-up terraform/packer containers after running them
      --public-api-endpoint                              Override kubeconfig to point to cluster's public API endpoint
  -v, --verbose                                          enable verbose logging
      --wing-dev-mode                                    use a bundled wing version rather than a tagged release from GitHub

SEE ALSO
~~~~~~~~

* `tarmak clusters snapshot <tarmak_clusters_snapshot.html>`_ 	 - Manage snapshots of remote consul and etcd clusters
* `tarmak clusters snapshot consul restore <tarmak_clusters_snapshot_consul_restore.html>`_ 	 - restore consul cluster with source snapshot
* `tarmak clusters snapshot consul save <tarmak_clusters_snapshot_consul_save.html>`_ 	 - save consul cluster snapshot to target path

