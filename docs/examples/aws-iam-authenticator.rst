AWS IAM Authenticator
---------------------

`AWS IAM Authenticator <https://github.com/kubernetes-sigs/aws-iam-authenticator>`_ is a daemon that lets you authenticate to the 
Kubernetes RBAC system via Amazon Web Services - Identity and Access Management users and roles

You can initialise the cluster to use this with the following configuration snippet in tarmak.yaml:

.. code-block:: yaml

    ...
    kubernetes:
      apiServer:
        amazon:
          awsIAMAuthenticatorInit: true
    ...

You can configure the IAM authenticator server with the following config map and daemonset, 
replacing ``000000000000`` with your AWS account ID and ``your-tarmak-cluster`` with your cluster name, 
including the ``-cluster`` suffix in a single cluster environment:

.. code-block:: yaml

    apiVersion: v1
    kind: ConfigMap
    metadata:
      namespace: kube-system
      name: aws-iam-authenticator
      labels:
        k8s-app: aws-iam-authenticator
    data:
      config.yaml: |
        # a unique-per-cluster identifier to prevent replay attacks
        # (good choices are a random token or a domain name that will be unique to your cluster)
        clusterID: your-tarmak-cluster
        server:
          mapRoles:
          # statically map arn:aws:iam::<your account id>:role/KubernetesAdmin to a cluster admin
          - roleARN: arn:aws:iam::000000000000:role/KubernetesAdmin
            username: kubernetes-admin
            groups:
            - system:masters
    ---
    apiVersion: extensions/v1beta1
    kind: DaemonSet
    metadata:
      namespace: kube-system
      name: aws-iam-authenticator
      labels:
        k8s-app: aws-iam-authenticator
    spec:
      updateStrategy:
        type: RollingUpdate
      template:
        metadata:
          annotations:
            scheduler.alpha.kubernetes.io/critical-pod: ""
          labels:
            k8s-app: aws-iam-authenticator
        spec:
          # run on the host network (don't depend on CNI)
          hostNetwork: true
          # run on each master node
          nodeSelector:
            node-role.kubernetes.io/master: ""
          tolerations:
          - effect: NoSchedule
            key: node-role.kubernetes.io/master
          - key: CriticalAddonsOnly
            operator: Exists
          containers:
          - name: aws-iam-authenticator
            image: gcr.io/heptio-images/authenticator:v0.3.0
            args:
            - server
            - --config=/etc/aws-iam-authenticator/config.yaml
            - --state-dir=/var/aws-iam-authenticator
            - --generate-kubeconfig=/etc/kubernetes/aws-iam-authenticator/kubeconfig.yaml
            - --kubeconfig-pregenerated=true
            resources:
              requests:
                memory: 20Mi
                cpu: 10m
              limits:
                memory: 20Mi
                cpu: 100m
        securityContext:
          privileged: true
        volumeMounts:
        - name: config
          mountPath: /etc/aws-iam-authenticator/
        - name: state
          mountPath: /var/aws-iam-authenticator/
      securityContext:
        fsGroup: 0
        runAsUser: 0
      volumes:
      - name: config
        configMap:
          name: aws-iam-authenticator
      - name: state
        hostPath:
          path: /var/aws-iam-authenticator/

You can then authenticate to the cluster with e.g. the following, as long as aws-iam-authenticator is 
downloaded and on your path:

.. code-block:: yaml

    apiVersion: v1
    clusters:
    - cluster:
        certificate-authority-data: <snip - get these from ~/.tarmak/your-cluster/kubeconfig>
        server: https://api.your-cluster.somedomain.io ##see above
      name: your-cluster
    contexts:
    - context:
        cluster: your-cluster
        namespace: default
        user: your-cluster
      name: your-cluster
    users:
    - name: your-cluster
      user:
        exec:
          apiVersion: client.authentication.k8s.io/v1alpha1
          args:
          - token
          - -i
          - your-cluster ##change me
          - -r
          - arn:aws:iam::000000000000:role/KubernetesAdmin  ##change me
          command: aws-iam-authenticator-aws
          env:
          - name: AWS_PROFILE
            value: your_profile ##change or remove me
