## kubernetes 简易多集群方案
<a href="./README.md">English</a> | <a href="./README-zh.md">简体中文</a>
### 项目思路与功能
项目背景：在目前云原生中，会常有需要同时操作"多集群"的场景，不论是多集群"查询"或是"分发资源"等操作，本项目采用 **informer** + **operator** 进行扩展封装，
实现**多集群**且**多资源**方案。

支持功能：
1. 支持"多集群"配置
2. 支持"多资源"配置
3. 支持跳过 restconfig tls 认证
4. 实现 http server 支持查询接口
5. 支持查询多集群命令行插件(list,describe,apply,delete,join,unjoin)
6. 支持多集群下发资源
7. 支持多集群**差异化配置**
8. 支持集群**动态加入与删除**

### 配置文件
- **重要** 配置文件可参考 config.yaml 中配置[这里](./config.yaml)，调用方只需要关注配置文件中的内容即可。
- 此配置文件在 deployment 部署时需要挂载 [参考](./deploy/deploy.yaml) 中 volumes 位置(name: config 的挂载)
```yaml
clusters:                     # 集群列表
  - metadata:
      clusterName: tencent1   # 自定义集群名
      insecure: false         # 是否开启跳过 tls 证书认证
      configPath: /Users/zhenyu.jiang/.kube/config # kube config 配置文件地址
  - metadata:
      clusterName: tencent2   # 自定义集群名
      insecure: true          # 是否开启跳过 tls 证书认证
      isMaster: true          # 标示主集群
      configPath: /Users/zhenyu.jiang/go/src/golanglearning/new_project/multi_resource/multiclusterresource/config1 # kube config配置文件地址
```
![](https://github.com/Kubernetes-Learning-Playground/multi-cluster-resource-storage/blob/main/image/%E6%97%A0%E6%A0%87%E9%A2%98-2023-08-10-2343.png?raw=true)

### 多集群命令行查询(也支持 http server 接口查询)
目前已支持**大部分 k8s 资源**的查询，需要输入资源对象的 **GVR** 如：v1/pods or batch/v1/jobs or v1/apps/deployments

- (除特殊资源外，如 metrics.k8s.io authentication.k8s.io authorization.k8s.io 这几种 group 目前不支持)
- core 组的资源对象支持输入是 core/v1/pods or v1/pods 两种形式，core 可写可不写

后缀参数：
- --namespace：按命名空间查询，不填默认所有命名空间
- --clusterName：按集群名查询，不填默认所有集群
- --name: 按名称查询，不填默认所有名称
```bash
➜  cmd git:(main) ✗ go run ctl_plugin/main.go list v1/configmaps --clusterName=tencent2
集群名称        NAME                                    NAMESPACE       DATA 
tencent2        multiclusterresource-configmap          default         3       
tencent2        kube-root-ca.crt                        kube-flannel    1       
tencent2        kube-flannel-cfg                        kube-flannel    2       
tencent2        kube-root-ca.crt                        kube-public     1       
tencent2        kube-root-ca.crt                        default         1       
tencent2        kube-root-ca.crt                        kube-node-lease 1       
tencent2        kube-root-ca.crt                        kube-system     1       
tencent2        cluster-info                            kube-public     1  

➜  cmd git:(main) ✗ go run ctl_plugin/main.go v1/configmaps --clusterName=tencent2 --name=coredns --namespace=kube-system       
集群名称        CONFIGMAP       NAMESPACE       DATA 
tencent2        coredns         kube-system     1       
```
查询多集群 pods 资源
```bash
➜  cmd git:(main) ✗ go run ctl_plugin/main.go list core/v1/pods --clusterName=tencent2                                   
集群名称         NAME                                                    NAMESPACE               POD IP          状态             容器名                           容器静像                                                                        
tencent2        virtual-kubelet-pod-test-bash                           default                                 Running         ngx1                            nginx:1.18-alpine                                                                    
tencent2        testpod1                                                default                                 Running         mytest                          nginx:1.18-alpine                                                                    
tencent2        loki-promtail-zxpvg                                     loki-stack                              Running         promtail                        docker.io/grafana/promtail:2.7.4                                                     
tencent2        node-exporter-srqk4                                     prometheus                              Running         node-exporter                   bitnami/node-exporter:1.4.0                                                          
tencent2        node-exporter-m5whb                                     prometheus                              Running         node-exporter                   bitnami/node-exporter:1.4.0                                                          
tencent2        loki-promtail-fcpsb                                     loki-stack                              Running         promtail                        docker.io/grafana/promtail:2.7.4                                                     
tencent2        testpod                                                 default                                 Pending         mytest                          nginx:1.18-alpine                                                                    
tencent2        nginx-kubelet                                           default                                 Running         nginx                           nginx:1.18-alpine                                                                    
tencent2        dep-test-8b4fcc97-pzbqd                                 default                 10.244.0.124    Running         dep-test-container              nginx:1.18-alpine                                                                    
tencent2        dep-test-8b4fcc97-jkkx7                                 default                 10.244.0.127    Running         dep-test-container              nginx:1.18-alpine                                                                    
tencent2        dep-test-8b4fcc97-wl6td                                 default                 10.244.0.128    Running         dep-test-container              nginx:1.18-alpine                                               

# 不指定 clusterName，默认查询所有集群
➜  multi_resource git:(main) ✗ go run cmd/ctl_plugin/main.go list core/v1/pods                           
集群名称         NAME                                                            NAMESPACE                               NODE                    POD IP          状态             容器名                        容器静像                                                                            
tencent1        patch-deployment-7877dfff-975bn                                 default                                 minikube                10.244.1.40     Running         nginx                        nginx:1.15.2                                                                            
tencent1        patch-deployment-7877dfff-dwpxj                                 default                                 minikube                10.244.1.39     Running         nginx                        nginx:1.15.2                                                                            
tencent2        virtual-kubelet-pod-test-bash                                   default                                 mynode                                  Running         ngx1                         nginx:1.18-alpine                                                                       
tencent1        kueue-controller-manager-56987d8f8c-69gr7                       kueue-system                            minikube                10.244.1.16     Running         manager                      registry.k8s.io/kueue/kueue:v0.4.1                                                      
tencent2        testpod1                                                        default                                 my-sample-kubelet                       Running         mytest                       nginx:1.18-alpine                                                                       
tencent2        loki-promtail-zxpvg                                             loki-stack                              my-sample-kubelet                       Running         promtail                     docker.io/grafana/promtail:2.7.4                                                        
tencent2        node-exporter-srqk4                                             prometheus                              my-sample-kubelet                       Running         node-exporter                bitnami/node-exporter:1.4.0                                                             
tencent2        node-exporter-m5whb                                             prometheus                              myk8s                                   Running         node-exporter                bitnami/node-exporter:1.4.0                                                             
tencent2        loki-promtail-fcpsb                                             loki-stack                              myk8s                                   Running         promtail                     docker.io/grafana/promtail:2.7.4                                                        
tencent2        testpod                                                         default                                 myjtthink                               Pending         mytest                       nginx:1.18-alpine                                                                       
tencent2        nginx-kubelet                                                   default                                 myjtthink                               Running         nginx                        nginx:1.18-alpine                                                                       
tencent2        dep-test-8b4fcc97-pzbqd                                         default                                 vm-0-16-centos          10.244.0.124    Running         dep-test-container           nginx:1.18-alpine                                                                       
tencent2        dep-test-8b4fcc97-jkkx7                                         default                                 vm-0-16-centos          10.244.0.127    Running         dep-test-container           nginx:1.18-alpine                                                                       
tencent2        dep-test-8b4fcc97-wl6td                                         default                                 vm-0-16-centos          10.244.0.128    Running         dep-test-container           nginx:1.18-alpine                                                                       
tencent2        dep-test-8b4fcc97-znlp5                                         default                                 vm-0-16-centos          10.244.0.125    Running         dep-test-container           nginx:1.18-alpine                                                                       
tencent2        dep-test-8b4fcc97-vxf55                                         default                                 vm-0-16-centos          10.244.0.126    Running         dep-test-container           nginx:1.18-alpine                                                                       
tencent2        inspect-script-task-task3--1-fjxm9                              default                                 vm-0-16-centos          10.244.0.94     Pending         default                      inspect-operator/script-engine:v1
```
查询多集群 deployments 资源
```bash
➜  cmd git:(main) ✗ go run ctl_plugin/main.go list apps/v1/deployments --clusterName=cluster2
集群名称         NAME                                    NAMESPACE               TOTAL   AVAILABLE       READY 
tencent1        dep-test                                default                 5       5               5       
tencent1        testngx                                 default                 10      10              10      
tencent1        test-pod-maxnum-scheduler               kube-system             1       1               1       
tencent1        myingress-controller                    default                 1       1               1       
tencent1        myapi                                   default                 1       1               1       
```
查询多集群 pods 资源详细
```bash
➜  cmd git:(main) ✗ go run ctl_plugin/main.go describe core/v1/pods --clusterName=tencent2 --namespace=default --name=myredis-0
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2023-01-18T15:14:48Z"
  managedFields:
  - apiVersion: v1
...  
```

#### 部署命令行工具
- 编译 ctl 文件，并放入 /bin
```bash
# 注意，需要 go version 1.20 版本以上
[root@VM-0-16-centos multiclusteroperator]# chmod +x build-ctl-command.sh
[root@VM-0-16-centos multiclusteroperator]# ./build-ctl-command.sh
Go build successful.
Copied kubectl-multicluster to /usr/local/bin. 
```
- 生成配置文件 (.multi-cluster-operator/config)
```bash
# 进入目录并进行编译
➜  ctl_plugin git:(main) ✗ pwd
/xxxxx/multi_resource/cmd/ctl_plugin
➜  ctl_plugin git:(main) ✗ go build -o kubectl-multicluster .
➜  ctl_plugin git:(main) ✗ chmod 777 kubectl-multicluster                     
➜  ctl_plugin git:(main) ✗ mv kubectl-multicluster ~/go/bin/ 
➜  ~ kubectl-multicluster list v1.pods --clusterName=tencent1 --name=multiclusterresource-deployment-75d98bb7bd-xj5z5
集群名称        NAME                                                    NAMESPACE       NODE            POD IP          状态    容器名                  容器镜像            
tencent1        multiclusterresource-deployment-75d98bb7bd-xj5z5        default         vm-0-12-centos  10.244.29.32    Running example-container       nginx:1.19.0-alpine   
```

注意：本项目默认命令行会读取 ~/.multi-cluster-operator/config 配置。
如果使用命令行时有相关报错，可以自行处理相关部分。
```bash
[root@VM-0-16-centos ~]# cat .multi-cluster-operator/config
serverIP: localhost
serverPort: 31888
masterClusterKubeConfigPath: /root/.kube/config
[root@VM-0-16-centos ~]#  
```

### 多集群下发资源
思路：上述的配置文件中会指定主集群(isMaster)字段，主集群用于部署 operator 控制器，即：多集群中只有主集群能下发资源与删除资源，其他集群都只能被动接收。

![](./image/%E6%97%A0%E6%A0%87%E9%A2%98-2023-08-11-2343.png?raw=true)

- crd 资源对象如下，更多信息可以参考 [参考](./yaml)
  - template 字段：资源模版，内部填写需要的 k8s 原始资源
  - placement 字段：选择下发集群
  - customize 字段：集群间差异化配置
```yaml
apiVersion: mulitcluster.practice.com/v1alpha1
kind: MultiClusterResource
metadata:
  name: mypod.pod
  namespace: default
spec:
   # 资源模版，内部填写需要的 k8s 原始资源
   template:
     apiVersion: v1
     kind: Pod
     metadata:
       name: multicluster-pod
       namespace: default
     spec:
       containers:
         - image: busybox
           command:
             - sleep
             - "3600"
           imagePullPolicy: IfNotPresent
           name: busybox
       restartPolicy: Always
   # 可以自行选择不同集群下发，如果修改后，
   # 也会相应的新增或删除特定集群的资源 
   # 注：如果 placement 删除，customize 相应的集群也需要删除！！
   placement:
     clusters:
       - name: tencent1
       - name: tencent2
       - name: tencent4
   # 可以不填写
   # 多集群间差异化配置
   customize:
     clusters:
       - name: tencent1
         action:
           # 替换镜像
           - path: "/spec/containers/0/image"
             op: "replace"
             value:
               - "nginx:1.19.0-alpine"
       - name: tencent2
         action:
           # 新增 annotations
           - path: "/metadata/annotations/example"
             value:
               - "example"
             op: "add"
```
- 使用

可以看出，当在主集群创建 CRD 后，会自动下发到其他集群。
```bash
# apply
➜  multi_resource git:(main) ✗ kubectl apply -f yaml/test.yaml    
multiclusterresource.mulitcluster.practice.com/mypod.pod created
# 查询
➜  multi_resource git:(main) ✗ kubectl get multiclusterresources.mulitcluster.practice.com    
NAME        AGE
mypod.pod   45m
➜  multi_resource git:(main) ✗ go run cmd/ctl_plugin/main.go list pods  --namespace=default --name=multicluster-pod
集群名称        NAME                    NAMESPACE       NODE            POD IP          状态    容器名  容器静像 
tencent4        multicluster-pod        default         vm-0-17-centos  10.244.167.193  Running busybox busybox         
tencent1        multicluster-pod        default         minikube        10.244.1.48     Running busybox busybox         
tencent2        multicluster-pod        default         vm-0-16-centos  10.244.0.142    Running busybox busybox   
# 使用 customize 实现差异化部署
➜  cmd git:(main) ✗ go run ctl_plugin/main.go list deployments --name=multiclusterresource-deployment
集群名称        NAME                            NAMESPACE       TOTAL   AVAILABLE       READY 
tencent4        multiclusterresource-deployment default         3       3               3       
tencent1        multiclusterresource-deployment default         2       2               2       
tencent2        multiclusterresource-deployment default         1       1               1       
```

### 部署应用
- 注：项目依赖 mysql，推荐使用  mariadb:10.5 镜像，依赖配置在 [deploy.yaml](./deploy/deploy.yaml) args 字段中设置。
- 注：部署前需要先创建 mysql 对应的库与表，表结构可参考 [mysql 表结构](./mysql/resources.sql)。
- 创建 mysql 库与表参考步骤 [参考](./mysql)
1. docker 镜像
```bash
[root@VM-0-16-centos multi_resource_operator]# docker build -t multi-cluster-operator:v1 .
Sending build context to Docker daemon  1.262MB
Step 1/17 : FROM golang:1.20.7-alpine3.17 as builder
 ---> 864c54ad9c0d
Step 2/17 : WORKDIR /app
 ---> Using cache
 ---> b1cd56e3903a
Step 3/17 : COPY go.mod go.mod
```

- 目前支持 helm 部署，可参考 [helm 部署](./helm)

2. 部署应用 deployment
```bash
[root@VM-0-16-centos multi_resource_operator]# kubectl apply -f deploy/rbac.yaml
serviceaccount/multi-cluster-operator-sa unchanged
clusterrole.rbac.authorization.k8s.io/multi-cluster-operator-clusterrole unchanged
clusterrolebinding.rbac.authorization.k8s.io/multi-cluster-operator-ClusterRoleBinding unchanged
[root@VM-0-16-centos multi_resource_operator]# kubectl apply -f deploy/deploy.yaml
deployment.apps/multi-cluster-operator unchanged
service/multi-cluster-operator-svc unchanged
```
3. 使用 kubectl log or exec 查看项目应用
```bash
[root@VM-0-16-centos ~]# kubectl get pods | grep multi-cluster-operator
multi-cluster-operator-7c477d7b58-z4bbd            1/1     Running            0          10d
[root@VM-0-16-centos ~]# kubectl logs multi-cluster-operator-7c477d7b58-z4bbd
I0928 07:29:45.956381       1 init_ctl_config.go:37] multi-cluster-ctl config file created successfully.
I0928 07:29:46.045113       1 multi_cluster.go:117] [tencent1] informer watcher start..
I0928 07:29:46.045147       1 handler.go:27] worker queue start...
I0928 07:29:46.146385       1 multi_cluster.go:117] [tencent2] informer watcher start..
I0928 07:29:46.146407       1 handler.go:27] worker queue start...
I0928 07:29:46.252256       1 multi_cluster.go:117] [tencent4] informer watcher start..
I0928 07:29:46.252275       1 handler.go:27] worker queue start...
I0928 07:29:46.353115       1 main.go:85] operator manager start...
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
```
4. 生成命令行配置文件
```bash
[root@VM-0-16-centos multiclusteroperator]# chmod +x generate-config.sh
[root@VM-0-16-centos multiclusteroperator]# ./generate-config.sh --ip localhost --port 31888 --config /root/.kube/config
配置文件已生成：/root/.multi-cluster-operator/config
[root@VM-0-16-centos ~]# cat .multi-cluster-operator/config
serverIP: localhost
serverPort: 31888
masterClusterKubeConfigPath: /root/.kube/config
[root@VM-0-16-centos ~]#
```