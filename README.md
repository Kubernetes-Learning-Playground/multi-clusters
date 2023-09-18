## kubernetes多集群多资源存储方案

### 项目思路与功能
项目背景：在多集群应用场景中，会时常有根据不同集群查询资源的需求，基于此需求，本项目采用informer进行扩展封装，实现"**多集群**"且"**多资源**"的
存储方案。

支持功能：
1. 支持"多集群"配置
2. 支持多资源配置
3. 支持跳过tls认证
4. 实现 http server 支持查询接口

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
        - rType: core/v1/pods
        - rType: core/v1/configmaps
  - metadata:
      clusterName: cluster2   # 自定义集群名
      insecure: true          # 是否开启跳过tls证书认证
      configPath: /Users/zhenyu.jiang/go/src/golanglearning/new_project/multi_resource/resources/config1 # kube config配置文件地址
      resources:
        - rType: apps/v1/deployments
        - rType: core/v1/pods
        - rType: core/v1/configmaps
```

