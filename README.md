# DHCP Backend

- [DHCP Backend](#dhcp-backend)
  - [编译项目](#%e7%bc%96%e8%af%91%e9%a1%b9%e7%9b%ae)
    - [Prerequisite](#prerequisite)
  - [DNS](#dns)
    - [添加/覆盖 DNS 解析记录](#%e6%b7%bb%e5%8a%a0%e8%a6%86%e7%9b%96-dns-%e8%a7%a3%e6%9e%90%e8%ae%b0%e5%bd%95)
    - [删除 DNS 解析记录](#%e5%88%a0%e9%99%a4-dns-%e8%a7%a3%e6%9e%90%e8%ae%b0%e5%bd%95)
    - [删除 DNS 解析记录](#%e5%88%a0%e9%99%a4-dns-%e8%a7%a3%e6%9e%90%e8%ae%b0%e5%bd%95-1)
  - [设备接入](#%e8%ae%be%e5%a4%87%e6%8e%a5%e5%85%a5)
    - [使用 WorkerID 和 Ticket 换取 Key](#%e4%bd%bf%e7%94%a8-workerid-%e5%92%8c-ticket-%e6%8d%a2%e5%8f%96-key)
    - [初始化 VPN](#%e5%88%9d%e5%a7%8b%e5%8c%96-vpn)
      - [初始化 OpenVPN](#%e5%88%9d%e5%a7%8b%e5%8c%96-openvpn)
      - [初始化 WireGuard](#%e5%88%9d%e5%a7%8b%e5%8c%96-wireguard)
    - [切换 VPN 地址](#%e5%88%87%e6%8d%a2-vpn-%e5%9c%b0%e5%9d%80)
      - [切换 OpenVPN](#%e5%88%87%e6%8d%a2-openvpn)
      - [切换 WireGuard](#%e5%88%87%e6%8d%a2-wireguard)
    - [VPN Agent 汇报服务器状态](#vpn-agent-%e6%b1%87%e6%8a%a5%e6%9c%8d%e5%8a%a1%e5%99%a8%e7%8a%b6%e6%80%81)

## 编译项目

### Prerequisite

- `go >= 1.12`

```shell
go generate main.go
go build
```

## DNS

### 添加/覆盖 DNS 解析记录

URL: `/api/v1/dns`
Method: `POST`

```HTTP
POST /api/v1/dns HTTP/1.1
Host: 10.55.2.114:1054
{
    "name": "jx.skydns.local",
    "host": "127.0.0.1",
    "port": 8080,
    "ttl": 10,
    "priority": 10,
    "weight": 10
}
```

- Body Params

| 参数名     | 类型     | 能否省略/默认值 | 描述                              |
| ---------- | -------- | --------------- | --------------------------------- |
| `name`     | `string` | ❌               | 域名                              |
| `host`     | `string` | ❌               | host 地址                         |
| `port`     | `int`    | `80`            | 服务的端口（`SRV`/`TXT`记录需要） |
| `ttl`      | `int`    | `360`           |
| `priority` | `int`    | `10`            | 优先级                            |
| `weight`   | `int`    | `0`             | 权重                              |

### 删除 DNS 解析记录

URL: `/api/v1/dns`
Method: `DELETE`

```HTTP
DELETE /api/v1/dns HTTP/1.1
Host: 10.55.2.114:1054
{
    "name": "jx.skydns.local"
}
```

- Body Params

| 参数名 | 类型     | 能否省略/默认值 | 描述 |
| ------ | -------- | --------------- | ---- |
| `name` | `string` | ❌               | 域名 |

### 删除 DNS 解析记录

URL: `/api/v1/dns`
Method: `DELETE`

```HTTP
GET /api/v1/dns?name=jx.skydns.local HTTP/1.1
Host: 10.55.2.114:1054
```

- Query Params

| 参数名 | 类型     | 能否省略/默认值 | 描述                     |
| ------ | -------- | --------------- | ------------------------ |
| `name` | `string` | ❌               | 域名，`.` 开头是模糊查找 |

- Response 

```json
{
    "data": {
        "records": [
            {
                "name": "jx.skydns.local",
                "host": "127.0.0.1",
                "port": 8080,
                "priority": 10,
                "weight": 10,
                "text": "127.0.0.1:8080",
                "ttl": 10
            }
        ]
    },
    "desc": "success"
}
```

## 设备接入

### 使用 WorkerID 和 Ticket 换取 Key

URL: `/api/v1/bootstrap`
Method: `POST`

```HTTP
POST /api/v1/bootstrap HTTP/1.1
Content-Type: application/json

{
    "wid": "hello",
    "ticket": "jiangxing123"
}
```

- Body Params

| 参数名   | 类型     | 描述     |
| -------- | -------- | -------- |
| `wid`    | `string` | WorkerID |
| `ticket` | `string` | Ticket   |

- Response

```json
{
    "data": {
        "deadLine": 135596800,
        "key": "AyN1SrtbqbeKyBStaA0TycK64DK0+9UbBDLtM3jx64N+Nrli7Akq5UNiv3NV+Fwa+2UX+0FopNqR7IKQ3wZgZi/tMm6JMfnmxpT1myEKcvLG3JHXb0G4TwoU0sYTF9VYQbeiEje4cENoQc6iGC+FJutDQRI/j2IM2N/hRp3Ds4A=",
        "remainCount": 99999
    },
    "desc": "success"
}
```

| Http 状态码 | 描述                                                       |
| ----------- | ---------------------------------------------------------- |
| `200`       | 操作成功                                                   |
| `401`       | ticket 失效，或不存在                                      |
| `500`       | 服务内部错误，数据库中保存的 Ticket 格式与当前后端格式不符 |

| 字段名        | 类型     | 描述                           |
| ------------- | -------- | ------------------------------ |
| `key`         | `string` | 生成的 Key                     |
| `deadLine`    | `int`    | Unix 时间戳，Ticket 的超时时间 |
| `remainCount` | `int`    | Ticket 的剩余使用次数          |

### 初始化 VPN

#### 初始化 OpenVPN

URL: `/api/v1/openvpn/register`
Method: `POST`

```HTTP
```

#### 初始化 WireGuard

URL: `/api/v1/wg/register`
Method: `POST`

```HTTP
```

### 切换 VPN 地址

#### 切换 OpenVPN

URL: `/api/v1/openvpn/arrange`
Method: `POST`

#### 切换 WireGuard

URL: `/api/v1/wg/arrange`
Method: `POST`

### VPN Agent 汇报服务器状态

URL: `/api/v1/agent/report`
Method: `POST`
