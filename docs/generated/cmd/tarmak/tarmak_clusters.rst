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
  -v, --verbose                        enable verbose logging
      --wing-dev-mode                  use a bundled wing version rather than a tagged release from GitHub

SEE ALSO
~~~~~~~~

* `tarmak <tarmak.rst>`_ 	 - Tarmak is a toolkit for provisioning and managing Kubernetes clusters.
* `tarmak clusters apply <tarmak_clusters_apply.rst>`_ 	 - Create or update the currently configured cluster
* `tarmak clusters debug <tarmak_clusters_debug.rst>`_ 	 - Operations for debugging a cluster
* `tarmak clusters destroy <tarmak_clusters_destroy.rst>`_ 	 - Destroy the current cluster
* `tarmak clusters force-unlock <tarmak_clusters_force-unlock.rst>`_ 	 - Remove remote lock using lock ID
* `tarmak clusters images <tarmak_clusters_images.rst>`_ 	 - Operations on images
* `tarmak clusters init <tarmak_clusters_init.rst>`_ 	 - Initialize a cluster
* `tarmak clusters instances <tarmak_clusters_instances.rst>`_ 	 - Operations on instances
* `tarmak clusters kubectl <tarmak_clusters_kubectl.rst>`_ 	 - Run kubectl on the current cluster
* `tarmak clusters list <tarmak_clusters_list.rst>`_ 	 - Print a list of clusters
* `tarmak clusters plan <tarmak_clusters_plan.rst>`_ 	 - Plan changes on the currently configured cluster
* `tarmak clusters set-current <tarmak_clusters_set-current.rst>`_ 	 - Set current cluster in config
* `tarmak clusters ssh <tarmak_clusters_ssh.rst>`_ 	 - Log into an instance with SSH

