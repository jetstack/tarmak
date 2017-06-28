# Debugging hints

## Vault

### Consul status

```
```

### Vault status

```
```

## Etcd

## Kubernetes

## Calico Overlay

### Get calicoctl access

```
# create debug pod
cat <<EOF | kubectl create -f -
apiVersion: v1
kind: Pod
metadata:
  name: calico-ctl
  namespace: kube-system
spec:
  containers:
  - command:
    - sleep
    - "3600"
    env:
    - name: ETCD_ENDPOINTS
      valueFrom:
        configMapKeyRef:
          key: etcd_endpoints
          name: calico-config
    - name: ETCD_CA_CERT_FILE
      valueFrom:
        configMapKeyRef:
          key: etcd_ca
          name: calico-config
    - name: ETCD_KEY_FILE
      valueFrom:
        configMapKeyRef:
          key: etcd_key
          name: calico-config
    - name: ETCD_CERT_FILE
      valueFrom:
        configMapKeyRef:
          key: etcd_cert
          name: calico-config
    image: quay.io/calico/ctl:latest
    imagePullPolicy: Always
    name: calico-ctl
    resources:
      requests:
        cpu: 250m
    securityContext:
      privileged: true
    volumeMounts:
    - mountPath: /etc/etcd/ssl
      name: etcd-certs
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-gzvzf
      readOnly: true
  hostNetwork: true
  terminationGracePeriodSeconds: 1
  volumes:
  - hostPath:
      path: /etc/etcd/ssl
    name: etcd-certs
EOF

# run calicoctl
kubectl exec -t -i --namespace kube-system calico-ctl -- /calicoctl get ippools -o yaml
```

### Get etcdctl access

```
# create debug pod
cat <<EOF | kubectl create -f -
apiVersion: v1
kind: Pod
metadata:
  name: etcdctl
  namespace: kube-system
spec:
  containers:
  - command:
    - sh
    - -c
    - apk --update add curl && curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz && mkdir -p /tmp/test-etcd && tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /tmp/test-etcd --strip-components=1 && mv /tmp/test-etcd/etcdctl /usr/local/bin && sleep 3600
    env:
    - name: ETCD_VER
      value: v2.3.8
    - name: DOWNLOAD_URL
      value: https://github.com/coreos/etcd/releases/download
    - name: ETCDCTL_ENDPOINTS
      valueFrom:
        configMapKeyRef:
          key: etcd_endpoints
          name: calico-config
    - name: ETCDCTL_CA_FILE
      valueFrom:
        configMapKeyRef:
          key: etcd_ca
          name: calico-config
    - name: ETCDCTL_KEY_FILE
      valueFrom:
        configMapKeyRef:
          key: etcd_key
          name: calico-config
    - name: ETCDCTL_CERT_FILE
      valueFrom:
        configMapKeyRef:
          key: etcd_cert
          name: calico-config
    image: alpine:3.5
    imagePullPolicy: Always
    name: etcdctl
    resources:
      requests:
        cpu: 250m
    securityContext:
      privileged: true
    terminationMessagePath: /dev/termination-log
    volumeMounts:
    - mountPath: /etc/etcd/ssl
      name: etcd-certs
  terminationGracePeriodSeconds: 1
  hostNetwork: true
  volumes:
  - hostPath:
      path: /etc/etcd/ssl
    name: etcd-certs
EOF

# run etcdctl
kubectl exec -t -i --namespace kube-system etcdctl -- etcdctl ls --recursive
```


