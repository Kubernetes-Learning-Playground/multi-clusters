#!/bin/bash

# 编译 Go 代码
go build -o kubectl-multicluster ./cmd/ctl_plugin/main.go

# 检查编译是否成功
if [ $? -eq 0 ]; then
    echo "Go build successful."
    #
    chmod +x kubectl-multicluster
    # 放入可执行文件到 /usr/local/bin
    sudo mv kubectl-multicluster /usr/local/bin/

    # 检查复制是否成功
    if [ $? -eq 0 ]; then
        echo "Copied kubectl-multicluster to /usr/local/bin."
    else
        echo "Failed to copy kubectl-multicluster to /usr/local/bin."
    fi
else
    echo "Go build failed."
fi