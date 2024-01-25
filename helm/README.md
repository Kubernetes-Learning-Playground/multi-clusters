## helm 部署

### 修改配置文件
用户需要**自行修改** [配置文件](./values.yaml)。
- base: 基础配置，镜像需要自行构建 (docker build...)
- rbac: 用于创建 rbac 使用
- db: 数据库配置
- service: service 配置
- MultiClusterConfiguration: 多集群中 每个集群的 kubeconfig 文件需要挂载到 pod 中

注：如果部署有任何问题，欢迎提 issue 或 直接联系
```yaml
# 基础配置
base:
  # 副本数
  replicaCount: 1
  # 应用名或 namespace
  name: multi-cluster-operator
  namespace: default
  serviceName: multi-cluster-operator-svc
  # 镜像名
  image: multi-cluster-operator:v1
  # 多集群配置文件目录 (注意：在容器中的位置)
  config: /app/file/config.yaml
  # 是否指定调度到某个节点，如果不需要则不填
  nodeName: vm-0-16-centos


# 用于创建 rbac 使用
rbac:
  serviceaccountname: multi-cluster-operator-sa
  namespace: default
  clusterrole: multi-cluster-operator-clusterrole
  clusterrolebinding: multi-cluster-operator-ClusterRoleBinding

# db 配置
db:
  dbuser: root
  dbpassword: 123456
  # 注意：必须容器网络可达
  dbendpoint: 10.0.0.16:30110
  dbdatabase: resources
  debugmode: false

# service 配置
service:
  type: NodePort
  ports:
    - port: 8888      # 容器端口
      nodePort: 31888 # 对外暴露的端口
      name: server
    - port: 29999     # 健康检查端口
      nodePort: 31889 # 对外暴露的端口
      name: health

# 多集群中 每个集群的 kubeconfig 文件需要挂载到 pod 中
MultiClusterConfiguration:
  volumeMounts:
    # 挂载不同集群的 kubeconfig，请自行修改
    - name: tencent1
      mountPath: /app/file/config-tencent1
    - name: tencent2
      mountPath: /app/file/config-tencent2
    - name: tencent4
      mountPath: /app/file/config-tencent4
    # 配置文件 config.yaml
    - name: config
      mountPath: /app/file/config.yaml
    - name: migrate
      mountPath: /app/migrations
  # 节点上的位置
  volumes:
    - name: tencent1
      hostPath:
        path: /root/multi_resource_operator/resources/config-tencent1
    - name: tencent2
      hostPath:
        path: /root/multi_resource_operator/resources/config-tencent2
    - name: tencent4
      hostPath:
        path: /root/multi_resource_operator/resources/config-tencent4
    - name: config
      hostPath:
        path: /root/multi_resource_operator/config.yaml
    - name: migrate
      hostPath:
        path: /root/multi_resource_operator/migrations
```
