# Kubernetes 网络 - Agent 使用指南

> **网络ID**: k8s-network  
> **版本**: 1.0.0  
> **标签**: Kubernetes, 拓扑架构, 运维

## 网络概览

本知识网络描述 Kubernetes 集群的核心资源拓扑结构，采用多文件组织方式。

### 核心对象

| 对象 | 文件路径 | 说明 |
|------|----------|------|
| Pod | `object_types/objects.bkn` | 最小部署单元 |
| Node | `object_types/objects.bkn` | 集群工作节点 |
| Service | `object_types/objects.bkn` | 服务暴露与负载均衡 |

### 核心关系

| 关系 | 文件路径 | 说明 |
|------|----------|------|
| pod_belongs_node | `relation_types/relations.bkn` | Pod 归属节点关系 |
| service_routes_pod | `relation_types/relations.bkn` | 服务路由到 Pod |

### 可用行动

| 行动 | 文件路径 | 说明 |
|------|----------|------|
| restart_pod | `action_types/actions.bkn` | 重启指定 Pod |
| cordon_node | `action_types/actions.bkn` | 隔离节点，禁止调度 |

## 拓扑结构

```
┌─────────────┐     routes      ┌─────────┐
│  Service    │ ───────────────→│   Pod   │
└─────────────┘                 └────┬────┘
                                     │ belongs
                                     ↓
                                ┌─────────┐
                                │  Node   │
                                └─────────┘
```

## 使用建议

### 查询场景

1. **获取所有对象定义**
   - 读取 `object_types/objects.bkn`

2. **查找关系定义**
   - 读取 `relation_types/relations.bkn`

### 运维场景

1. **执行运维操作**
   - 查看 `action_types/actions.bkn` 中的行动定义
   - 了解触发条件和参数绑定

## 索引表

### 按类型索引

- **对象定义**: `object_types/objects.bkn`
- **关系定义**: `relation_types/relations.bkn`
- **行动定义**: `action_types/actions.bkn`

### 按功能索引

- **资源管理**: Pod, Node, Service
- **网络路由**: service_routes_pod
- **调度归属**: pod_belongs_node
- **运维操作**: restart_pod, cordon_node

## 注意事项

1. 本示例采用单文件多定义的组织方式
2. 所有定义集中存储在各自的类型文件中
3. 适合中小型知识网络的管理
