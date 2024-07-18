#!/bin/bash

# 定义包含多个.proto文件名的数组
proto_files=("./proto/uc.proto" "./proto/public.proto")

# 遍历数组中的每个文件名
for proto_file in "${proto_files[@]}"; do
    # 检查文件是否存在（这一步是可选的，但建议加上以增加脚本的健壮性）
    if [ ! -f "$proto_file" ]; then
        echo "Error: File '$proto_file' does not exist."
        continue
    fi

    # 运行protoc命令
    protoc -I ./proto \
        --go_out=paths=source_relative:./internal/protoc \
        --go_opt=paths=source_relative \
        --go-grpc_out=paths=source_relative:./internal/protoc \
        --go-grpc_opt=paths=source_relative \
        "$proto_file"

    # 检查命令执行是否成功
    if [ $? -ne 0 ]; then
        echo "Failed to generate code for $proto_file"
        exit 1
    fi

    echo "Successfully generated code for $proto_file"
done

echo "All proto files processed successfully."