.. _tarmak_clusters_snapshot_etcd_restore:

tarmak clusters snapshot etcd restore
-------------------------------------

restore etcd cluster with source snapshots

Synopsis
~~~~~~~~


restore etcd cluster with source snapshots

::

  tarmak clusters snapshot etcd restore [flags]

Options
~~~~~~~

::

  -h, --help                help for restore
      --k8s-events string   location of k8s-events snapshot backup
      --k8s-main string     location of k8s-main snapshot backup
      --overlay string      location of overlay snapshot backup

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

* `tarmak clusters snapshot etcd <tarmak_clusters_snapshot_etcd.html>`_ 	 - Manage snapshots on remote etcd clusters

