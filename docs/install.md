## Installation
Installation of SA is fairly simple using [helm chart](helm-charts link).

Inorder to delete VMs, SA would need access to service account that has access to delete VMs. We can achive this in two ways.
1. Mounting service account file in pod
2. Using Workload identity in the cluster

### Authentication by mounting key file.

Create a service account, assign **container.admin** and  **compute.admin** roles and create a service account key.

```
$ export PROJECT_ID=<PROJECT>

$ export SERVICE_ACCOUNT=silent-assassin

$ gcloud iam --project=$PROJECT_ID service-accounts create $SERVICE_ACCOUNT \
    --display-name $SERVICE_ACCOUNT

$ export SERVICE_ACCOUNT_EMAIL=$(gcloud iam --project=$PROJECT_ID service-accounts list --filter $SERVICE_ACCOUNT --format 'value([email])')

$ gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:${SERVICE_ACCOUNT_EMAIL} \
    --role=roles/compute.admin
$ gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member=serviceAccount:${SERVICE_ACCOUNT_EMAIL} \
    --role=roles/container.admin

gcloud iam service-accounts keys create key.json \
  --iam-account ${$SERVICE_ACCOUNT_EMAIL}

```
Keep the `key.json` file safely. We have to use the content of the file as value for parameter `secret.googleServiceAccountKeyfileJson` in helm chart.

### Authentication using workload identity
Enable workload identity in the cluster.
You can refer the steps for enabling WLI in the cluster [here](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#enable_on_cluster).

Create a service account and give  **compute.instances.delete** permission.

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

After creation of the service-account using above steps, we need to associate k8s service account that we use in the SA deployment with the GCP service account.

```
$ export NAMESPACE=<namespace_of_k8s_service_account>
$ export K8S_SERVICE_ACCOUNT=<.Release.Name> #This is the service account(helm-charts/templates/service-account.yaml) used in SA deployment. You can set the name as the helm release name
$ gcloud iam service-accounts add-iam-policy-binding   --role roles/iam.workloadIdentityUser   --member "serviceAccount:${PROJECT_ID}.svc.id.goog[${NAMESPACE}/$SERVICE_ACCOUNT]" $SERVICE_ACCOUNT@${PROJECT_ID}.iam.gserviceaccount.com --project=${PROJECT_ID}
```

Node-pools created after enabling WLI in the cluster can use WLI, but not the old node-pools. If you are planning to deploy SA in a node-pool that was created before WLI was enabled, migrate it to use WLI. See this [link](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) for more details.

```
export NODEPOOL_NAME=<nodepool-name>
export CLUSTER_NAME=<cluster-name>
export REGION=<region>
gcloud container node-pools update $NODEPOOL_NAME \
  --cluster=$CLUSTER_NAME \
  --workload-metadata=GKE_METADATA \
  --region $REGION
 ```

Use the email id as the value for parameter `workloadIdentityServiceAccount.email` in helm chart.

### Installation using Helm.
Follow this [link](../helm-charts/silent-assassin/) for details on installation using helm.