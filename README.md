# VolumeReplication Operator Shim

VolumeReplication Operator shim is a Ceph-RBD specific VolumeReplication kubernetes CRD operator that,

- Uses the VolumeReplication [CRD](config/crd/bases/replication.storage.ramen.io_volumereplications.yaml) to manage ceph-rbd [mirroring](https://docs.ceph.com/en/latest/rbd/rbd-mirroring/)
    - The ceph-rbd image that is managed is as per the `dataSource` in the VolumeReplication [CR](config/samples/replication_v1alpha1_volumereplication.yaml) and only handles PVCs as the `dataSource` at present
    - The operator reconciles the VolumeReplication CR to enable/promote/demote/force-promote/resync ceph-rbd images as desired by the `state` in the CR

NOTE: Currently the operator supports (or, is tested with) Ceph Octopus version (v15.y.z), and kubernetes v1.19 and above

## Build

The code is generated using the [operator-sdk](https://sdk.operatorframework.io/) and comes with standard SDK targets in the Makefile.

The most common way to build an image would be `$ make docker-build IMG=volrep-shim-operator:latest`

To push the image to a docker repository use `$ make docker-push docker-push IMG=volrep-shim-operator:latest`

## Deploy

Deploy to a kubernetes instance using `$ make deploy IMG=volrep-shim-operator:latest`

The artifacts are deployed in the `volreplication-shim-system` namespace and are,

```
$ kubectl get all -n volreplication-shim-system
NAME                                                          READY   STATUS    RESTARTS   AGE
pod/volreplication-shim-controller-manager-744cb9dc78-47ptb   2/2     Running   0          3m56s

NAME                                                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/volreplication-shim-controller-manager-metrics-service   ClusterIP   10.104.17.255   <none>        8443/TCP   3m56s

NAME                                                     READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/volreplication-shim-controller-manager   1/1     1            1           3m56s

NAME                                                                DESIRED   CURRENT   READY   AGE
replicaset.apps/volreplication-shim-controller-manager-744cb9dc78   1         1         1       3m56s
```

## Test

NOTE: At some point you would need 2 kubernetes clusters with Ceph managed by Rook to play with this operator and one such example to create such an environment using [minikube](https://minikube.sigs.k8s.io/docs/) is presented [here](https://www.mrajanna.com/setup-rbd-async-mirroring-with-rook/)

The sample VolumeReplication CR can be applied to the cluster running the operator as follows,
`$ kubectl apply -f config/samples/replication_v1alpha1_volumereplication.yaml`

The reconcile logs can be viewed in parallel for the reconcile of the above created CR as follows,
`$ kubectl logs -fn volreplication-shim-system deployment.apps/volreplication-shim-controller-manager -c manager`
