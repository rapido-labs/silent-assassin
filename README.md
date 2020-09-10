# silent-assassin

Preemptible Node Killer.

### Development

```
go build -o silent-assassin cmd/silent-assassin/*.go
```

### GCP

* #### Authentication

    * Using [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)

In order to have the silent-assassin delete nodes, create a service account and give the compute.instances.delete permissions.

You can either create the service account and associate the role using the GCloud web console or the cli:

```
$ export project_id=<PROJECT>
```
```
$ gcloud iam --project=$project_id service-accounts create silent-assassin \
    --display-name silent-assassin
```
```
$ gcloud iam --project=$project_id roles create computeInstanceDelete \
    --project $project_id \
    --title compute-instance-delete \
    --description "Delete compute instances" \
    --permissions compute.instances.delete
```
```
$ export service_account_email=$(gcloud iam --project=$project_id service-accounts list --filter silent-assassin --format 'value([email])')
```
```
$ gcloud projects add-iam-policy-binding $project_id \
    --member=serviceAccount:${service_account_email} \
    --role=projects/${project_id}/roles/computeInstanceDelete
```

**Below step can only be executed after the service account is created in GKE.**

```
$ export namespace=<namespace_of_k8s_service_account>

$ gcloud iam service-accounts add-iam-policy-binding   --role roles/iam.workloadIdentityUser   --member "serviceAccount:${project_id}.svc.id.goog[${namespace}/silent-assassin]" silent-assassin@${project_id}.iam.gserviceaccount.com --project=${project_id}
```

**Migrating an existing node-pool to use Workload Identity. See [link](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) for more details.**

```
gcloud container node-pools update <nodepool-name> \
  --cluster=<cluster-name> \
  --workload-metadata=GKE_METADATA
  --region <region>
 ```

 - This is only required if we are running SA in an existing node-pool.
 - New node-pools on cluster with WI enabled will have this set by default.

## Disaster Recovery

What if all preemptive nodes start misbehaving. There is a possibility that all preemptive compute instances might get preempted at once. How should we handle such scenario?

You could use the below steps in such cases.

1. Disable Auto Scaling in the preemptive nodePool
2. Cordon all nodes in the preemptive nodepool.
3. Shift the pods from preemptive nodepool to the back up on demand nodepool. The shift can be perfromed by rolling restart of the deployments or manual restart of the pods.
4. Set number of nodes in the preemptible nodepool to 0 or delete the nodepool.

Below is the example for *istio-gateway* nodepool.

We have two node pools for *istio-gateway* workloads. ***istio-gateway-p-1*** a preemptible nodepool and ***istio-gateway-2*** an on-demand nodepool which acts as a backup nodepool.

1. Go to GCP console, edit **istio-gateway-p-1** nodepool and disable auto scaling manually, currently gcloud commandline does not support for disabling autoscaling.

2. Cordon preemptive nodes.
```
    kubectl cordon -l cloud.google.com/gke-preemptible=true,component=istio-gateway
```
3. Restart all pods running on preemptive nodes.
```
    kubectl rollout restart deployment istio-ingressgateway
```
4. Set number of nodes of  ***istio-gateway-p-1*** to 0 or delete the nodepool.

