.. getting-started:

User guide
==========

Getting started with AWS
------------------------

In this getting started guide, we walk through how to use initialise Tarmak with a new Provider (AWS) and Environment and provision a Kubernetes cluster. 
This will comprise Kubernetes master and worker nodes, etcd clusters, Vault and a bastion node with a public IP address 
(see :doc:`Architecture overview` for details of cluster components)

Prerequisites
~~~~~~~~~~~~~

* Docker
* An AWS account
* A public DNS zone that can be delegated to AWS Route 53
* Optional: Vault with the `AWS secret backend <https://www.vaultproject.io/docs/secrets/aws/index.html>`_ configured

Overview of steps to follow
~~~~~~~~~~~~~~~~~~~~~~~~~~~

* Initialise cluster configuration
* Build an image (AMI) 
* Create the cluster
* Destroy the cluster

Initialise configuration
~~~~~~~~~~~~~~~~~~~~~~~~

 in this step, a Provider and an Environment is configured.

Simply run ``tarmak init`` to initialise configuration for the first time. This will set-up a Provider (AWS) and an Environment.
You will be prompted for the necessary configuration.

.. note::
   If you are not using Vault's AWS secret backend, you can authenticate with AWS in the same way as the AWS CLI. More details can be found at `Configuring the AWS CLI <http://docs.aws.amazon.com /cli/latest/userguide/cli-chap-getting-started.html>`_.

::

  % tarmak init
  What should be the name of the cluster?

  The name consists of two parts seperated by a dash. The first part is the environment name, second part the cluster name. Both names should be matching [a-z0-9]+

  Enter a value: dev-cluster

  Do you want to use vault to get credentials for AWS? [Y/N]
  Enter a value (Default is N): Y

  Which path should be used for AWS credentials?
  Enter a value (Default is jetstack/aws/jetstack-dev/sts/admin): jetstack/aws/jetstack-dev/sts/admin

  Which region should be used?
  Enter a value (Default is eu-west-1): eu-west-1

  What bucket prefix should be used?
  Enter a value (Default is tarmak-): tarmak-

  What public zone should be used?

  Please make sure you can delegate this zone to AWS!

  Enter a value: k8s.jetstack.io

  What private zone should be used?
  Enter a value (Default is tarmak.local): tarmak.local

  What is the mail address of someone responsible?
  Enter a value: luke.addison@jetstack.io

  What is the project name?
  Enter a value (Default is k8s-playground): k8s-playground

  %

.. todo::
   Make sure this ``tarmak init`` stuff is up to date (tarmak is asking different questions now)

By default the configuration will be created at ``$HOME/.tarmak/tarmak.yaml``.

Create an AMI
~~~~~~~~~~~~~
Next we create an AMI for this environment by running ``tarmak clusters images build`` (this is the step that requires Docker to be installed locally).

::

  % tarmak clusters images build
  <output omitted>

Create the cluster
~~~~~~~~~~~~~~~~~~
To create the cluster, run ``tarmak clusters apply``.

::

  % tarmak clusters apply
  <output omitted>

.. warning::
   The first time this command is run, Tarmak will create a `hosted zone <http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/CreatingHostedZone.html>`_ and then fail with the following error.

   ::

      * failed verifying delegation of public zone 5 times, make sure the zone k8s.jetstack.io is delegated to nameservers [ns-100.awsdns-12.com ns-1283.awsdns-32.org ns-1638.awsdns-12.co.uk ns-842.awsdns-41.net]

You should now change the nameservers of your domain to the four listed in the error. If you only wish to delegate a subdomain containing your zone to AWS without delegating the parent domain see `Creating a Subdomain That Uses Amazon Route 53 as the DNS Service without Migrating the Parent Domain <http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/CreatingNewSubdomain.html>`_.

To complete the cluster provisioning, run ``tarmak clusters apply`` once again.

.. note::
   This process may take 30-60 minutes to complete.
   You can stop it by sending the signal `SIGTERM` or `SIGINT` (Ctrl-C) to the process.
   Tarmak will not exit immediately.
   It will wait for the currently running step to finish and then exit.
   You can complete the process by re-running the command.

Destroy the cluster
~~~~~~~~~~~~~~~~~~~
To destroy the cluster, run ``tarmak clusters destroy``.

::

  % tarmak clusters destroy
  <output omitted>

.. note::
   This process may take 30-60 minutes to complete.
   You can stop it by sending the signal ``SIGTERM`` or ``SIGINT`` (Ctrl-C) to the process.
   Tarmak will not exit immediately.
   It will wait for the currently running step to finish and then exit.
   You can complete the process by re-running the command.
