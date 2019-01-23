.. _tarmak_clusters_snapshot_etcd:

tarmak clusters snapshot etcd
-----------------------------

Manage snapshots on remote etcd clusters

Synopsis
~~~~~~~~


Manage snapshots on remote etcd clusters

Options
~~~~~~~

::

  -h, --help   help for etcd

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
* `tarmak clusters snapshot etcd restore <tarmak_clusters_snapshot_etcd_restore.html>`_ 	 - restore etcd cluster with source snapshots
* `tarmak clusters snapshot etcd save <tarmak_clusters_snapshot_etcd_save.html>`_ 	 - save etcd snapshot to target path prefix, i.e 'backup-'

