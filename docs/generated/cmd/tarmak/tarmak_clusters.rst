.. _tarmak_clusters:

tarmak clusters
---------------

Operations on clusters

Synopsis
~~~~~~~~


Operations on clusters

Options
~~~~~~~

::

  -h, --help   help for clusters

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

* `tarmak <tarmak.html>`_ 	 - Tarmak is a toolkit for provisioning and managing Kubernetes clusters.
* `tarmak clusters apply <tarmak_clusters_apply.html>`_ 	 - Create or update the currently configured cluster
* `tarmak clusters debug <tarmak_clusters_debug.html>`_ 	 - Operations for debugging a cluster
* `tarmak clusters destroy <tarmak_clusters_destroy.html>`_ 	 - Destroy the current cluster
* `tarmak clusters force-unlock <tarmak_clusters_force-unlock.html>`_ 	 - Remove remote lock using lock ID
* `tarmak clusters images <tarmak_clusters_images.html>`_ 	 - Operations on images
* `tarmak clusters init <tarmak_clusters_init.html>`_ 	 - Initialize a cluster
* `tarmak clusters instances <tarmak_clusters_instances.html>`_ 	 - Operations on instances
* `tarmak clusters kubeconfig <tarmak_clusters_kubeconfig.html>`_ 	 - Verify and print path to Kubeconfig
* `tarmak clusters kubectl <tarmak_clusters_kubectl.html>`_ 	 - Run kubectl on the current cluster
* `tarmak clusters list <tarmak_clusters_list.html>`_ 	 - Print a list of clusters
* `tarmak clusters plan <tarmak_clusters_plan.html>`_ 	 - Plan changes on the currently configured cluster
* `tarmak clusters set-current <tarmak_clusters_set-current.html>`_ 	 - Set current cluster in config
* `tarmak clusters snapshot <tarmak_clusters_snapshot.html>`_ 	 - Manage snapshots of remote consul and etcd clusters
* `tarmak clusters ssh <tarmak_clusters_ssh.html>`_ 	 - Log into an instance with SSH

