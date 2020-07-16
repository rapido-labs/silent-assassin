# silent-assassin

Preemptible Node Killer.

### Development

* Build

```
skaffold -v trace build -p local
```

### Steps to create Google service account
In order to have the silent-assassin delete nodes, reate a service account and give the compute.instances.delete permissions.

You can either create the service account and associate the role using the GCloud web console or the cli:
```
$ export project_id=<PROJECT>
$ gcloud iam --project=$project_id service-accounts create silent-assassin \
    --display-name silent-assassin
$ gcloud iam --project=$project_id roles create computeInstanceDelete \
    --project $project_id \
    --title compute-instance-delete \
    --description "Delete compute instances" \
    --permissions compute.instances.delete
$ export service_account_email=$(gcloud iam --project=$project_id service-accounts list --filter silent-assassin --format 'value([email])')

$ gcloud projects add-iam-policy-binding $project_id \
    --member=serviceAccount:${service_account_email} \
    --role=projects/${project_id}/roles/computeInstanceDelete

#Below step can only be executed when the service account in created in GKE.
export namespace=<namespace_of_k8s_service_account>
$ gcloud iam service-accounts add-iam-policy-binding   --role roles/iam.workloadIdentityUser   --member "serviceAccount:${project_id}.svc.id.goog[${namespace}/silent-assassin]" silent-assassin@${project_id}.iam.gserviceaccount.com --project=${project_id}
```

## Migrate node-pool to use Workload Identity(WI)
```
gcloud container node-pools update <nodepool-name> \
  --cluster=<cluster-name> \
  --workload-metadata=GKE_METADATA
  --region <region>
 ```
 - This is only required if we are running SA in an existing node-pool.
 - New node-pools on cluster with WI already enabled will have this set by default.
