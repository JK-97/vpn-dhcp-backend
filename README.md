# DNS Backend

- [DNS Backend](#dns-backend)
  - [DNS](#dns)
    - [添加/覆盖 DNS 解析记录](#%e6%b7%bb%e5%8a%a0%e8%a6%86%e7%9b%96-dns-%e8%a7%a3%e6%9e%90%e8%ae%b0%e5%bd%95)
    - [删除 DNS 解析记录](#%e5%88%a0%e9%99%a4-dns-%e8%a7%a3%e6%9e%90%e8%ae%b0%e5%bd%95)
    - [删除 DNS 解析记录](#%e5%88%a0%e9%99%a4-dns-%e8%a7%a3%e6%9e%90%e8%ae%b0%e5%bd%95-1)

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
