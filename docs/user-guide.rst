.. getting-started:

User guide
==========

Getting started with AWS
------------------------

In this getting started guide, we walk through how to use initialise Tarmak with a new Provider (AWS) and Environment and provision a Kubernetes cluster.
This will comprise Kubernetes master and worker nodes, etcd clusters, Vault and a bastion node with a public IP address
(see :ref:`Architecture overview <architecture_overview>` for details of cluster components)

Prerequisites
~~~~~~~~~~~~~

* Docker
* An AWS account that has `accepted the CentOS licence terms <aws-centos-guide.html>`_
* A public DNS zone that can be delegated to AWS Route 53
* Optional: Vault with the `AWS secret backend <https://www.vaultproject.io/docs/secrets/aws/index.html>`_ configured

Overview of steps to follow
~~~~~~~~~~~~~~~~~~~~~~~~~~~

* :ref:`Initialise cluster configuration <init_config>`
* :ref:`Build an image (AMI) <create_ami>`
* :ref:`Create the cluster <create_cluster>`
* :ref:`Destroy the cluster <destroy_cluster>`

.. _init_config:

Initialise configuration
~~~~~~~~~~~~~~~~~~~~~~~~

Simply run ``tarmak init`` to initialise configuration for the first time. You will be prompted for the necessary configuration
to set-up a new :ref:`Provider <providers_resource>` (AWS) and :ref:`Environment <environments_resource>`. The list below describes
the questions you will be asked.

.. note::
   If you are not using Vault's AWS secret backend, you can authenticate with AWS in the same way as the AWS CLI. More details can be found at `Configuring the AWS CLI <http://docs.aws.amazon.com /cli/latest/userguide/cli-chap-getting-started.html>`_.

* Configuring a new :ref:`Provider <providers_resource>`
   * Provider name: must be unique
   * Cloud: Amazon (AWS) is the default and only option for now (more clouds to come)
   * Credentials: Amazon CLI auth (i.e. env variables/profile) or Vault (optional)
   * Name prefix: for state buckets and DynamoDB tables
   * Public DNS zone: will be created if not already existing, must be delegated from the root

* Configuring a new :ref:`Environment <environments_resource>`
   * Environment name: must be unique
   * Project name: used for AWS resource labels
   * Project administrator mail address
   * Cloud region: pick a region fetched from AWS (using Provider credentials)

* Configuring new :ref:`Cluster(s) <clusters_resource>`
   * Single or multi-cluster environment
   * Cloud availability zone(s): pick zone(s) fetched from AWS

Once initialised, the configuration will be created at ``$HOME/.tarmak/tarmak.yaml`` (default).

.. _create_ami:

Create an AMI
~~~~~~~~~~~~~
Next we create an AMI for this environment by running ``tarmak clusters images build`` (this is the step that requires Docker to be installed locally).

::

  % tarmak clusters images build
  <output omitted>

.. _create_cluster:

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

.. _destroy_cluster:

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

Configuration Options
---------------------

After generating your `tarmak.yaml` configuration file there are a number of options you can set that are not exposed via `tarmak init`.

Pod Security Policy
~~~~~~~~~~~~~~~~~~~
**Note:** For cluster versions greater than 1.8.0 this is applied by default.
For cluster versions before 1.6.0 is it not applied.

To enable Pod Security Policy to an environment, include the following to the
configuration file under the kubernetes field of that environment:

::

    kubernetes:
        podSecurityPolicy:
            enabled: true

The configuration file can be found at ``$HOME/.tarmak/tarmak.yaml`` (default).
The Pod Security Policy manifests can be found within the tarmak directory at
``puppet/modules/kubernetes/templates/pod-security-policy.yaml.erb``

Logging
~~~~~~~

Each Kubernetes cluster can be configured with a number of logging sinks. The
only sink currently supported is Elasticsearch. An example configuration is
shown below:

.. code-block:: yaml

  apiVersion: api.tarmak.io/v1alpha1
  kind: Config
  clusters:
  - loggingSinks:
  - types:
    - application
    - platform
    elasticsearch:
      host: example.amazonaws.com
      port: 443
      logstashPrefix: test
      tls: true
      tlsVerify: false
      httpBasicAuth:
        username: administrator
        password: mypassword
  - types:
    - all
    elasticsearch:
      host: example2.amazonaws.com
      port: 443
      tls: true
      amazonESProxy:
        port: 9200
  ...


A full list of the configuration parameters are shown below:

* General configuration parameters

    * ``types`` - the types of logs to ship. The accepted values are:

        * platform (kernel, systemd and platform namespace logs)

        * application (all other namespaces)

        * audit (apiserver audit logs)

        * all

* Elasticsearch configuration parameters
    * ``host`` - IP address or hostname of the target Elasticsearch instance

    * ``port`` - TCP port of the target Elasticsearch instance

    * ``logstashPrefix`` - Shipped logs are in a Logstash compatible format.
      This field specifies the Logstash index prefix * ``tls`` - enable or
      disable TLS support

    * ``tlsVerify`` - force certificate validation (only valid when not using
      the AWS ES Proxy)

    * ``tlsCA`` - Custom CA certificate for Elasticsearch instance (only valid
      when not using the AWS ES Proxy)

    * ``httpBasicAuth`` - configure basic auth (only valid when not using the
      AWS ES Proxy)

        * ``username``

        * ``password``

    * ``amazonESProxy`` - configure AWS ES Proxy

        * ``port`` - Port to listen on (a free port will be chosen for you if
          omitted)

Jenkins
~~~~~~~

You can install Jenkins as part of your hub. This can be achieved by adding an extra instance pool to your hub.
This instance pool can be extended with an annotation ``tarmak.io/jenkins-certificate-arn``. The value of this annotation will be an arn pointing to an Amazon Certificate.
When you set this annotation, your Jenkins will be secured with https. You need to make sure your SSL certificate is valid for jenkins.<environment>.<zone>.

``` code-block:: yaml

  - image: centos-puppet-agent
    maxCount: 1
    metadata:
      annotations:
        tarmak.io/jenkins-certificate-arn: "arn:aws:acm:eu-west-1:228615251467:certificate/81e0c595-f5ad-40b2-8062-683b215bedcf"
      creationTimestamp: null
      name: jenkins
    minCount: 1
    size: large
    type: jenkins
    volumes:
    - metadata:
        creationTimestamp: null
        name: root
      size: 16Gi
      type: ssd
    - metadata:
        creationTimestamp: null
        name: data
      size: 16Gi
      type: ssd
```

Setting up an AWS hosted Elasticsearch Cluster
++++++++++++++++++++++++++++++++++++++++++++++

AWS provides a hosted Elasticsearch cluster that can be used for log
aggregation. This snippet will setup an Elasticsearch domain in your account
and create a policy along with it that will allow shipping of logs into the
cluster:


.. literalinclude:: user-guide/aws-elasticsearch/elasticsearch.tf


Once terraform has been successfully run it will output, the resulting AWS
Elasticsearch endpoint and the policy that allow shipping to it:

::

  Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
  
  Outputs:
  
  elasticsearch_endpoint = search-tarmak-logs-xyz.eu-west-1.es.amazonaws.com
  elasticsearch_shipping_policy_arn = arn:aws:iam::1234:policy/tarmak-logs-shipping

Both of those outputs can then be used in the tarmak configuration:

.. code-block:: yaml

  apiVersion: api.tarmak.io/v1alpha1
  clusters:
  - name: cluster
    loggingSinks:
    - types: ["all"]
      elasticsearch:
        host: ${elasticsearch_endpoint}
        tls: true
        amazonESProxy: {}
    amazon:
      additionalIAMPolicies:
      - ${elasticsearch_shipping_policy_arn}
