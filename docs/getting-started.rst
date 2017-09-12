.. getting-started:

Getting Started
================

Here we will walk through how to use Tarmak to spin up a Kubernetes cluster in AWS. This will deploy the Kubernetes master nodes, an etcd cluster, worker nodes, vault and a bastion node with a public IP address.

Prerequisites
-------------

* Docker
* An AWS account
* A public zone that you can delegate to AWS
* Vault with the `AWS secret backend <https://www.vaultproject.io/docs/secrets/aws/index.html>`_ configured (optional)

Steps
-----

The plan will be to first initialise our cluster configuration for our environment, build an image for our configuration and then finally apply our configuration to create our cluster.

Run `tarmak init` to initialise our configuration.

Note that if you are not using Vault's AWS secret backend you can authenticate with AWS in the same ways as the AWS CLI. More details can be found at `Configuring the AWS CLI <http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html>`_.

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

By default the configuration will be created at $HOME/.tarmak/tarmak.yaml. 

Now we create an image for our environment by running `tarmak image-build` (this is the step that requires docker to be installed locally).

::

  % tarmak image-build 
  <long output omitted>

We can now apply our configuration to AWS by running `tarmak terraform-apply`. Note that the first time you run this command Tarmak will create a `hosted zone <http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/CreatingHostedZone.html>`_ for you and then fail with the following error. 

::

  * failed verifying delegation of public zone 5 times, make sure the zone k8s.jetstack.io is delegated to nameservers [ns-100.awsdns-12.com ns-1283.awsdns-32.org ns-1638.awsdns-12.co.uk ns-842.awsdns-41.net]

To fix this, change the nameservers of your domain to the four listed in the error. If you only want to delegate a subdomain containing your zone to AWS without delegating the parent domain see `Creating a Subdomain That Uses Amazon Route 53 as the DNS Service without Migrating the Parent Domain <http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/CreatingNewSubdomain.html>`_. 

We can now re-run `tarmak terraform-apply` to spin up our cluster. If we want to tear down the cluster at any point we can run `tarmak terraform-destroy`.