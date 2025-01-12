# Dummy Controller

This controller is a small and simple Kubernetes Custom Controller written using
[Operator SDK](https://github.com/operator-framework/operator-sdk) in Golang.

## Description

Comes with this controller is a new custom resource type called "dummies", of group "interview.com".
Whenever a new instance of this CRD is created, a new Nginx pod of name "<dummy-name>-nginx" ought to
be created in the same namespace, and it will be cleaned as the CR is cleaned.

## Getting Started

### Prerequisites
- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to [Docker Hub](https://hub.docker.com/). The image of our custom controller is publicly available at `lthh91/dummy-controller:latest`

### Create a cluster

- If you don't have cluster-admin rights in any kubernetes clusters, it's possible to create one with either [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
or [minikube](https://minikube.sigs.k8s.io/docs/start/?arch=%2Flinux%2Fx86-64%2Fstable%2Fbinary+download)

#### With Kind

```sh
kind create cluster
```

#### With minikube

```sh
minikube start
```

- Verify that your cluster is accessible:

```sh
kubectl get nodes
```

### To Deploy on the cluster

```sh
make deploy
```
You can apply the samples (examples) from the sample directory:

```sh
kubectl apply -f samples/v1alpha1_dummy.yaml
```

You should see that a new nginx pod being created in `default` namespace.

### To Uninstall

**Uninstall the sample dummy**:

```sh
kubectl delete -f samples/v1alpha1_dummy.yaml
```

The aforementioned nginx pod should be removed from `default` namespace.

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

### To cleanup the cluster

#### With Kind

```sh
kind delete cluster
```

#### With minikube

```sh
minikube stop
```

## NOTES


Our image was built on a Ubuntu 24.04 machine with AMD64 chip, so in case your cluster is not running Linux 
(or is created on top of a non-linux) machine, you may need to rebuild the image, push it to a repository
and deploy as followed:

```sh
IMG=<your-image-path> make docker-build
IMG=<your-image-path> make docker-push
IMG=<your-image-path> make deploy
```
