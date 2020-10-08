# Disaster Recovery

What if all preemptive nodes start misbehaving. There is a possibility that all preemptive compute instances might get preempted at once. How should we handle such scenario?

You could use the below steps in such cases.

1. Disable Auto Scaling in the preemptive node-pool
2. Cordon all nodes in the preemptive node-pool.
3. Shift the pods from preemptive nodepool to the back up on demand nodepool. The shift can be perfromed by rolling restart of the deployments or manual restart of the pods.
4. Set number of nodes in the preemptible nodepool to 0 or delete the nodepool.

Below is the example for *services* nodepool.

We have two node pools for *services* workloads. ***services-p-1*** a preemptible nodepool and ***services-np-1*** an on-demand nodepool which acts as a backup nodepool.

1. Go to GCP console, edit **services-p-1** nodepool and disable auto scaling manually, currently gcloud commandline does not support for disabling autoscaling.

2. Cordon preemptive nodes.
```
    kubectl cordon -l cloud.google.com/gke-preemptible=true,component=services
```
3. Restart all pods running on preemptive nodes.
```
    kubectl rollout restart deployment istio-ingressgateway
```
4. Set number of nodes of  ***services-p-1*** to 0 or delete the nodepool.
