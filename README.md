## kubernetes多集群多资源存储方案

### 项目思路与功能
项目背景：在多集群应用场景中，会时常有根据不同集群查询资源的需求，基于此需求，本项目采用informer进行扩展封装，实现"**多集群**"且"**多资源**"的
存储方案。

支持功能：
1. 支持"多集群"配置
2. 支持多资源配置
3. 支持跳过tls认证
4. 实现 http server 支持查询接口
5. 支持查询多集群命令行插件

![](https://github.com/Kubernetes-Learning-Playground/multi-cluster-resource-storage/blob/main/image/%E6%97%A0%E6%A0%87%E9%A2%98-2023-08-10-2343.png?raw=true)

### 配置文件
- **重要** 配置文件可参考config.yaml中配置，调用方只需要关注配置文件中的内容即可。
```yaml
clusters:                     # 集群列表
  - metadata:
      clusterName: cluster1   # 自定义集群名
      insecure: false          # 是否开启跳过tls证书认证
      configPath: /Users/zhenyu.jiang/.kube/config # kube config配置文件地址
      # 资源类型
      resources:
        - rType: apps/v1/deployments
        - rType: core/v1/resource
        - rType: core/v1/configmaps
  - metadata:
      clusterName: cluster2   # 自定义集群名
      insecure: true          # 是否开启跳过tls证书认证
      configPath: /Users/zhenyu.jiang/go/src/golanglearning/new_project/multi_resource/resource/config1 # kube config配置文件地址
      resources:
        - rType: apps/v1/deployments
        - rType: core/v1/resource
        - rType: core/v1/configmaps
```

### 多集群命令行查询
目前支持查询资源
- pods
- configmaps
- deployments
注：后缀参数：
- --namesapce：按命名空间查询，不填默认所有命名空间
- --clusterName：按集群名查询，不填默认所有集群
- --name: 按名称查询，不填默认所有名称
```bash
➜  cmd git:(main) ✗ go run ctl_plugin/main.go configmaps --clusterName=cluster2      
集群名称        CONFIGMAP                               NAMESPACE               DATA 
cluster2        test-scheduling-config                  kube-system             1       
cluster2        loki-loki-stack-test                    loki-stack              1       
cluster2        kube-root-ca.crt                        loki-stack              1       
cluster2        loki-loki-stack                         loki-stack              1       
cluster2        kube-root-ca.crt                        etcd01                  1       
cluster2        kube-root-ca.crt                        mycsi                   1  

➜  cmd git:(main) ✗ go run ctl_plugin/main.go configmaps --clusterName=cluster2 --name=coredns --namespace=kube-system       
集群名称        CONFIGMAP       NAMESPACE       DATA 
cluster2        coredns         kube-system     1       

```

```bash
➜  cmd git:(main) ✗ go run ctl_plugin/main.go pods --clusterName=cluster2                                   
集群名称         POD名称                                                  NAMESPACE               POD IP          状态             容器名                           容器静像                                                                        
cluster2        virtual-kubelet-pod-test-bash                           default                                 Running         ngx1                            nginx:1.18-alpine                                                                    
cluster2        testpod1                                                default                                 Running         mytest                          nginx:1.18-alpine                                                                    
cluster2        loki-promtail-zxpvg                                     loki-stack                              Running         promtail                        docker.io/grafana/promtail:2.7.4                                                     
cluster2        node-exporter-srqk4                                     prometheus                              Running         node-exporter                   bitnami/node-exporter:1.4.0                                                          
cluster2        node-exporter-m5whb                                     prometheus                              Running         node-exporter                   bitnami/node-exporter:1.4.0                                                          
cluster2        loki-promtail-fcpsb                                     loki-stack                              Running         promtail                        docker.io/grafana/promtail:2.7.4                                                     
cluster2        testpod                                                 default                                 Pending         mytest                          nginx:1.18-alpine                                                                    
cluster2        nginx-kubelet                                           default                                 Running         nginx                           nginx:1.18-alpine                                                                    
cluster2        dep-test-8b4fcc97-pzbqd                                 default                 10.244.0.124    Running         dep-test-container              nginx:1.18-alpine                                                                    
cluster2        dep-test-8b4fcc97-jkkx7                                 default                 10.244.0.127    Running         dep-test-container              nginx:1.18-alpine                                                                    
cluster2        dep-test-8b4fcc97-wl6td                                 default                 10.244.0.128    Running         dep-test-container              nginx:1.18-alpine                                               
```

```bash
➜  cmd git:(main) ✗ go run ctl_plugin/main.go deployments --clusterName=cluster2
集群名称        DEPLOYMENT                              NAMESPACE               TOTAL   AVAILABLE       READY 
cluster2        dep-test                                default                 5       5               5       
cluster2        testngx                                 default                 10      10              10      
cluster2        test-pod-maxnum-scheduler               kube-system             1       1               1       
cluster2        myingress-controller                    default                 1       1               1       
cluster2        myapi                                   default                 1       1               1       
```