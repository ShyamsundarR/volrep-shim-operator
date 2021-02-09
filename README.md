# VolumeReplication Operator Shim

VolumeReplication Operator is a **shim** for a sample VolumeReplication kubernetes CRD that,

- Uses the VolumeReplication [CRD](config/crd/bases/replication.storage.ramen.io_volumereplications.yaml) and manages its reconciliation
- The PVC that is managed is as per the `dataSource` in the VolumeReplication [CR](config/samples/replication_v1alpha1_volumereplication.yaml) and `dataSource` only handles PVCs at present
- The operator reconciles the VolumeReplication CR based on PVC health (i.e it should be bound to a PV), and reflects the following in `status`
    - `observedState` reflects the state observed at the generation in `observedGeneration`
    - `observedGeneration` reflects the generation of the most recently observed volume replication
    - `conditions`
        - Type: "Reconciled" denotes resource was reconciled
            - Status: "Complete" denotes reconciliation was completed
            - Status: "Error" denotes reconciliation had errors

NOTE: Currently the shim operator supports (or, is tested with) kubernetes v1.19 and above

## Build

The code is generated using the [operator-sdk](https://sdk.operatorframework.io/) and comes with standard SDK targets in the Makefile.

The most common way to build an image would be,

`$ make docker-build IMG=volrep-shim-operator:latest`

To push the image to a docker repository use,

`$ make docker-push docker-push IMG=volrep-shim-operator:latest`

## Deploy

Deploy to a kubernetes instance using,

`$ make deploy IMG=volrep-shim-operator:latest`

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

The sample VolumeReplication CR can be applied to the cluster running the operator as follows,

`$ kubectl apply -f config/samples/replication_v1alpha1_volumereplication.yaml`

**NOTE:** Sample above requires that a PVC named "sample-pvc" exists in the same namespace as the VolumeReplication resource

The reconcile logs can be viewed in parallel for the reconcile of the above created CR as follows,

`$ kubectl logs -fn volreplication-shim-system deployment.apps/volreplication-shim-controller-manager -c manager`

The status of reconciliation can be observed using,

`$ kubectl get volumereplication volumereplication-sample -o jsonpath='{.status}'`

Testing with [minikube](https://minikube.sigs.k8s.io/docs/) is the simplest, and a sample minikube PVC is present [here](config/samples/replication_v1alpha1_volumereplication.yaml).

If using podman to build and minikube with docker, then a helper script to copy the built image into minikube is present [here](hack/pushpodmantodocker_minikube.sh)
