package eks

import (
	"testing"

	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/stretchr/testify/assert"
)

var (
	// CloudConfigData is retrieved from EC2
	CloudConfigData = `#cloud-config
packages: null
runcmd:
- - /var/lib/cloud/scripts/eksctl/bootstrap.al2.sh
write_files:
- content: |
    # eksctl-specific systemd drop-in unit for kubelet, for Amazon Linux 2 (AL2)

    [Service]
    # Local metadata parameters: REGION, AWS_DEFAULT_REGION
    EnvironmentFile=/etc/eksctl/metadata.env
    # Global and static parameters: CLUSTER_DNS, NODE_LABELS, NODE_TAINTS
    EnvironmentFile=/etc/eksctl/kubelet.env
    # Local non-static parameters: NODE_IP, INSTANCE_ID
    EnvironmentFile=/etc/eksctl/kubelet.local.env

    ExecStart=
    ExecStart=/usr/bin/kubelet \
    --node-ip=${NODE_IP} \
    --node-labels=${NODE_LABELS},alpha.eksctl.io/instance-id=${INSTANCE_ID} \
    --max-pods=${MAX_PODS} \
    --register-node=true --register-with-taints=${NODE_TAINTS} \
    --cloud-provider=aws \
    --container-runtime=docker \
    --network-plugin=cni \
    --cni-bin-dir=/opt/cni/bin \
    --cni-conf-dir=/etc/cni/net.d \
    --pod-infra-container-image=${AWS_EKS_ECR_ACCOUNT}.dkr.ecr.${AWS_DEFAULT_REGION}.${AWS_SERVICES_DOMAIN}/eks/pause:3.3-eksbuild.1 \
    --kubeconfig=/etc/eksctl/kubeconfig.yaml \
    --config=/etc/eksctl/kubelet.yaml
  owner: root:root
  path: /etc/systemd/system/kubelet.service.d/10-eksclt.al2.conf
  permissions: "0644"
- content: |-
    AWS_DEFAULT_REGION=eu-west-1
    AWS_EKS_CLUSTER_NAME=test3
    AWS_EKS_ENDPOINT=https://A7BADAD1856073EDEC64B69608E5940F.gr7.eu-west-1.eks.amazonaws.com
    AWS_EKS_ECR_ACCOUNT=602401143452
  owner: root:root
  path: /etc/eksctl/metadata.env
  permissions: "0644"
- content: |-
    NODE_LABELS=alpha.eksctl.io/cluster-name=test3,alpha.eksctl.io/nodegroup-name=ng-6909846a
    NODE_TAINTS=
  owner: root:root
  path: /etc/eksctl/kubelet.env
  permissions: "0644"
- content: |
    address: 0.0.0.0
    apiVersion: kubelet.config.k8s.io/v1beta1
    authentication:
    anonymous:
    enabled: false
    webhook:
    cacheTTL: 2m0s
    enabled: true
    x509:
    clientCAFile: /etc/eksctl/ca.crt
    authorization:
    mode: Webhook
    webhook:
    cacheAuthorizedTTL: 5m0s
    cacheUnauthorizedTTL: 30s
    cgroupDriver: systemd
    clusterDNS:
    - 10.100.0.10
    clusterDomain: cluster.local
    featureGates:
    RotateKubeletServerCertificate: true
    kind: KubeletConfiguration
    kubeReserved:
    cpu: 70m
    ephemeral-storage: 1Gi
    memory: 574Mi
    serverTLSBootstrap: true
  owner: root:root
  path: /etc/eksctl/kubelet.yaml
  permissions: "0644"
- content: |
    -----BEGIN CERTIFICATE-----
    MIICyDCCAbCgAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl
    cm5ldGVzMB4XDTIxMDMwMzA4MjU1NVoXDTMxMDMwMTA4MjU1NVowFTETMBEGA1UE
    AxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKve
    nTs6xoIsHNgupAt5N0OpH/xPcEM+JI6CLhSislQAWBj1JkhFC8ZUg3gQRXpE/Row
    sVQYLSaPtgoY8Ou3bG0dQvJsW1WASYwbC0AUMnTZ6Ykt6d79j0yHn6itnHIC15Ay
    /Q4RJu1wKxrUenhE8bp+H1FpInuhtRZbAYhAiUrnfG1DKHhuwrvFDCqRQQsMFDLV
    sblFhboIQXvvZgNrUPOEClqZqdwBES2fAosYkXcd7mvvbJID+5oE8lH5PTeL20al
    j1lJY9IMFaHr7AfyOUDkgOCsCHnXYdAtX/9aneQlno0nDK3izPiaVzHmUAWq/8k/
    OC1kiWU7wfLZKJvX2+MCAwEAAaMjMCEwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB
    /wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAGkbyV2nIlaiC0yfWhut9gcGuTbj
    YvH2uB06MdIPmwEgR8NpA89c1PEBijJVtLh4W4YVDxbqqwhVdZOxIzTfvX+WVjbC
    2GsgBd3gSGHNa1jAV6pCVUkfzUfindwYxBgN2igNJXpJOcqcHnHB7BM+r0euWxHL
    pyIChuZCzUPJWQl1C2SsSLqbVQ7OyDbxm5fCsL+X2v1QEypgGDQ+EBJoA8T0rKTl
    1MoeZpk72vMdz2DmpS7yWnWlyIENhJK3lANdnlsejX7RsSkxOWQ2KrVIc2M3arRi
    e6WG3azW6hfE4sR9d1bC2dV9yDM4Rx3XWPpVIxXXDaqFi5qbFcA3xP1UXHA=
    -----END CERTIFICATE-----
  owner: root:root
  path: /etc/eksctl/ca.crt
  permissions: "0644"
- content: |
    apiVersion: v1
    clusters:
    - cluster:
        certificate-authority: /etc/eksctl/ca.crt
        server: https://A7BADAD1856073EDEC64B69608E5940F.gr7.eu-west-1.eks.amazonaws.com
      name: test3.eu-west-1.eksctl.io
    contexts:
    - context:
        cluster: test3.eu-west-1.eksctl.io
        user: kubelet@test3.eu-west-1.eksctl.io
      name: kubelet@test3.eu-west-1.eksctl.io
    current-context: kubelet@test3.eu-west-1.eksctl.io
    kind: Config
    preferences: {}
    users:
    - name: kubelet@test3.eu-west-1.eksctl.io
      user:
        exec:
          apiVersion: client.authentication.k8s.io/v1alpha1
          args:
          - eks
          - get-token
          - --cluster-name
          - test3
          - --region
          - eu-west-1
          command: aws
          env:
          - name: AWS_STS_REGIONAL_ENDPOINTS
            value: regional
  owner: root:root
  path: /etc/eksctl/kubeconfig.yaml
  permissions: "0644"
- content: |
    c5d.18xlarge 737
    d3.8xlarge 59
    r4.16xlarge 737
    h1.4xlarge 234
  owner: root:root
  path: /etc/eksctl/max_pods.map
  permissions: "0644"
- content: |
    {
    "bridge": "none",
    "exec-opts": [
    "native.cgroupdriver=systemd"
    ],
    "log-driver": "json-file",
    "log-opts": {
    "max-size": "10m",
    "max-file": "10"
    },
    "live-restore": true,
    "max-concurrent-downloads": 10
    }
  owner: root:root
  path: /etc/docker/daemon.json
  permissions: "0644"
- content: |
    #!/bin/bash

    set -o errexit
    set -o pipefail
    set -o nounset

    function get_max_pods() {
    while read instance_type pods; do
    if  [[ "${instance_type}" == "${1}" ]] && [[ "${pods}" =~ ^[0-9]+$ ]] ; then
        echo ${pods}
        return
    fi
    done < /etc/eksctl/max_pods.map
    }

    # Use IMDSv2 to get metadata
    TOKEN="$(curl --silent -X PUT -H "X-aws-ec2-metadata-token-ttl-seconds: 600" http://169.254.169.254/latest/api/token)"
    function get_metadata() {
    curl --silent -H "X-aws-ec2-metadata-token: $TOKEN" "http://169.254.169.254/latest/meta-data/$1"
    }

    NODE_IP="$(get_metadata local-ipv4)"
    INSTANCE_ID="$(get_metadata instance-id)"
    INSTANCE_TYPE="$(get_metadata instance-type)"
    AWS_SERVICES_DOMAIN="$(get_metadata services/domain)"


    source /etc/eksctl/kubelet.env # this can override MAX_PODS

    INSTANCE_LIFECYCLE="$(get_metadata instance-life-cycle)"
    NODE_LABELS="${NODE_LABELS},node-lifecycle=${INSTANCE_LIFECYCLE}"


    cat > /etc/eksctl/kubelet.local.env <<EOF
    NODE_IP=${NODE_IP}
    INSTANCE_ID=${INSTANCE_ID}
    INSTANCE_TYPE=${INSTANCE_TYPE}
    AWS_SERVICES_DOMAIN=${AWS_SERVICES_DOMAIN}
    MAX_PODS=${MAX_PODS:-$(get_max_pods "${INSTANCE_TYPE}")}
    NODE_LABELS=${NODE_LABELS}
    EOF

    systemctl daemon-reload
    systemctl enable kubelet
    systemctl start kubelet
  owner: root:root
  path: /var/lib/cloud/scripts/eksctl/bootstrap.al2.sh
  permissions: "0755"
`
	shellScript = `MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="//"

--//
Content-Type: text/x-shellscript; charset="us-ascii"
#!/bin/bash
set -ex
B64_CLUSTER_CA=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeE1ETXdNekUzTWpZd05Gb1hEVE14TURNd01URTNNall3TkZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBSnAxCjVrN2FPRWpyWnBJTmN0cG5MZWl0aDBkM3UrR1E0SlpnU1VQU2hSL0lIdWZ2WTNMandTVUc5M05QNDgrV1pkUUQKM1Z4WnJWdXpsOTRMR3gwa29TUytSUmhmcWpJTHlYSm5mT2lJWjJkTWdjZURYdG1jWHFuSGpFQ1l1MVp5ejFQago2WHJ3dFIwYXJMaS9XSXpzMTlFa2M3V2RZMDJKV0VWZUV0ZWpnNzFzaUUvRk52VmtBV0p4MXpGOXRxWnBOY25ECmdpbDhtZ1VsWjdHVVZyTExnZEgwUjFxdCtDdzRvcGFTMWhCa0tCeVdQUFhJY25JSFZIWUxoZnRCL09JV2FOM0IKQytMQnhMaEFvQmQ4RUZoa2hxTjZwRjFiMTlaSi9XdFI3bWVVaWgrcUxmdy9HTWZpME1xUkY0MDFWdVNvbktrZAp3aGVGTm9CajU0d2xaLzVIVTdNQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFGYWhXZ2hManQ1bkUxSUxQNnozNTkxaE5HQTQKMHdncXRZMndPczZsejMvdkZ2M2s1R1hYNjVNUGk1UFEwYm1Dc29QeWdRMnFyck1oWkVUbk52d2d4RXRDeW84TwowTzV6a3BwYmMvM1ZDcHJpVXJXSHl4bnVIclVVa0xJK0J6aWNWVk42blJhYlJGLzhIYnBJTUMwMVEyMWJ5WVEyCnFjUFpjYVdyTzd2OURrMW5YVk9QUTVDRmZPMmFxb1dvSG1ER3ZOZlFMZzhiUDg3R2VCZStTTzJrZGxBVmJGZWMKS0xBdDlSblE3cjBTNk5Xa09xNS85d1hPYk42aVlZUVNiV1l5Z2tBZk56RWkvMUtKeU93aEdSVjd3VEErZ2FSTwpGUVFKT1lhcUlyQUtTdUw4elVjbHpBK3paYWtQbmV2aUhkR1hNMWY3enVZS1UyMklDQWVBdGkxZnVTaz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
API_SERVER_URL=https://36C6589489FA0EE0A77C38D3A0552682.gr7.eu-west-1.eks.amazonaws.com
K8S_CLUSTER_DNS_IP=10.100.0.10
/etc/eks/bootstrap.sh test4 --kubelet-extra-args '--node-labels=eks.amazonaws.com/sourceLaunchTemplateVersion=1,alpha.eksctl.io/cluster-name=test4,alpha.eksctl.io/nodegroup-name=ng-8be0f92d,eks.amazonaws.com/nodegroup-image=ami-08848b4d899d0b3de,eks.amazonaws.com/capacityType=ON_DEMAND,eks.amazonaws.com/nodegroup=ng-8be0f92d,eks.amazonaws.com/sourceLaunchTemplateId=lt-05c479b6852ad90d4' --b64-cluster-ca $B64_CLUSTER_CA --apiserver-endpoint $API_SERVER_URL --dns-cluster-ip $K8S_CLUSTER_DNS_IP`
)

func TestParseCloudConfigNoGzip(t *testing.T) {

	kubeConfig, err := ParseCloudConfig([]byte(CloudConfigData))
	if err != nil {
		t.Errorf("err: %v", err)
	}

	// if you change the data above, make sure the cluster name matches here
	assert.Contains(t, kubeConfig.Clusters, "test3.eu-west-1.eksctl.io")
}

func TestParseCloudConfigGzip(t *testing.T) {
	// gzip CloudConfigData
	gzipCloudCloudConfigData, err := common.GzipData([]byte(CloudConfigData))
	if err != nil {
		t.Errorf("err: %v", err)
	}

	kubeConfig, err := ParseCloudConfig(gzipCloudCloudConfigData)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	// if you change the data above, make sure the cluster name matches here
	assert.Contains(t, kubeConfig.Clusters, "test3.eu-west-1.eksctl.io")
}
