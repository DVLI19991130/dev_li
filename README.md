# Mock

高性能接口 Mock CLI 工具，用于项目接口性能压测时对目标接口进行 Mock 或代理，支持通过 `mock.json` 文件进行动态配置。

## 功能特性

- **多协议支持**: HTTP、Dubbo
- **双服务模式**: Mock（模拟响应）、Proxy（请求代理）
- **注册中心**: 支持 Zookeeper、Nacos 服务注册（仅 Dubbo）
- **动态响应**: 支持流水号、时间戳、UUID、随机数等动态值生成
- **负载均衡**: 权重随机算法，支持多后端服务
- **可配置延迟**: 支持设置接口响应延迟

## 快速开始

### 安装

```bash
# 安装依赖
go mod tidy

# 构建并安装到 $GOPATH/bin
go install

# 或使用 make 构建
make build
```

### CLI 命令

```bash
mock init -s <serverMode> -p <protocol> [-r <registry>]  # 初始化配置
mock check -f mock-http.json                                 # 验证配置文件
mock run -f mock-http.json                                   # 启动 Mock 服务
```

**init 命令参数**:
- `-s, --serverMode`: 服务模式 (`mock`|`proxy`)，必填
- `-p, --protocol`: 协议类型 (`http`|`dubbo`)，必填
- `-r, --registry`: 注册中心类型 (`nacos`|`zookeeper`)，可选，仅 dubbo 支持

### HTTP Mock 示例

```json
{
    "serverName": "mock-server",
    "serverPort": 6666,
    "serverMode": "mock",
    "protocol": "http",
    "api": [
        {
            "url": "/api/user",
            "method": "GET",
            "delay": 0,
            "response": {
                "code": "0000",
                "message": "success",
                "data": {
                    "userId": 10001,
                    "username": "test"
                }
            }
        },
        {
            "url": "/api/order/create",
            "method": "POST",
            "delay": 1000,
            "response": {
                "code": "0000",
                "orderNo": "$(flowNo)"
            }
        }
    ]
}
```

### HTTP Proxy 示例

```json
{
    "serverName": "proxy-server",
    "serverPort": 6666,
    "serverMode": "proxy",
    "protocol": "http",
    "api": [
        {
            "url": "/api/user",
            "method": "GET",
            "backend": {
                "addr": [
                    "127.0.0.1:8081",
                    "127.0.0.1:8082"
                ],
                "weight": [100, 100]
            },
            "delay": -1
        }
    ]
}
```

### Dubbo Mock 示例

```json
{
    "serverName": "dubbo-mock-server",
    "serverPort": 6668,
    "serverMode": "mock",
    "protocol": "dubbo",
    "registry": {
        "zookeeper": {
            "addr": ["127.0.0.1:2181"]
        }
    },
    "api": [
        {
            "interface": "com.example.service.UserService",
            "version": "1.0.0",
            "group": "",
            "serialization": "hessian2",
            "registry": "zookeeper",
            "methods": [
                {
                    "name": "getUser",
                    "delay": -1,
                    "paramType": ["java.lang.String"],
                    "response": {
                        "userId": "10001",
                        "username": "test"
                    },
                    "responseTypes": {
                        "userId": "string",
                        "username": "string"
                    }
                }
            ]
        }
    ]
}
```

## 配置说明

### 基础配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| serverName | string | 是 | 服务名称 |
| serverPort | int | 是 | 服务端口 (1-65535) |
| serverMode | string | 是 | 服务模式: `mock` 或 `proxy` |
| protocol | string | 是 | 协议类型: `http` 或 `dubbo` |
| registry | object | 否 | 注册中心配置 |
| api | array | 是 | API 接口配置 |

### 注册中心配置

```json
"registry": {
    "zookeeper": {
        "addr": ["127.0.0.1:2181"]
    },
    "nacos": {
        "addr": ["127.0.0.1:8848"],
        "namespace": "public"
    }
}
```

### 后端服务配置 (Proxy 模式)

```json
"backend": {
    "addr": ["127.0.0.1:8081", "127.0.0.1:8082"],
    "weight": [100, 100]
}
```

## 动态响应

支持在响应中使用 `$(funcName,args...)` 语法调用动态生成函数：

| 函数 | 说明 | 示例 |
|------|------|------|
| `$(flowNo)` | 流水号 (格式: yyMMddHHmmss + 8位序号) | `$(flowNo)` → `26041510385954006127` |
| `$(timestamp)` | 当前时间戳 (秒) | `$(timestamp)` → `1713153538` |
| `$(uuid)` | UUID v4 | `$(uuid)` → `550e8400-e29b-41d4-a716-446655440000` |
| `$(random)` | 随机整数 | `$(random)` → `48291` |

## 延迟配置

| 值    | 说明 |
|------|------|
| -1   | 不延迟，立即响应 |
| 0    | 不延迟 |
| \> 0 | 延迟指定毫秒数 |

## 项目结构

```
mock/
├── main.go                    # 程序入口
├── cmd/                       # CLI 命令 (cobra)
├── pkg/                       # 公共工具
├── internal/
│   ├── conf/                  # 配置加载与验证
│   ├── server/                # 服务调度
│   ├── protocol/              # 协议处理器
│   │   ├── http/             # HTTP 协议 (fiber)
│   │   └── dubbo/            # Dubbo 协议 (dubbo-go)
│   ├── registry/             # 注册中心客户端
│   ├── slb/                  # 负载均衡
│   ├── dynamic/              # 动态响应生成
│   └── logger/               # 日志组件
```

## 构建

```bash
# 全平台构建 (linux/macos_amd/macos_arm)
make build

# 特定平台
make build linux      # Linux
make build macos_amd  # Mac Intel
make build macos_arm  # Mac ARM

# Docker 镜像构建
make docker
```

## 测试

```bash
go test ./...
```
