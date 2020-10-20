# Silent-assassin

**Silent-Assassin(SA)** is a project built to solve the problems of using Preemptible Virtual Machines(PVM) in Production environment.
PVMs are unused VMs in GCP, that come at 1/4th of the cost of regular on demand VMs. While the cost part is sweet, they have two limitations.

1) **They last a maximum of 24 hours**
2) **GCP provides no availability guarantees**

Because of these limitations, GCP recommends using PVMs for short-lived jobs and fault-tolerant workloads. So what are the problems we will face if we use PVMs to serve our stateless micro services ?

1) **Potential large scale disruption**
2) **Unanticipated preemption**

SA solves the problem of mass deletion (Problem 1) by deleting the VMs randomly after 12 hours and before 24 hours of its creation, during non-business hours. It solves the 2nd problem, which is the unpredicted loss of pods due to early preemption, by triggering a drain through kubernetes in the event of a preemption.

This tool is inspired by two projects aimed at solving the above problems [estafette-gke-preemption-killer](https://github.com/estafette/estafette-gke-preemptible-killer) and [k8s-node-termination-handler](https://github.com/GoogleCloudPlatform/k8s-node-termination-handler). We wanted to make alterations to the above projects ans combine the functionalities into one tool. So we built SA.


## Archicture
The architecture of SA is breifly explained [here](docs).

## Installation

Steps to enable authentication to GCP and install using helm-chart is explained [here](docs/install.md).
## Deployments.

How to plan for nodepools and pod affinity is explained [here](docs/nodepools-and-deployment.md).

## Disaster Recovery

What are the steps to perform when PVMs misbehave? It is explained [here](docs/disaster-recovery.md).


## Development

Building from source.

```
make all
```

## Contribution
If you find any issues in using this project, you can raise issues.This project is [Apache 2.0 licensed](LICENSE) and we accept contributions via GitHub pull requests.