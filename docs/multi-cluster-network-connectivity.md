### 跨集群Pod网络连通方案
- 使用开源方案 [Submariner](https://github.com/submariner-io) 实现跨集群网络流量
- 测试: 以集群 kind 为例

1. cluster1
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: cluster1
nodes:
  - role: control-plane
    image: kindest/node:v1.24.15
    extraPortMappings:
      - containerPort: 6443
        hostPort: 36443  #安全组开放
        protocol: tcp
    extraMounts:
      - hostPath: /root/kind/node1-1
        containerPath: /files

  - role: worker
    image: kindest/node:v1.24.15
    extraPortMappings:   # 将 node 的端口映射到主机   我们会想办法把 ingress gateway部署在这个节点上
      - containerPort: 80
        hostPort: 30080
        protocol: tcp
      - containerPort: 443
        hostPort: 30443
        protocol: tcp
    labels:
      gateway: true
    extraMounts:
      - hostPath: /root/kind/node1-2
        containerPath: /files
  - role: worker
    image: kindest/node:v1.24.15
    extraMounts:
      - hostPath: /root/kind/node1-3
        containerPath: /files
networking:
  apiServerAddress: "172.19.0.12"
  apiServerPort: 6443
  podSubnet: "10.6.0.0/16" #自定义 pod IP 地址范围
  serviceSubnet: "10.96.0.0/16"
  kubeProxyMode: "ipvs"
```
2. cluster2
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: cluster2  #第二个集群
nodes:
  - role: control-plane
    image: kindest/node:v1.24.15
    extraPortMappings:
      - containerPort: 6445   # 注意这里的端口，不是 6443 否则会和jt1 冲突
        hostPort: 46443  #安全组开放
        protocol: tcp
    extraMounts:
      - hostPath: /root/kind/node2-1
        containerPath: /files

  - role: worker
    image: kindest/node:v1.24.15
    extraMounts:
      - hostPath: /root/kind/node2-2
        containerPath: /files

networking:
  apiServerAddress: "172.19.0.12"
  apiServerPort: 6445
  podSubnet: "10.7.0.0/16" #自定义 pod IP 地址范围
  serviceSubnet: "10.97.0.0/16"
  kubeProxyMode: "ipvs"
```
- 安装 subctl 工具
```bash
# 安装 subctl 工具
wget https://github.com/submariner-io/releases/releases/download/v0.16.2/subctl-v0.16.2-linux-amd64.tar.gz
tar -zxvf subctl-v0.16.2-linux-amd64.tar.gz
cd subctl-v0.16.2/
cd subctl-v0.16.2/
chmod +x subctl
mv subctl /usr/local/bin
```
- 安装 kind 测试集群
```bash
root@VM-0-12-ubuntu:~# kind create cluster --config=cluster1.yaml
Creating cluster "cluster1" ...
 ✓ Ensuring node image (kindest/node:v1.24.15) 🖼
 ✓ Preparing nodes 📦 📦 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
 ✓ Joining worker nodes 🚜
Set kubectl context to "kind-cluster1"
You can now use your cluster with:

kubectl cluster-info --context kind-cluster1

Thanks for using kind! 😊
root@VM-0-12-ubuntu:~# kind create cluster --config=cluster2.yaml
Creating cluster "cluster2" ...
 ✓ Ensuring node image (kindest/node:v1.24.15) 🖼
 ✓ Preparing nodes 📦 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
 ✓ Joining worker nodes 🚜
Set kubectl context to "kind-cluster2"
You can now use your cluster with:

kubectl cluster-info --context kind-cluster2

Have a question, bug, or feature request? Let us know! https://kind.sigs.k8s.io/#community 🙂
```

- 安装 submariner operator 
```yaml
root@VM-0-12-ubuntu:~# subctl --context kind-cluster1 deploy-broker
  ✓ Setting up broker RBAC
  ✓ Deploying the Submariner operator
  ✓ Created operator CRDs
✓ Created operator namespace: submariner-operator
  ✓ Created operator service account and role
  ✓ Created submariner service account and role
  ✓ Created lighthouse service account and role
  ✓ Deployed the operator successfully
  ✓ Deploying the broker
  ✓ Saving broker info to file "broker-info.subm"
  ✓ Backed up previous file "broker-info.subm" to "broker-info.subm.2024-03-19T22_34_14+08_00"
root@VM-0-12-ubuntu:~# kubectl get pods --context kind-cluster1 -nsubmariner-operator
NAME                                  READY   STATUS    RESTARTS   AGE
submariner-operator-f8b9cdbbf-dlmtl   1/1     Running   0          60s
```
- 加入集群到网络平面中
```bqsh
root@VM-0-12-ubuntu:~# subctl --context kind-cluster1 join broker-info.subm --clusterid cluster1
 ✓ broker-info.subm indicates broker is at https://172.19.0.12:6443
 ✓ Discovering network details
        Network plugin:  kindnet
        Service CIDRs:   [10.96.0.0/16]
        Cluster CIDRs:   [10.6.0.0/16]
 ✓ Retrieving the gateway nodes
 ✓ Retrieving all worker nodes
? Which node should be used as the gateway? cluster1-worker
 ✓ Labeling node "cluster1-worker" as a gateway
 ✓ Gathering relevant information from Broker
 ✓ Retrieving Globalnet information from the Broker
 ✓ Validating Globalnet configuration
 ✓ Deploying the Submariner operator
 ✓ Created operator namespace: submariner-operator
 ✓ Creating SA for cluster
 ✓ Connecting to Broker
 ✓ Deploying submariner
 ✓ Submariner is up and running
root@VM-0-12-ubuntu:~# kubectl get pods --context kind-cluster1 -nsubmariner-operator
NAME                                             READY   STATUS    RESTARTS   AGE
submariner-gateway-xnmnw                         1/1     Running   0          48s
submariner-lighthouse-agent-84dd959f45-kgx6p     1/1     Running   0          47s
submariner-lighthouse-coredns-77d855c7c5-kl6vq   1/1     Running   0          46s
submariner-lighthouse-coredns-77d855c7c5-z8hhj   1/1     Running   0          46s
submariner-metrics-proxy-n6mdn                   1/1     Running   0          47s
submariner-operator-f8b9cdbbf-dlmtl              1/1     Running   0          3m31s
submariner-routeagent-cw8g5                      1/1     Running   0          47s
submariner-routeagent-krcgb                      1/1     Running   0          47s
submariner-routeagent-xxj8l                      1/1     Running   0          47s
root@VM-0-12-ubuntu:~# subctl --context kind-cluster2 join broker-info.subm --clusterid cluster2
 ✓ broker-info.subm indicates broker is at https://172.19.0.12:6443
 ✓ Discovering network details
        Network plugin:  kindnet
        Service CIDRs:   [10.97.0.0/16]
        Cluster CIDRs:   [10.7.0.0/16]
 ✓ Retrieving the gateway nodes
 ✓ Retrieving all worker nodes
? Which node should be used as the gateway? cluster2-worker
 ✓ Labeling node "cluster2-worker" as a gateway
 ✓ Gathering relevant information from Broker
 ✓ Retrieving Globalnet information from the Broker
 ✓ Validating Globalnet configuration
 ✓ Deploying the Submariner operator
 ✓ Created operator CRDs
 ✓ Created operator namespace: submariner-operator
 ✓ Created operator service account and role
 ✓ Created submariner service account and role
 ✓ Created lighthouse service account and role
 ✓ Deployed the operator successfully
 ✓ Creating SA for cluster
 ✓ Connecting to Broker
 ✓ Deploying submariner
 ✓ Submariner is up and running
root@VM-0-12-ubuntu:~# kubectl get pods --context kind-cluster2 -nsubmariner-operator
NAME                                             READY   STATUS              RESTARTS   AGE
submariner-gateway-htgwv                         0/1     ContainerCreating   0          9s
submariner-lighthouse-agent-7f6667685f-ks2bm     0/1     ContainerCreating   0          8s
submariner-lighthouse-coredns-7bc5f84dbb-58xd2   0/1     ContainerCreating   0          8s
submariner-lighthouse-coredns-7bc5f84dbb-qb28c   0/1     ContainerCreating   0          8s
submariner-metrics-proxy-vw7mf                   0/1     ContainerCreating   0          8s
submariner-operator-f8b9cdbbf-h6zg6              1/1     Running             0          14s
submariner-routeagent-g86fh                      0/1     Init:0/1            0          8s
submariner-routeagent-m4n6s                      0/1     Init:0/1            0          8s
```