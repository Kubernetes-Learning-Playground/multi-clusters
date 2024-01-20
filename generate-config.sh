#!/bin/bash
# 用于生成 ctl 命令使用的配置文件


# 默认配置参数
# operator ip
serverIP="localhost"
# operator port
serverPort="31888"
# 主集群 kube config 路径。
# 注：如果 .multi-cluster-operator 是远程调用，
# 此变量不适用，亦即下列命令不生效：
# kubectl-multicluster apply -f yaml/multi_cluster_resource_configmap_example.yaml
# kubectl-multicluster delete -f yaml/multi_cluster_resource_configmap_example.yaml
masterClusterKubeConfigPath="/root/.kube/config"

# 解析选项标记
while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
    --port)
      serverPort="$2"
      shift
      shift
      ;;
    --ip)
      serverIP="$2"
      shift
      shift
      ;;
    --config)
      masterClusterKubeConfigPath="$2"
      shift
      shift
      ;;
    *)
      echo "错误：未知选项标记 $key"
      exit 1
      ;;
  esac
done

# 生成用户主目录下的配置文件目录
configDir="$HOME/.multi-cluster-operator"
if [ ! -d "$configDir" ]; then
  mkdir -p "$configDir"
fi

# 生成配置文件
configFile="$configDir/config"
cat <<EOF > "$configFile"
serverIP: $serverIP
serverPort: $serverPort
masterClusterKubeConfigPath: $masterClusterKubeConfigPath
EOF

echo "配置文件已生成：$configFile"