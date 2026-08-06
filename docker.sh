#!/bin/bash

IMAGE="mock:1.0.0"
CONTAINER_NAME="mock-server"

usage() {
    echo "Usage: ./docker.sh <start|stop|restart|logs|remove> [-f <host_config_file>] [-p <host_port>...]"
    echo ""
    echo "Commands:"
    echo "  start   启动 Mock 容器"
    echo "  stop    停止 Mock 容器"
    echo "  restart 重启 Mock 容器"
    echo "  logs    查看容器日志"
    echo "  remove  停止并删除容器"
    echo ""
    echo "Options:"
    echo "  -f      宿主机配置文件路径 (start/remove 需要)"
    echo "  -p      宿主机映射端口 (默认从配置文件中提取，可指定多个)"
    echo ""
    echo "Note:"
    echo "  日志挂载到 ./logs 目录"
    echo ""
    echo "Examples:"
    echo "  ./docker.sh start -f ./mock.json"
    echo "  ./docker.sh start -f ./mock.json -p 6666 -p 8080"
    echo "  ./docker.sh stop"
    echo "  ./docker.sh remove"
}

# 解析配置文件的端口列表
get_ports() {
    local config_file=$1
    grep -o '"serverPort"[[:space:]]*:[[:space:]]*[0-9]*' "$config_file" | sed 's/[^0-9]//g'
}

start() {
    local port_mappings=""
    local i=0

    # 构建端口映射参数
    for container_port in $CONTAINER_PORTS; do
        local host_port=$(echo "$HOST_PORTS" | cut -d',' -f$((i+1)))
        host_port=${host_port:-$container_port}
        port_mappings="$port_mappings -p ${host_port}:${container_port}"
        i=$((i + 1))
    done

    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
            echo "容器 ${CONTAINER_NAME} 已在运行中"
            return
        fi
        echo "启动已有容器 ${CONTAINER_NAME}..."
        docker start ${CONTAINER_NAME}
    else
        echo "创建并启动容器 ${CONTAINER_NAME}..."
        echo "端口映射: $port_mappings"
        echo ""
        echo "执行命令:"
        echo "docker run -d \\"
        echo "  --name ${CONTAINER_NAME} \\"
        echo "  -v ${HOST_CONFIG}:/app/mock.json \\"
        echo "  -v $(pwd)/logs:/app/logs \\"
        echo "  $port_mappings \\"
        echo "  ${IMAGE}"
        echo ""
        # 确保宿主机 logs 目录存在
        mkdir -p logs
        docker run -d \
            --name ${CONTAINER_NAME} \
            -v ${HOST_CONFIG}:/app/mock.json \
            -v $(pwd)/logs:/app/logs \
            $port_mappings \
            ${IMAGE}
    fi
    echo "容器已启动"
}

stop() {
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo "停止容器 ${CONTAINER_NAME}..."
        docker stop ${CONTAINER_NAME}
        echo "容器已停止"
    else
        echo "容器 ${CONTAINER_NAME} 未在运行"
    fi
}

restart() {
    stop
    start
}

logs() {
    docker logs -f ${CONTAINER_NAME}
}

remove() {
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
            echo "停止容器 ${CONTAINER_NAME}..."
            docker stop ${CONTAINER_NAME}
        fi
        echo "删除容器 ${CONTAINER_NAME}..."
        docker rm ${CONTAINER_NAME}
        echo "容器已删除"
    else
        echo "容器 ${CONTAINER_NAME} 不存在"
    fi
}

# 全局变量 (逗号分隔的字符串)
HOST_PORTS=""
CONTAINER_PORTS=""
HOST_CONFIG=""
CMD=""

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        start|stop|restart|logs|remove)
            CMD=$1
            shift
            ;;
        -f)
            HOST_CONFIG="$2"
            shift 2
            ;;
        -p)
            if [[ -z "$HOST_PORTS" ]]; then
                HOST_PORTS="$2"
            else
                HOST_PORTS="$HOST_PORTS,$2"
            fi
            shift 2
            ;;
        *)
            echo "未知参数: $1"
            usage
            exit 1
            ;;
    esac
done

# 验证参数
if [[ "$CMD" == "start" ]]; then
    if [[ -z "$HOST_CONFIG" ]]; then
        echo "错误: -f 参数必需"
        usage
        exit 1
    fi
    if [[ ! -f "$HOST_CONFIG" ]]; then
        echo "错误: 配置文件不存在或不是文件: $HOST_CONFIG"
        exit 1
    fi

    # 从配置文件提取所有端口
    CONTAINER_PORTS=$(get_ports "$HOST_CONFIG")
    if [[ -z "$CONTAINER_PORTS" ]]; then
        echo "错误: 无法从配置文件提取 serverPort"
        exit 1
    fi

    # 统计端口数量
    container_count=$(echo "$CONTAINER_PORTS" | grep -c .)
    echo "检测到 $container_count 个服务端口: $(echo "$CONTAINER_PORTS" | tr '\n' ' ')"

    # 如果未指定宿主机端口，用容器端口补充
    host_count=$(echo "$HOST_PORTS" | tr ',' '\n' | grep -c .)
    if [[ "$host_count" -lt "$container_count" ]]; then
        i=0
        for port in $CONTAINER_PORTS; do
            existing=$(echo "$HOST_PORTS" | cut -d',' -f$((i+1)))
            if [[ -z "$existing" ]]; then
                if [[ -z "$HOST_PORTS" ]]; then
                    HOST_PORTS="$port"
                else
                    HOST_PORTS="$HOST_PORTS,$port"
                fi
            fi
            i=$((i + 1))
        done
    fi
fi

case "$CMD" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    logs)
        logs
        ;;
    remove)
        remove
        ;;
    *)
        usage
        ;;
esac
