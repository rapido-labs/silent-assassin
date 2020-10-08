# Silent-assassin

**Silent-Assassin(SA)** is a project built to solve the problems of using Preemptible Virtual Machines(PVM) in Production environment.
PVMs are unused VMs in GCP, that come at 1/4th of the cost of regular on demand VMs. While the cost part is sweet, they have two limitations.

1) **They last a maximum of 24 hours**
2) **GCP provides no availability guarantees**

Because of these limitations, GCP recommends using PVMs for short-lived jobs and fault-tolerant workloads. So what are the problems we will face if we use PVMs to serve our stateless micro services ?

1) **Potential large scale disruption**
2) **Unanticipated preemption**

SA solves the problem of mass deletion (Problem 1) by deleting the VMs randomly after 12 hours and before 24 hours of its creation, during non-business hours. It solves the 2nd problem, which is the unpredicted loss of pods due to early preemption, by triggering a drain through kubernetes in the event of a preemption.



## Installation
Installation of SA is fairly simple using [helm chart](helm-charts link).
SA would also use workload identity to access GCP and delete PVMs, so workload identintity should be enabled in the cluster.
You can refer the steps for enabling WLI in the cluster [here](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#enable_on_cluster).

### Authentication using workload identity
Create a service account and give the compute.instances.delete permissions.

```
$ export PROJECT_ID=<PROJECT>

$ export SERVICE_ACCOUNT=silent-assassin

$ gcloud iam --project=$PROJECT_ID service-accounts create $SERVICE_ACCOUNT \
    --display-name $SERVICE_ACCOUNT

$ gcloud iam --project=$PROJECT_ID roles create computeInstanceDelete \
    --project $PROJECT_ID \
    --title compute-instance-delete \
    --description "Delete compute instances" \
    --permissions compute.instances.delete

$ export $SERVICE_ACCOUNT_EMAIL=$(gcloud iam --project=$PROJECT_ID service-accounts list --filter $SERVICE_ACCOUNT --format 'value([email])')

$ gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:${$SERVICE_ACCOUNT_EMAIL} \
    --role=projects/${PROJECT_ID}/roles/computeInstanceDelete
```

After creation of service-account using above steps, we need to associate k8s service account that we use in the SA deployment with the GCP service account.

```
$ export NAMESPACE=<namespace_of_k8s_service_account>
$ export K8S_SERVICE_ACCOUNT=<.Release.Name> #This is the [service account](helm-chart link) used in SA deployment. You can set the name as the helm relese name
$ gcloud iam service-accounts add-iam-policy-binding   --role roles/iam.workloadIdentityUser   --member "serviceAccount:${PROJECT_ID}.svc.id.goog[${NAMESPACE}/$SERVICE_ACCOUNT]" $SERVICE_ACCOUNT@${PROJECT_ID}.iam.gserviceaccount.com --project=${PROJECT_ID}
```

Node-pools created after enabling WLI in the cluster can use WLI, but not the old nodepools. If you are planning to deploy SA in a node-pool that was created before WLI was enabled, migrate it to use WLI. See this [link](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) for more details.

```
export NODEPOOL_NAME=<nodepool-name>
export CLUSTER_NAME=<cluster-name>
export REGION=<region>
gcloud container node-pools update $NODEPOOL_NAME \
  --cluster=$CLUSTER_NAME \
  --workload-metadata=GKE_METADATA
  --region $REGION
 ```

### Installation using Helm.
```
helm install --name <Relese_Name> --namespace <namespace> ./helm-charts/silent-assassin
```

## Disaster Recovery

What if all preemptive nodes start misbehaving. There is a possibility that all preemptive compute instances might get preempted at once. How should we handle such scenario?

You could use the below steps in such cases.

1. Disable Auto Scaling in the preemptive node-pool
2. Cordon all nodes in the preemptive node-pool.
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


### Development

```
go build -o silent-assassin cmd/silent-assassin/*.go
```