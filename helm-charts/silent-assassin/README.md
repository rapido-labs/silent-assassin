# Helm Chart for Rapido silent-assassin in Kubernetes

This project provides Helm Chart for deploying silent-assassin application into any Kubernetes based cloud.

The templates require your application to built into a Docker image. 

This project provides the following files:

| File                                              | Description                                                        |
|---------------------------------------------------|-----------------------------------------------------------------------|  
| `Chart.yaml`                    | The definition file for your application                           | 
| `values.yaml`                   | Configurable values that are inserted into the following template files      | 
| `templates/deployment.yaml`     | Template to configure your application deployment.                 | 
| `templates/service.yaml`        | Template to configure your application service.                 | 
| `templates/hpa.yaml`            | Template to configure your application autoscaler.                 | 
| `templates/istio-virtual-service.yaml`          | Template to configure your application virtual service.                 | 
| `templates/NOTES.txt`           | Helper to enable locating your application IP and PORT        | 

In order to use these template files, copy the files from this project into your application directory. You should only need to edit the `values.yaml` files.

## Prerequisites

Using the template Helm charts assumes the following pre-requisites are complete:  

1. You have a Kubernetes cluster  
  This could be one hosted by a cloud provider or running locally, for example using [Minikube](https://kubernetes.io/docs/setup/minikube/)
  
2. You have kubectl installed and configured for your cluster  
  The [Kuberenetes command line](https://kubernetes.io/docs/tasks/tools/install-kubectl/) tool, `kubectl`, is used to view and control your Kubernetes cluster. 

3. You have the Helm command line and Tiller backend installed  
   [Helm and Tiller](https://docs.helm.sh/using_helm/) provide the command line tool and backend service for deploying your application using the Helm chart. 
   
4. You have created and published a Docker image for your application  

5. Your application has a "health" endpoint  
  This allows Kubernetes to restart your application if it fails or becomes unresponsive.


## Configuring the Chart for your Application

The following table lists the configurable parameters of the template Helm chart and their default values.

| Parameter                  | Description                                     | Default                                                    |
| -----------------------    | ---------------------------------------------   | ---------------------------------------------------------- |
| `image.repository`         | image repository                                | `asia.gcr.io/obelus-x1/silent-assassin`                                 |
| `image.tag`                | Image tag                                       | `latest`                                                    |
| `image.pullPolicy`         | Image pull policy                               | `Always`                                                   |
| `livenessProbe.initialDelaySeconds`   | How long to wait before beginning the checks our pod(s) are up |   30                             |
| `livenessProbe.periodSeconds`         | The interval at which we'll check if a pod is running OK before being restarted     | 10          |
| `service.name`             | Kubernetes service name                                | `Node`                                                     |
| `service.type`             | Kubernetes service type exposing port                  | `NodePort`                                                 |
| `service.port`             | TCP Port for this service                       | 3000                                                       |
| `resources.limits.memory`  | Memory resource limits                          | `128m`                                                     |
| `resources.limits.cpu`     | CPU resource limits                             | `100m`                                                     |



## Using the Chart to deploy your Application to Kubernetes

In order to use the Helm chart to deploy and verify your applicaton in Kubernetes, run the following commands:

1. From the directory containing `Chart.yaml`, run:  

  ```sh
  helm install --name silent-assassin
  ```  
  This deploys and runs your applicaton in Kubernetes, and prints the following text to the console:  


## Uninstalling your Application
If you installed your application with:  

```sh
helm install --name silent-assassin .
```
then you can:

* Find the deployment using `helm list --all` and searching for an entry with the chart name "nodeserver".
* Remove the application with `helm delete --purge nodeserver`.
