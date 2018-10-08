## etcd-ansible-operator

This operator implements the [etcd-operator](https://github.com/coreos/etcd-operator/) using ansible built on top of [ansible-operator](https://github.com/water-hole/ansible-operator).

__***Note:***__ This is a work in progress. Features like backup and restore will be added soon. This operator uses a different version of base image to support `failover` feature.

### Pre-requisites:

Some sort of kubernetes cluster deployed with `kubectl` correctly configured. [Minikube](https://github.com/kubernetes/minikube/) is the easiest way to get started.

### Steps to bring an etcd cluster up

1. Create RBAC `kubectl create -f https://raw.githubusercontent.com/water-hole/etcd-ansible-operator/master/deploy/rbac.yaml`
2. Create CRD `kubectl create -f https://raw.githubusercontent.com/water-hole/etcd-ansible-operator/master/deploy/crd.yaml`
3. Deploy the operator `kubectl create -f https://raw.githubusercontent.com/water-hole/etcd-ansible-operator/master/deploy/operator.yaml`
4. Create an etcd cluster `kubectl create -f https://raw.githubusercontent.com/water-hole/etcd-ansible-operator/master/deploy/cr.yaml`
5. Verify that cluster is up by `kubectl get pods -l app=etcd`. You should see something like this
    ```
    $ kubectl get pods -l app=etcd
    NAME                              READY     STATUS    RESTARTS   AGE
    example-etcd-cluster-1a7d2c2f8b   1/1       Running   0          14m
    example-etcd-cluster-5afd8f00ce   1/1       Running   0          14m
    example-etcd-cluster-e43636bc7c   1/1       Running   0          14m
    ```

### Accessing the etcd cluster

If you are using minikube:

1. Create a service to access etcd cluster from outside the cluster by `kubectl create -f https://raw.githubusercontent.com/coreos/etcd-operator/master/example/example-etcd-cluster-nodeport-service.json`
2. Install [etcdctl](https://coreos.com/etcd/docs/latest/getting-started-with-etcd.html)
3. Set etcd version `export ETCDCTL_API=3`
4. Set etcd endpoint `export ETCDCTL_ENDPOINTS=$(minikube service example-etcd-cluster-client-service --url)`
5. Set a key in etcd `etcdctl put hello world`

If you are inside the cluster, set the etcd endpoint to: `http://<cluster-name>-client.<namespace>.svc:2379` and it should work. If you are using secure client, use `https` protocol for the endpoint.

### Delete a cluster
1. Bring a cluster up.
2. Delete the cluster by `kubectl delete etcdcluster example-etcd-cluster`. This should delete all the pods and services created because of this cluster

### Scale cluster up

1. Bring a cluster up as discussed above
2. Edit the example cr just created  with command 
`kubectl edit etcdcluster example-etcd-cluster`. Change the size from `3` to `5`

    ```
    apiVersion: "etcd.database.coreos.com/v1beta2"
    kind: "EtcdCluster"
    metadata:
      name: "example-etcd-cluster"
    spec:
      size: 5
      version: "3.2.13"
    ```
   This should scale up the cluster by 2 pods.
  
3. Verify that the cluster has scaled up by `$kubectl get pods -l app=etcd`. You should see something like this:
    ```
    $ kubectl get pods -l app=etcd
    NAME                              READY     STATUS    RESTARTS   AGE
    example-etcd-cluster-1a7d2c2f8b   1/1       Running   0          18m
    example-etcd-cluster-1c497c44c5   1/1       Running   0          29s
    example-etcd-cluster-5afd8f00ce   1/1       Running   0          18m
    example-etcd-cluster-a3f3b02a1b   1/1       Running   0          18s
    example-etcd-cluster-e43636bc7c   1/1       Running   0          18m
    ```

### Check failure recovery
1. Bring a cluster up.
2. Delete a pod to simulate a failure `$kubectl delete pod example-etcd-cluster-1a7d2c2f8b`
3. Within sometime, you should see the deleted pod going away and being replaced by a new pod, something like this:
    
    ```$ kubectl get pods -l app=etcd
       NAME                              READY     STATUS    RESTARTS   AGE
       example-etcd-cluster-1c497c44c5   1/1       Running   0          3m
       example-etcd-cluster-25f6bd225a   1/1       Running   0          8s
       example-etcd-cluster-5afd8f00ce   1/1       Running   0          21m
       example-etcd-cluster-a3f3b02a1b   1/1       Running   0          3m
       example-etcd-cluster-e43636bc7c   1/1       Running   0          21m   
   ```
       
### TLS

To create certificates, do the following:
1. Bring [minikube](https://github.com/kubernetes/minikube/) up in your host.
2. Install [ansible] (https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html#running-from-source). Note the ansible version should be greater than 2.6
3. Run `tls_playbook.yaml` as follows:

    ```
        ansible-playbook ansible/tls_playbook.yaml
    ```
   This should create certs in `/tmp/etcd/etcdtls/example-etcd-cluster/` directory. This should also create 3 kubernetes secrets
4. Verify by running:
    ```
    $ kubectl get secrets
    NAME                  TYPE                                  DATA      AGE
    default-token-zhqgh   kubernetes.io/service-account-token   3         1d
    etcd-client-tls       Opaque                                3         3h
    etcd-peer-tls         Opaque                                3         3h
    etcd-server-tls       Opaque                                3         3h
    ```
5. Create rbac if not already created `$kubectl create -f https://raw.githubusercontent.com/water-hole/etcd-ansible-operator/master/deploy/rbac.yaml`
6. Create crd if not already created `$kubectl create -f https://raw.githubusercontent.com/water-hole/etcd-ansible-operator/master/deploy/crd.yaml`
7. Deploy operator if nor already deployed `$kubectl create -f https://raw.githubusercontent.com/water-hole/etcd-ansible-operator/master/deploy/operator.yaml`
8. Create etcd cluster with tls using:
    ```
    kubectl create -f https://raw.githubusercontent.com/water-hole/etcd-ansible-operator/master/deploy/cr_tls.yaml
    ```


### Upgrades

The operator supports version upgrades for etcd. Steps to try it:

1. Bring up an etcd cluster as discussed above.
2. Check the version of the images
    ```
    $ kubectl get pods -l app=etcd -o=jsonpath='{range .items[*]}{"\n"}{.metadata.name}{":\t"}{range .spec.containers[*]}{.image}{", "}{end}{end}' |sort
       
       example-etcd-cluster-1d139522e2:        quay.io/coreos/etcd:v3.2.13,
       example-etcd-cluster-7e9909fce8:        quay.io/coreos/etcd:v3.2.13,
       example-etcd-cluster-bb0a9b3ec8:        quay.io/coreos/etcd:v3.2.13,
   ```
3. Run the command: `$kubectl apply -f https://raw.githubusercontent.com/water-hole/etcd-ansible-operator/master/deploy/update_cr.yaml`
4. Verify with the following command 
    ```
    $ kubectl get pods -l app=etcd -o=jsonpath='{range .items[*]}{"\n"}{.metadata.name}{":\t"}{range .spec.containers[*]}{.image}{", "}{end}{end}' |sort
    
    example-etcd-cluster-1d139522e2:        quay.io/coreos/etcd:v3.3,
    example-etcd-cluster-7e9909fce8:        quay.io/coreos/etcd:v3.3,
    example-etcd-cluster-bb0a9b3ec8:        quay.io/coreos/etcd:v3.3,
    
    ```
