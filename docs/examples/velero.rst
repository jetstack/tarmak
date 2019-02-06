Velero
------

`Velero <https://heptio.github.io/velero/>`_ is a tool maintained by Heptio that allows
you to back up and restore your Kubernetes cluster along with its persistent volumes.

In addition to being a backup tool, it can also be used to replicate an existing
cluster in a different environment  in order to establish a testing environment.
Velero comprises a server that runs in your cluster and a local command-line client.
Cluster resources can be backed up locally or to a cloud object storage service like
AWS S3, Google Cloud Storage or Azure Blob Storage.

Setup
~~~~~

Configure AWS S3
++++++++++++++++

Once the `latest release <https://github.com/heptio/velero/releases>`_ of Velero
has been installed on your local machine, the Velero documentation has a
`step-by-step guide <https://heptio.github.io/velero/v0.10.0/aws-config>`_
outlining how to configure Velero to interact with your AWS environment.

The following steps, outlined in detail in the Velero documentation, are required:

* Create an S3 bucket in which cluster backups can be stored. Heptio recommends a unique S3 bucket for each Kubernetes cluster
* Create an IAM user for Velero. Heptio recommends a unique username per cluster

.. note::
   If you are using Tarmak with kube2iam, Velero can be used alongside kube2iam
   by defining a Trust Policy. This process is defined in Velero's
   `step-by-step AWS guide
   <https://heptio.github.io/velero/v0.10.0/aws-config>`_ .

Install Velero on remote cluster
+++++++++++++++++++++++++++++++++

::

  $ git clone https://github.com/heptio/velero.git
  $ tarmak kubectl apply -f velero/examples/common/00-prereqs.yaml
  $ tarmak kubectl apply -f velero/examples/minio/

The ``00-prereqs.yaml`` file creates a `heptio-velero` namespace, an `velero`
service account and applies RBAC rules to grant permissions to that service
account, as well as CustomResourceDefinitions for the resources used by
`velero`.

The ``minio`` YAMLs install `Minio <https://github.com/minio/minio>`_, an object storage server 
compatible with AWS S3 (and other object storage services).

Operation
~~~~~~~~~

Using Velero on a Tarmak cluster
+++++++++++++++++++++++++++++++++

In order to run Velero operations against the cluster (i.e. ``velero backup`` /
``velero restore`` / ``velero schedule``), run the following tarmak command to
ensure that an SSH tunnel is open, and that the current  cluster's kubeconfig
file has been saved locally and set as the ``KUBECONFIG`` environment variable:

::

  $ export $(tarmak cluster kubeconfig)

Velero will now we able to interact with the Tarmak cluster. If you have
deployed velero in a non-default namespace (default is `heptio-velero`) on your
cluster, you'll need to specify this with a ``--namespace`` flag.

Recovery and migration
++++++++++++++++++++++

Guidance for using Velero for `Disaster recovery
<https://heptio.github.io/velero/v0.10.0/disaster-case>`_ and `Cluster migration
<https://heptio.github.io/velero/v0.10.0/migration-case>`_ are outlined on the
Velero documentation pages.
