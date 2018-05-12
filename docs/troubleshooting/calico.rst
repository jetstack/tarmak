Calico
------

Calico as the overlay network policy plugin, plays a vital role in the cluster
communications.


Install calicoctl
*****************

The easiest form of accessing the calico manifest store is using calicoctl in
the controller pod:

::

  # exec into calico-controller container of the cluster
  kubectl exec -n kube-system -t -i $(kubectl get pods -n kube-system -l k8s-app=calico-policy -o go-template='{{range .items}}{{.metadata.name}}{{end}}') /bin/sh
  
  # download calicoctl
  apk --update add curl
  curl -O -L https://github.com/projectcalico/calicoctl/releases/download/v3.1.1/calicoctl
  chmod +x calicoctl 
  
  # request node objects
  ./calicoctl get nodes
  
  # leave container
  exit
