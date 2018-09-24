Navigator
---------

Navigator is a Kubernetes extension for managing common stateful services. It
is fully compatible with clusters provisioned by Tarmak.

To get started with Navigator, we suggest following the project's
`quick-start guide  <https://navigator-dbaas.readthedocs.io/en/v0.1.0-alpha.1/quick-start.html>`_.

**Note: Storage Classes**

However, please note that the example manifests reference a ``storageClass``
called ``default``. See here in the example manifests:

- `Cassandra <https://github.com/jetstack/navigator/blob/v0.1.0-alpha.1/docs/quick-start/cassandra-cluster.yaml#L15>`_.
- `Elasticsearch <https://github.com/jetstack/navigator/blob/v0.1.0-alpha.1/docs/quick-start/es-cluster-demo.yaml#L39>`_.

Tarmak clusters do not have a ``storageClass`` with this name. Instead they have
``fast`` & ``slow`` (``fast`` is the default). Omitting the ``storageClass`` setting
will cause Navigator to use the default, in our case the ``fast`` class. You
may also choose to specify a different name for a storage class you defined
after creating the cluster.

Failure to update this will cause the scheduling of database pods to fail with
the following error: ``pod has unbound PersistentVolumeClaims``. Pods will remain
in ``Pending`` until the PVC can be bound.

**Note: Node Sizes for Elasticsearch**

The `Elasticsearch example
<https://github.com/jetstack/navigator/blob/v0.1.0-alpha.1/docs/quick-start/es-cluster-demo.yaml#L26-L54>`_
creates five pods each requesting 2Gi of RAM. You may wish to provision a
dedicated instance pool for this.
