Horizontal Pod Autoscaling
--------------------------

This will give an example setup of `HPA <https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/>`_.
We are using `Prometheus <https://prometheus.io/>_` and the `Prometheus-adapter <https://github.com/DirectXMan12/k8s-prometheus-adapter>`_.

Prerequisite
~~~~~~~~~~~~

Make sure `HELM <https://www.helm.sh/>`_ is `activated <https://docs.tarmak.io/user-guide.html#tiller>`_ on the Tarmak cluster.
You also need to make sure you can connect to the cluster with your HELM install.

.. code-block:: bash

    helm version
    Client: &version.Version{SemVer:"v2.9.1", GitCommit:"20adb27c7c5868466912eebdf6664e7390ebe710", GitTreeState:"clean"}
    Server: &version.Version{SemVer:"v2.9.1", GitCommit:"20adb27c7c5868466912eebdf6664e7390ebe710", GitTreeState:"clean"}


Setup
~~~~~

Prometheus
++++++++++

You need Prometheus to scrape and store metrics from applications in its 
time series database. These metrics will be used by the HPA to decide if it 
has to scale the application. You can use an already running Prometheus in 
your environment or opt to set one up with the following steps.

.. warning::
   This will only setup a simple Prometheus. If you want to use Alertmanager and other
   more advanced options, take a look at the `kube-prometheus <https://github.com/coreos/prometheus-operator/tree/master/helm/kube-prometheus>`_ chart.


First activate the HELM repository of Prometheus

.. code-block:: bash

    helm repo add coreos https://s3-eu-west-1.amazonaws.com/coreos-charts/stable/


Now install the Prometheus operator which can be used to install and manage
Prometheus, Alertmanager and configure ServiceMonitors.

.. code-block:: bash

    helm install coreos/prometheus-operator --name prometheus-operator --namespace application-monitoring

After installing the Prometheus operator you can install Prometheus. To accomplish
this you use a yaml values file. This yaml file defines you want 2 Prometheus
pods running, keep data for 14 days and want to have a persistent volume
of 20 GB per Prometheus pod. You can tweak the Prometheus install further
by following the official `documentation <https://github.com/coreos/prometheus-operator/tree/master/helm/prometheus>`__.

.. code-block:: yaml

    replicaCount: 2 # You want HA
    retention: 336h # 14 days of retention
    storageSpec:
        volumeClaimTemplate:
            spec:
                class: gp2
                resources:
                    requests:
                        storage: 20Gi # Prefarably use bigger for iops (eg 100Gi)
    podAntiAffinity: hard # You want to be 100% sure you don't land on the same node with our prometheus instances
    serviceMonitorsSelector:
        matchExpressions:
            - key: prometheus
              operator: Exists

Save the yaml as a ``prometheus.yaml`` and run the following command:

.. code-block:: bash

    helm install coreos/prometheus --name prometheus-applications --namespace application-monitoring -f prometheus.yaml


Prometheus Adapter
++++++++++++++++++

You need to create a CA and SSL cert to validate your APIService with Kubernetes.

.. code-block:: bash

    export PURPOSE=server
    openssl req -x509 -sha256 -new -nodes -days 365 -newkey rsa:2048 -keyout ${PURPOSE}-ca.key -out ${PURPOSE}-ca.crt -subj "/CN=ca"
    echo '{"signing":{"default":{"expiry":"43800h","usages":["signing","key encipherment","'${PURPOSE}'"]}}}' > "${PURPOSE}-ca-config.json"

    export SERVICE_NAME=prometheus-adapter
    export ALT_NAMES='"prometheus-adapter.application-monitoring","prometheus-adapter.application-monitoring.svc"'
    echo '{"CN":"'${SERVICE_NAME}'","hosts":['${ALT_NAMES}'],"key":{"algo":"rsa","size":2048}}' | cfssl gencert -ca=server-ca.crt -ca-key=server-ca.key -config=server-ca-config.json - | cfssljson -bare apiserver


.. warning::
   Make sure the ``SERVICE_NAME`` and ``ALT_NAMES`` match your application release
   name and namespace where it is deployed.

Now create an ``prometheus-adapater.yaml`` with the following content:

.. code-block:: yaml

    tls:
        enable: true
        ca: |-
            <replace with content of server-ca.crt>
        key: |-
            <replace with content of apiserver-key.pem>
        certificate: |-
            <replace with content of apiserver.pem>

    # Change URL and port if you setup your own Prometheus server.
    prometheus:
        url: http://prometheus-applications.application-monitoring.svc
        port: 9090

    replicas: 2

Install the Prometheus Adapter:

.. code-block:: bash

    helm repo add incubator https://kubernetes-charts-incubator.storage.googleapis.com/
    helm install stable/prometheus-adapter --name prometheus-adapter --namespace application-monitoring -f prometheus-adapter.yaml


You can test if HPA works by running the following command against your 
Kubernetes cluster.

.. code-block:: bash

    kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1

    {"kind":"APIResourceList","apiVersion":"v1","groupVersion":"custom.metrics.k8s.io/v1beta1","resources":[]}

Usage
~~~~~

To start scaling based on custom-metrics, you need to have an application 
or Prometheus exporter that exposes metrics in Prometheus format. Another
requisite is to have an Kubernetes endpoint for your application. That 
endpoint will be used to discover your Pods. If your application meets 
these  requirements, you can add a ``ServiceMonitor`` to start monitoring
your application with Prometheus.

.. code-block:: yaml

    apiVersion: monitoring.coreos.com/v1
    kind: ServiceMonitor
    metadata:
        name: <example>
        namespace: application-monitoring
    labels:
        prometheus: prometheus-applications
    spec:
        endpoints:
        - interval: 30s
          targetport: <port>
          path: /metrics
        namespaceSelector:
            matchNames:
            - <application_namespace>
        selector:
            matchLabels:
                <key>: <value that matches your application>

When adding the ``ServiceMonitor``, make sure to keep ``prometheus`` as an key
in labels, that is how Prometheus discovers the different ServiceMonitors.
The ``ServiceMonitor`` has to be deployed in the same namespace as your Prometheus.

After applying the ``ServiceMonitor``, Prometheus should start discovering
all your application pods and start to monitor them.

You can find the correct metric by querying the custom.metrics API endpoint.

.. code-block:: bash

    kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1 | jq
    kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/<aplication_namespace>/pods/*/<metric_name>" | jq
    kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/<aplication_namespace>/service/<service_name>/<metric_name> | jq

When you found the correct metric to scale on, you can create your 
``HorizontalPodAutoscaler``.

.. code-block:: yaml

    kind: HorizontalPodAutoscaler
    apiVersion: autoscaling/v2beta1
    metadata:
        name: <example>
        namespace: <example_namespace>
    spec:
        scaleTargetRef:
            apiVersion: apps/v1beta2
            kind: Deployment
            name: example
        minReplicas: 2
        maxReplicas: 4
        metrics:
        - type: Pods
          pods:
            metricName: <metric_name>
            targetAverageValue: <metric_value>

Watch the horizontal pod autoscaler:

.. code-block:: bash
    
    kubectl describe hpa example


More examples can be found in the kubernetes `documentation <https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/>`__.

.. warning::
    Certainly take a look a types ``Object`` and ``Pods`` for HPA based on custom-metrics.

A fully worked out example with a example application that has metrics can be found in 
`luxas repo <https://github.com/luxas/kubeadm-workshop/blob/master/demos/monitoring/sample-metrics-app.yaml#L51>`__.
