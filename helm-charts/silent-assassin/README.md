# Helm Chart
This chart bootstraps a SA deployment and informer deamonset on a Kubernetes cluster using the Helm package manager.

## Prerequisites
- GKE 1.14.10+
- Helm 2.15.2+ or Helm 3.0

## Installing the Chart

SA configuration is also part of helm chart values. Check the [table](#parameters) below for the values and their description.

`affinity` is required to make sure that the pod gets deployed in an on-demand nodepool.

`workloadIdentityServiceAccount.email` is required if Workload Identity is enabled.

`secret.googleServiceAccountKeyfileJson` is required if Workload Identity is not enabled and you are mounting service account key file in pod.

```
$ helm install my-release ./
```

The command deploys silent-assassin server and deamonset on the Kubernetes cluster in the default configuration. The Parameters section lists the parameters that can be configured during installation.

## Uninstalling the Chart
To uninstall/delete the my-release deployment:
```
$ helm delete --purge my-release
```
The command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters
The following table lists the configurable parameters of the Silent-assassin chart and their default values.
We have shortned `silent-assassin` to `sa` in last few rows as the long name was overflowing the column.

| Parameter                                              | Description                                                   |  Default                                   |
|--------------------------------------------------------|---------------------------------------------------------------|--------------------------------------------|
| `replicaCount`                                         | replicas of SA server.1 by default.(Dont change)              | `1`                                        |
| `revisionHistoryLimit`                                 | Nuber of resplicaset revision history                         | `4`                                        |
| `containerPort`                                        | Port on which SA server runs                                  | `8080`                                     |
| `imageConfig.pullPolicy`                               | Image Pull policy                                             | `Always`                                   |
| `imageConfig.image`                                    | Image used for SA deployment and deamonset                    | `rapidolabs/silent-assassin:latest`        |
| `resources.enabled`                                    | Enable CPU/Memory resource requests/limits                    | `true`                                     |
| `resources.requests.cpu`                               | CPU resource request                                          | `20m`                                      |
| `resources.requests.memory`                            | Memory resource request                                       | `20Mi`                                     |
| `resources.limits.cpu`                                 | CPU resource limit                                            | `100m`                                     |
| `resources.limits.memory`                              | Memory resource limit                                         | `100Mi`                                    |
| `daemonset.resources.enabled`                          | Enable CPU/Memory resource requests/limits                    | `true`                                     |
| `daemonset.resources.requests.cpu`                     | CPU resource request                                          | `10m`                                      |
| `daemonset.resources.requests.memory`                  | Memory resource request                                       | `10Mi`                                     |
| `daemonset.resources.limits.cpu`                       | CPU resource limit                                            | `20m`                                      |
| `daemonset.resources.limits.memory`                    | Memory resource limit                                         | `20Mi`                                     |
| `service.type`                                         | Service Type                                                  | `ClusterIP`                                |
| `service.servicePort`                                  | Service Port                                                  | `80`                                       |
| `workloadIdentityServiceAccount.enabled`               | Workload identity is enabled in cluster                       | `true`                                     |
| `workloadIdentityServiceAccount.email`                 | Email of GCP SA created for for WLI (required)                | ``                                         |
| `secret.valuesAreBase64Encoded`                        | Encode contents of secret                                     | `false`                                    |
| `secret.googleServiceAccountKeyfileJson`               | Content of GCP service account key file without new lines     | `{"type":"service_account","project_id".}` |
| `affinity`                                             | Map of node/pod affinities                                    | `{}`                                       |
| `silent_assassin.node_selectors`                       | node selectors for which sa should act                        | `cloud.google.com/gke-preemptible=true`    |
| `silent_assassin.logger_level`                         | logging level of SA (debug|info|warn|error)                   | `warn`                                     |
| `silent_assassin.k8s_run_mode`                         | SA run mode (InCluster|OutCluster)                            | `InCluster`                                |
| `silent_assassin.spotter.poll_interval`                | Spotter polling interval                                      | `1s`                                       |
| `sa.spotter.white_list_interval_hours`                 | Interval for node kills                                       | `"06:30-08:30,18:30-00:30"`                |
| `sa.killer.poll_interval`                              | Killer Poll interval                                          | `1s`                                       |
| `sak.draining_timeout_when_node_expired`               | timeout for drain when node expired                           | `5m`                                       |
| `sak.draining_timeout_when_node_preempted`             | timeout for drain when node preempted                         |                                            |
| `sa.slack.webhook_url`                                 | Slack webhook URL                                             | ``                                         |
| `sa.slack.username`                                    | Username for Slack messages                                   | `SILENT-ASSASSIN`                          |
| `sa.slack.channel`                                     | slack channel name                                            | ``                                         |
| `sa.slack.icon_url`                                    | slack icon url                                                | ``                                         |
| `sa.client.server_retries`                             | client side retries for server in case of preemption          | `4`                                        |
| `sa.watch_maintainance_event`                          | watch for maintaintainance events along with preemption       | `false`                                    |
