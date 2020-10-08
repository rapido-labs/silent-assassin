# Node-Pools and Affinity

## Nodepools in the cluster
A node-pool is a group of nodes within a cluster that all have the same configuration.
As PVMs are available from a finite amount of Compute Engine resources, and might not always be available,so create two node-pools for a class of workloads. One is a node-pool with Preemptible VMs and another with on-demand VMs acting as a fallback node-pool in case of unavailability of Preemptible resources.

For example,for backend service workloads,create two node-pools.
1) services-np, node-pool with on-demand VMs
2) services-p, node-pool with PVMs

Add kubernetes label component:services to both node-pools.

## Affinity in Deployments
We can constrain a Pod to only be able to run on particular nodes, or to prefer to run on particular nodes by setting affinity in the deployments.
Below are the affinity and pod anti-affinity you should set such that pods will spread in all zones.
Also add soft-affinity to select PVMs over on-demand VMs. Below is the example affinity.

```
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: component
          operator: In
          values:
          - "services"
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 5
      preference:
        matchExpressions:
        - key: cloud.google.com/gke-preemptible
          operator: In
          values:
          - "true"
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: app
          operator: In
          values:
          - <app-name>
      topologyKey: "kubernetes.io/hostname"
```