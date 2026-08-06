# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 语言描述

* 所有对话、文档、openspec文档均使用**简体中文**。
* 代码注释、错误描述采用英文

## 项目描述

**mock** 是一个基于Go语言实现的高性能接口Mock Cli工具，用于项目接口性能压测时对目标接口进行Mock或代理，可通过`mock.json`文件进行动态配置。

## 常用命令

```bash
# 安装依赖
go mod tidy

# 构建并安装到 $GOPATH/bin
go install

# 构建特定平台程序
make build linux      # Linux
make build macos_amd  # Mac Intel
make build macos_arm  # Mac ARM
make build            # 全平台构建 (linux/macos_amd/macos_arm)
make docker           # Docker 镜像构建

# 运行测试
go test ./...

# CLI命令
mock init -s <serverMode> -p <protocol> [-r <registry>]  # 初始化配置
mock check -f mock-http.json                               # 验证配置文件
mock run -f mock-http.json                                 # 运行Mock服务
```

**init 命令参数**:
- `-s, --serverMode`: 服务模式 (`mock`|`proxy`)，必填
- `-p, --protocol`: 协议类型 (`http`|`dubbo`)，必填
- `-r, --registry`: 注册中心类型 (`nacos`|`zookeeper`)，可选，仅 dubbo 支持

## 架构

### 目录结构

```
mock/
├── main.go                    # 程序入口
├── cmd/                       # CLI命令入口 (cobra)
│   ├── root.go               # 根命令
│   ├── run.go                # run命令: 启动服务
│   ├── check.go              # check命令: 验证配置
│   └── init.go               # init命令: 初始化配置
│
├── pkg/                       # 公共工具包
│   ├── ip.go                # IP地址工具
│   └── trace_id.go          # 链路追踪ID生成
│
├── internal/
│   ├── conf/                 # 配置加载与验证
│   │   ├── conf.go          # 配置加载、验证逻辑
│   │   ├── conf_test.go    # 配置测试
│   │   └── model.go        # 配置结构体定义
│   │
│   ├── server/               # 服务调度
│   │   ├── server.go       # StartServer() 分发
│   │   ├── mock/           # Mock服务实现
│   │   │   └── mock.go
│   │   └── proxy/          # Proxy服务实现
│   │       └── proxy.go
│   │
│   ├── protocol/            # 协议处理器
│   │   ├── http/           # HTTP协议 (fiber)
│   │   │   ├── http_server.go
│   │   │   ├── http_proxy.go
│   │   │   └── middleware/
│   │   │       └── logging.go
│   │   └── dubbo/          # Dubbo协议 (dubbo-go)
│   │       ├── dubbo_server.go
│   │       ├── mock_codec.go
│   │       └── middleware/
│   │           └── logging.go
│   │
│   ├── registry/           # 注册中心
│   │   └── nacos/          # Nacos注册中心客户端
│   │       └── nacos.go
│   │
│   ├── slb/                # 负载均衡
│   │   ├── weighted_random.go
│   │   └── weighted_random_test.go
│   │
│   ├── dynamic/            # 动态响应生成
│   │   ├── generator.go
│   │   ├── copy.go
│   │   ├── copy_test.go
│   │   ├── generator_test.go
│   │   └── funcs/          # 内置动态函数
│   │       ├── flow_No.go
│   │       ├── random.go
│   │       ├── timestamp.go
│   │       └── uuid.go
│   │
│   └── logger/             # 日志组件
│       ├── logger.go
│       ├── async_writer.go
│       └── zerolog_adapter.go
```│
│   └── logger/             # 日志组件
│       ├── logger.go
│       ├── async_writer.go
│       └── zerolog_adapter.go
```

### 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI 层                                   │
│                     cmd/ (cobra)                                 │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        配置 层                                   │
│                 internal/conf/                                  │
│              conf.Load() → Validate()                           │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        服务 层                                   │
│              server.StartServer()                                │
│                      │                                           │
│         ┌────────────┴────────────┐                             │
│         ▼                         ▼                              │
│    ┌─────────┐              ┌─────────┐                         │
│    │  mock   │              │  proxy  │                         │
│    └────┬────┘              └────┬────┘                         │
└─────────┼────────────────────────┼──────────────────────────────┘
          │                        │
          ▼                        ▼
┌─────────────────────────────────────────────────────────────────┐
│                       协议 层                                    │
│         ┌────────────┬────────────┐                             │
│         ▼            ▼            ▼                              │
│    ┌─────────┐  ┌─────────┐  ┌─────────┐                        │
│    │  HTTP   │  │   TCP   │  │  Dubbo  │                        │
│    │ (fiber) │  │  (net)  │  │(dubbo-go)                        │
│    └─────────┘  └─────────┘  └─────────┘                        │
└─────────────────────────────────────────────────────────────────┘
```

### 启动流程

```
mock run -f xxx.json
    │
    ▼
cmd/run.go
    │
    ▼
conf.Load() ──验证配置──► Config
    │
    ▼
server.StartServer(cfg)
    │
    ├── ServerMode = "mock"
    │       │
    │       ▼
    │   mock.StartMock(cfg)
    │       │
    │       ▼
    │   protocol/http.StartHttpServer()  (当 protocol = http)
    │   protocol/dubbo.StartDubboServer() (当 protocol = dubbo)
    │
    └── ServerMode = "proxy"
            │
            ▼
        proxy.StartProxy(cfg)
            │
            ▼
        protocol/http.StartHttpProxy()  (当 protocol = http)
```

### 核心设计

1. **配置驱动**: `mock.json` 动态配置所有API路由、延迟、请求/响应模板
2. **协议抽象**: 按协议类型分目录实现，配置字段根据 `protocol` 和 `serverMode` 动态验证
3. **注册中心**: 支持 Nacos/Zookeeper 服务注册 (仅Dubbo协议)
4. **Mock vs Proxy**:
   - `mock`: 直接返回配置的响应
   - `proxy`: 转发请求到真实服务，支持权重负载均衡

## 配置结构

`mock.json` 配置文件结构:

```
[
   {
       "serverName": "moc-server", // 服务名
       "serverPort": 8080, // 服务端口，api接口均通过该接口访问
       "serverMode": "mock|proxy", // 服务类型, 支持mock/proxy
       "protocol": "http|dubbo", // api协议, 支持http/dubbo
       "registry": { // 注册中心，支持zookeeper/nacos
           "zookeeper": { // zookeeper
               "addr": ["127.0.0.1:2181","127.0.0.1:2181"]
           },
           "nacos": { // nacos
               "addr": ["127.0.0.1:8848","127.0.0.1:8848"]
               "namespace": "public"
           }
       },
       "api": [ // api接口配置, 根据serverMode和protocol配置不同
           {
               // 参考：API配置
           }
       ]
   }
]
```

### API配置

**HTTP Mock**
```
{
    "serverName": "moc-server",
    "serverPort": 6666,
    "serverMode": "mock",
    "protocol": "http",
    "registry": { // 可选
        "nacos": {
            "addr": [
                "127.0.0.1:8848",
                "127.0.0.1:8848"
            ],
            "namespace": "public"
        }
    },
    "api": [
        {
            "registry": "nacos", // 可选
            "url": "/unionPayCardPass/transSend",
            "method": "GET|POST|...",
            "delay": 1000, // 延迟响应(ms), -1 不延迟
            "response": {
                "resCode": "0000",
                "resMessage": "成功"
            }
        }
    ]
}
```

**Dubbo Mock**
```
{
    "serverName": "moc-server",
    "serverPort": 6668,
    "serverMode": "mock",
    "protocol": "dubbo",
    "registry": {
        "zookeeper": {
            "addr": [
                "127.0.0.1:2181",
                "127.0.0.1:2181"
            ]
        }
    },
    "api": [
        {
            "interface": "com.enfcpay.miss.common.service.interfaces.AccCommonService",
            "version": "1.0.0",
            "group": "",
            "serialization": "hessian2",
            "registry": "zookeeper",
            "methods": [
                {
                    "name": "queryPrice",
                    "delay": -1,
                    "paramType": ["java.lang.String", "java.lang.String", "java.lang.Long", "java.lang.Long", "java.lang.Integer"],
                    "response": {
                        "resCode": "0000",
                        "resMessage": "成功"
                    }
                },
                {
                    "method": "queryPrice",
                    "delay": 1000,
                    "paramType": ["java.lang.String", "java.lang.String", "java.lang.Long", "java.lang.Long"],
                    "response": {
                        "resCode": "0000",
                        "resMessage": "成功"
                    }
                }
            ]
        }
    ]
}
```

**Proxy**
```
# http代理
{
    "serverName": "proxy-server",
    "serverPort": 6666,
    "serverMode": "proxy",
    "protocol": "http",
    "api": [
        {
            "url": "/unionPayCardPass/transSend",
            "method": "GET|POST|...",
            "backend": { // 后端服务配置
                "addr": [
                    "127.0.0.1:8081",
                    "127.0.0.1:8082"
                ],
                "weight": [ // 权重, 与addr顺序一致, 范围[0, 100]
                    100,
                    100
                ]
            },
            "delay": -1
        }
    ]
}
```


## 依赖

```go
// go.mod
module mock

go 1.25.4

// 配置文件验证, 并支持中文翻译
github.com/go-playground/locales
github.com/go-playground/universal-translator
github.com/go-playground/validator/v10
// CLI框架
github.com/spf13/cobra
// HTTP服务
github.com/gofiber/fiber/v3
// HTTP Proxy
github.com/gofiber/fiber/v3/middleware/proxy
// Error
github.com/pkg/errors
// Dubbo 服务
dubbo.apache.org/dubbo-go/v3
// 注册中心客户端
github.com/dubbogo/nacos
github.com/dubbogo/zookeeper
// 日志
github.com/rs/zerolog
```
