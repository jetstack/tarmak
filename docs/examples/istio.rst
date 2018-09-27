Istio
-----

`Istio <https://istio.io>`_ is a service mesh that lets you connect, secure,
control, and observe services. Istio can be deployed on Tarmak Kubernetes
clusters.

However, if you have enabled Tarmak's default PodSecurityPolicy (see `User Guide
</user-guide.html#pod-security-policy>`_), then privileged sidecar containers
injected by Istio will be blocked. Events such as this will show on the
ReplicaSet:

.. code::

    Warning  FailedCreate  3m               replicaset-controller  Error creating: pods "details-v1-6865b9b99d-rm26k" is forbidden: unable to validate against any pod security policy: [capabilities.add: Invalid value: "NET_ADMIN": capability may not be added]

More details about the access requirements of the Istio containers can be found
`here <https://github.com/istio/old_issues_repo/issues/172>`_

For now, we recommend only enabling use of the ``psp:privileged`` in the
namespaces containing Istio-managed workloads - rather than allowing it across
the entire cluster.

To enable ``psp:privileged`` in a single namespace (called ``foobar`` in our
example), apply the following RoleBinding in that namespace.

.. code:: yaml

   apiVersion: rbac.authorization.k8s.io/v1
   kind: RoleBinding
   metadata:
    name: default:privileged
    namespace: foobar
   roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: psp:privileged
   subjects:
   - apiGroup: rbac.authorization.k8s.io
    kind: Group
    name: system:serviceaccounts:foobar
