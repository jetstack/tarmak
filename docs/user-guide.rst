.. getting-started:

User guide
==========

Getting started with AWS
------------------------

In this getting started guide, we walk through how to initialise Tarmak with a
new Provider (AWS), a new Environment and then provision a Kubernetes cluster.
This will comprise of Kubernetes master and worker nodes, etcd clusters, Vault
and a bastion node with a public IP address (see :ref:`Architecture overview
<architecture_overview>` for details of cluster components)

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

Simply run ``tarmak init`` to initialise configuration for the first time. You
will be prompted for the necessary configuration to set-up a new :ref:`Provider
<providers_resource>` (AWS) and :ref:`Environment <environments_resource>`. The
list below describes the questions you will be asked.

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

Next we create an AMI for this environment by running ``tarmak clusters images
build`` (this is the step that requires Docker to be installed locally).

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
   The first time this command is run, Tarmak will create a `hosted zone
   <http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/CreatingHostedZone.html>`_
   and then fail with the following error.

   ::

      * failed verifying delegation of public zone 5 times, make sure the zone k8s.jetstack.io is delegated to nameservers [ns-100.awsdns-12.com ns-1283.awsdns-32.org ns-1638.awsdns-12.co.uk ns-842.awsdns-41.net]

   When creating a multi-cluster environment, the hub cluster must first be
   applied . To change the current cluster use the flag ``--current-cluster``.
   See ``tarmak cluster help`` for more information.

You should now change the nameservers of your domain to the four listed in the
error. If you only wish to delegate a subdomain containing your zone to AWS
without delegating the parent domain see `Creating a Subdomain That Uses Amazon
Route 53 as the DNS Service without Migrating the Parent Domain
<http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/CreatingNewSubdomain.html>`_.

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

After generating your `tarmak.yaml` configuration file there are a number of
options you can set that are not exposed via `tarmak init`.

Pod Security Policy
~~~~~~~~~~~~~~~~~~~

**Note:** For cluster versions greater than 1.8.0 this is applied by default.
For cluster versions before 1.6.0 is it not applied.

To enable Pod Security Policy for an environment, include the following in the
configuration file under the Kubernetes field of that environment:

::

    kubernetes:
        podSecurityPolicy:
            enabled: true

(By default, the Tarmak configuration file is stored at
``$HOME/.tarmak/tarmak.yaml``).

The PodSecurityPolicy manifests - also listed below - can be found in the
``puppet/modules/kubernetes/templates/`` directory.

- `PodSecurityPolicy RBAC <https://github.com/jetstack/tarmak/blob/master/puppet/modules/kubernetes/templates/pod-security-policy-rbac.yaml.erb>`_
- `PodSecurityPolicy <https://github.com/jetstack/tarmak/blob/master/puppet/modules/kubernetes/templates/pod-security-policy.yaml.erb>`_

Cluster Autoscaler
~~~~~~~~~~~~~~~~~~

Tarmak supports deploying `Cluster Autoscaler
<https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler>`_ when
spinning up a Kubernetes cluster to autoscale worker instance pools. The following
`tarmak.yaml` snippet shows how you would enable Cluster Autoscaler.

.. code-block:: yaml

    kubernetes:
      clusterAutoscaler:
        enabled: true
    ...

The above configuration would deploy Cluster Autoscaler with an image of
`gcr.io/google_containers/cluster-autoscaler` using the recommend version based
on the version of your Kubernetes cluster. The configuration block accepts
three optional fields of `image`, `version` and `scaleDownUtilizationThreshold`
allowing you to change these defaults.  Note that the final image tag used when
deploying Cluster Autoscaler will be the configured version prepended with the
letter `v`.

The current implementation will configure the first instance pool of type worker
in your cluster configuration to scale between `minCount` and `maxCount`. We
plan to add support for an arbitrary number of worker instance pools.

Overprovisioning
++++++++++++++++

Tarmak supports overprovisioning to give a
fixed or proportional amount of headroom in the cluster. The technique used to
implement overprovisioning is the same as described in the `cluster autoscaler
documentation
<https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#user-content-how-can-i-configure-overprovisioning-with-cluster-autoscaler>`_.
The following `tarmak.yaml` snippet shows how to configure fixed
overprovisioning. Note that cluster autoscaling must also be enabled.

.. code-block:: yaml

    kubernetes:
      clusterAutoscaler:
        enabled: true
        overprovisioning:
          enabled: true
          reservedMillicoresPerReplica: 100
          reservedMegabytesPerReplica: 100
          replicaCount: 10
    ...

This will deploy 10 pause Pods with a negative `PriorityClass
<https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/>`_
so that they will be preempted by any other pending Pods. Each Pod will request
the specified number of millicores and megabytes. The following `tarmak.yaml`
snippet shows how to configure proportional overprovisioning.

.. code-block:: yaml

    kubernetes:
      clusterAutoscaler:
        enabled: true
        overprovisioning:
          enabled: true
          reservedMillicoresPerReplica: 100
          reservedMegabytesPerReplica: 100
          nodesPerReplica: 1
          coresPerReplica: 4
    ...

The `nodesPerReplica` and `coresPerReplica` configuration parameters are
described in the `cluster-proportional-autoscaler documentation
<https://github.com/kubernetes-incubator/cluster-proportional-autoscaler#user-content-linear-mode>`_.

The image and version used by the cluster-proportional-autoscaler can also be
specified using the `image` and `version` fields of the `overprovisioning`
block. These values default to
`k8s.gcr.io/cluster-proportional-autoscaler-amd64` and `1.1.2` respectively.

Logging
~~~~~~~

Each Kubernetes cluster can be configured with a number of logging sinks. The
only sink currently supported is Elasticsearch. An example configuration is
shown below:

.. code-block:: yaml

  apiVersion: api.tarmak.io/v1alpha1
  kind: Config
  clusters:
    loggingSinks:
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

EBS Encryption
~~~~~~~~~~~~~~

AWS offers encrypted EBS (`Elastic Block Storage <https://aws.amazon.com/ebs/details/>`_); however,
`Encryption of EBS volumes <https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html>`_
is not enabled by Tarmak by default. When enabled, building the image and applying a Tarmak cluster
will take considerably longer.

The following `tarmak.yaml` snippet shows how to enable encrypted EBS.

.. code-block:: yaml

    clusters:
    - amazon:
        ebsEncrypted: true
    ...

OIDC Authentication
~~~~~~~~~~~~~~~~~~~

Tarmak supports authentication using OIDC. The following snippet demonstrates
how you would configure OIDC authentication in `tarmak.yaml`. For details on
the configuration options, visit the Kubernetes documentation `here
<https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens>`_.
Note that if the version of your cluster is less than 1.10.0, the `signingAlgs`
parameter is ignored.

.. code-block:: yaml

    kubernetes:
        apiServer:
            oidc:
                clientID: 1a2b3c4d5e6f7g8h
                groupsClaim: groups
                groupsPrefix: "oidc:"
                issuerURL: https://domain/application-server
                signingAlgs:
                - RS256
                usernameClaim: preferred_username
                usernamePrefix: "oidc:"
    ...

For the above setup, ID tokens presented to the apiserver will need to contain
claims called `preferred_username` and `groups` representing the username and
groups associated with the client. These values will then be prepended with
`oidc:` before authorisation rules are applied, so it is important that this is
taken into account when configuring cluster authorisation.

Jenkins
~~~~~~~

You can install Jenkins as part of your hub. This can be achieved by adding an
extra instance pool to your hub.  This instance pool can be extended with an
annotation ``tarmak.io/jenkins-certificate-arn``. The value of this annotation
will be ARN pointing to an Amazon Certificate.  When you set this annotation,
your Jenkins will be secured with HTTPS. You need to make sure your SSL
certificate is valid for ``jenkins.<environment>.<zone>``.

.. code-block:: yaml

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
  ...

Dashboard
~~~~~~~~~

Tarmak supports deploying `Kubernetes Dashboard
<https://github.com/kubernetes/dashboard>`_ when spinning up a Kubernetes
cluster. The following `tarmak.yaml` snippet shows how you would enable
Kubernetes Dashboard.

.. code-block:: yaml

    kubernetes:
      dashboard:
        enabled: true
    ...

The above configuration would deploy Kubernetes Dashboard with an image of
`gcr.io/google_containers/kubernetes-dashboard-amd64` using the recommended
version based on the version of your Kubernetes cluster. The configuration block
accepts two optional fields of `image` and `version` allowing you to change
these defaults. Note that the final image tag used when deploying Tiller will be
the configured version prepended with the letter `v`.

.. warning::
   Before Dashboard version 1.7, when RBAC is enabled (from Kubernetes version
   1.6) cluster-wide ``cluster-admin`` privileges are granted to Dashboard. From
   Dashboard version 1.7, only minimal privileges are granted that allow
   Dashboard to work. See Dashboard's `access control documentation
   <https://github.com/kubernetes/dashboard/wiki/Access-control>`_ for more
   details.

Tiller
~~~~~~

Tarmak supports deploying Tiller, the server-side component of `Helm
<https://github.com/kubernetes/helm>`_, when spinning up a Kubernetes cluster.
Tiller is configured to listen on localhost only which prevents arbitrary Pods
in the cluster connecting to its unauthenticated endpoint. Helm clients can
still talk to Tiller by port forwarding through the Kubernetes API Server. The
following `tarmak.yaml` snippet shows how you would enable Tiller.

.. code-block:: yaml

    kubernetes:
      tiller:
        enabled: true
        image: 2.9.1
    ...

The above configuration would deploy version 2.9.1 of Tiller with an image
of `gcr.io/kubernetes-helm/tiller`. The configuration block accepts two optional
fields of `image` and `version` allowing you to change these defaults. Note that
the final image tag used when deploying Tiller will be the
configured version prepended with the letter `v`. The version is particularly
important when deploying Tiller since its minor version
must match the minor version of any Helm clients.

.. warning::
   Tiller is deployed with the ``cluster-admin`` ClusterRole bound to its
   service account and therefore has far reaching privileges. Helm's `security
   best practices
   <https://github.com/kubernetes/helm/blob/master/docs/securing_installation.md>`_
   should also be considered.

Prometheus
~~~~~~~~~~

By default Tarmak will deploy a `Prometheus <https://prometheus.io/>`_
installation and some exporters into the ``monitoring`` namespace.

This can be completely disabled with the following cluster configuration:

.. code-block:: yaml

  kubernetes:
    prometheus:
      enabled: false

Another possibility would be to use the Tarmak provisioned Prometheus only for
scraping exporters on instances that are not part of the Kubernetes cluster.
Using federation, those metrics could then be integrated into an existing
Prometheus deployment.

To have Prometheus only monitor nodes external to the cluster, use the
following configuration instead:

.. code-block:: yaml

  kubernetes:
    prometheus:
      enabled: true
      mode: ExternalScrapeTargetsOnly

Finally, you may wish to have Tarmak only install the exporters on the external
nodes. If this is your desired configuration, then set the following mode in
the yaml:

.. code-block:: yaml

  kubernetes:
    prometheus:
      enabled: true
      mode: ExternalExportersOnly

API Server
~~~~~~~~~~~

It is possible to let Tarmak create an public endpoint for your APIserver.
This can be used together with `Secure public endpoints <user-guide.html#secure-api-server>`__.

.. code-block:: yaml

  kubernetes:
    apiServer:
      public: true

Secure public endpoints
~~~~~~~~~~~~~~~~~~~~~~~

Public endpoints (Jenkins, bastion host and if enabled apiserver) can be secured
by limiting the access to a list of CIDR blocks. This can be configured on a
environment level for all public endpoint and if wanted can be overwritten on a
specific public endpoint.

Environment level
+++++++++++++++++

This can be done by adding an ``adminCIDRs`` list to an environments block,
if nothing has been set, the default is 0.0.0.0/0:

.. code-block:: yaml

    environments:
    - contact: hello@example.com
      location: eu-west-1
      metadata:
        name: example
      privateZone: example.local
      project: example-project
      provider: aws
      adminCIDRs:
      - x.x.x.x/32
      - y.y.y.y/24

Jenkins and bastion host
++++++++++++++++++++++++

The environment level can be overwritten for Jenkins and bastion host
by adding ``allowCIDRs`` in the instance pool block:

.. code-block:: yaml

  instancePools:
  - image: centos-puppet-agent
    allowCIDRs:
    - x.x.x.x/32
    maxCount: 1
    metadata:
      name: jenkins
    minCount: 1
    size: large
    type: jenkins

.. _secure-api-server:

API Server
++++++++++

For API server you can overwrite the environment level by adding ``allowCIDRs``
to the kubernetes block.

.. warning::
  For this to work, you need to set your `API Server public <user-guide.html#api-server>`__ first.

.. code-block:: yaml

  kubernetes:
    apiServer:
        public: true
        allowCIDRs:
        - y.y.y.y/24

Additional IAM policies
~~~~~~~~~~~~~~~~~~~~~~~

Additional IAM policies can be added by adding those ARNs to the ``tarmak.yaml``
config. You can add additional IAM policies to the ``cluster`` and
``instance pool`` blocks. When you define additional IAM policies on both
levels, they will be merged when applied to a specific instance pool.

Cluster
+++++++

You can add additional IAM policies that will be added to all the instance pools of
the whole cluster.

.. code-block:: yaml

    apiVersion: api.tarmak.io/v1alpha1
    clusters:
    - amazon:
        additionalIAMPolicies:
        - "arn:aws:iam::xxxxxxx:policy/policy_name"

Instance pool
+++++++++++++

It is possible to add extra policies to only a specific instance pool.

.. code-block:: yaml

  - image: centos-puppet-agent
    amazon:
      additionalIAMPolicies:
      - "arn:aws:iam::xxxxxxx:policy/policy_name"
    maxCount: 3
    metadata:
      name: worker
    minCount: 3
    size: medium
    subnets:
    - metadata:
      zone: eu-west-1a
    - metadata:
      zone: eu-west-1b
    - metadata:
      zone: eu-west-1c
    type: worker

Node Taints & Labels
~~~~~~~~~~~~~~~~~~~~

You might have added additional instance pools for a specific workload. In
these cases it might be useful to label and or taint the nodes in this instance
pool.

You add labels and taints in the tarmak yaml like this:

.. code-block:: yaml

  - image: centos-puppet-agent
    maxCount: 3
    metadata:
      name: worker
    minCount: 3
    size: medium
    type: worker
    labels:
    - key: "ssd"
      value: "true"
    taints:
    - key: "gpu"
      value: "gtx1170"
      effect: "NoSchedule"

**Note**, these are only applied when the node is first registered. Changes to
these values will not remove taints and labels from nodes that are already
registered.

API Server ELB Access Logs
~~~~~~~~~~~~~~~~~~~~~~~~~~

Tarmak features storing access logs of the internal and public, if enabled, API
server ELB. This is achieved through enabling configuration options in the
tarmak.yaml. You must specify at least the S3 bucket name with options to also
specify the bucket prefix and interval of 5 or 60 minutes. Interval defaults to
5 minutes.

.. code-block:: yaml

  kubernetes:
    apiServer:
      public: true
      amazon:
        internalELBAccessLogs:
          bucket: cluster-internal-accesslogs
        publicELBAccessLogs:
          bucket: cluster-public-accesslogs

Note that the S3 bucket needs to exist in the same region, with the correct S3
policy permissions. `Information on how to correctly set these permissions can
be found here
<https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-access-logs.html#attach-bucket-policy>`_.

Cluster Services
----------------

Grafana
~~~~~~~

Grafana is deployed as part of Tarmak. You can access Grafana through a
`Kubernetes cluster service <https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-services/>`_.
Do the following steps to access Grafana:

1. Create a proxy

.. code-block:: bash

    tarmak kubectl proxy

2. In the browser go to

.. code-block:: none

  http://127.0.0.1:8001/api/v1/namespaces/kube-system/services/monitoring-grafana/proxy/
