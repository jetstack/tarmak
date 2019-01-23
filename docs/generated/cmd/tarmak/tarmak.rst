.. _tarmak:

tarmak
------

Tarmak is a toolkit for provisioning and managing Kubernetes clusters.

Synopsis
~~~~~~~~


Tarmak is a toolkit for provisioning and managing Kubernetes clusters.

Options
~~~~~~~

::

  -c, --config-directory string                          config directory for tarmak's configuration (default "~/.tarmak")
      --current-cluster string                           override the current cluster set in the config
  -h, --help                                             help for tarmak
      --ignore-missing-public-key-tags ssh_known_hosts   ignore missing public key tags on instances, by falling back to populating ssh_known_hosts with the first connection (default true)
      --keep-containers                                  do not clean-up terraform/packer containers after running them
      --public-api-endpoint                              Override kubeconfig to point to cluster's public API endpoint
  -v, --verbose                                          enable verbose logging
      --wing-dev-mode                                    use a bundled wing version rather than a tagged release from GitHub

SEE ALSO
~~~~~~~~

* `tarmak clusters <tarmak_clusters.html>`_ 	 - Operations on clusters
* `tarmak environments <tarmak_environments.html>`_ 	 - Operations on environments
* `tarmak init <tarmak_init.html>`_ 	 - Initialize a cluster
* `tarmak kubeconfig <tarmak_kubeconfig.html>`_ 	 - Verify and print path to Kubeconfig
* `tarmak kubectl <tarmak_kubectl.html>`_ 	 - Run kubectl on the current cluster
* `tarmak providers <tarmak_providers.html>`_ 	 - Operations on providers
* `tarmak version <tarmak_version.html>`_ 	 - Print the version number of tarmak

