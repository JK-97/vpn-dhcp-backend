# DHCP Backend

- [DHCP Backend](#dhcp-backend)
  - [编译项目](#%e7%bc%96%e8%af%91%e9%a1%b9%e7%9b%ae)
    - [Prerequisite](#prerequisite)
    - [编译项目](#%e7%bc%96%e8%af%91%e9%a1%b9%e7%9b%ae-1)
  - [部署](#%e9%83%a8%e7%bd%b2)
    - [部署 etcd](#%e9%83%a8%e7%bd%b2-etcd)
      - [下载 etcd](#%e4%b8%8b%e8%bd%bd-etcd)
      - [添加配置到 Supervisor](#%e6%b7%bb%e5%8a%a0%e9%85%8d%e7%bd%ae%e5%88%b0-supervisor)
    - [部署 CoreDNS](#%e9%83%a8%e7%bd%b2-coredns)
      - [下载 CoreDNS](#%e4%b8%8b%e8%bd%bd-coredns)
      - [生成 CoreDNS 配置](#%e7%94%9f%e6%88%90-coredns-%e9%85%8d%e7%bd%ae)
      - [添加配置到 Supervisor](#%e6%b7%bb%e5%8a%a0%e9%85%8d%e7%bd%ae%e5%88%b0-supervisor-1)
    - [部署 MongoDB](#%e9%83%a8%e7%bd%b2-mongodb)
    - [部署 DHCP Backend](#%e9%83%a8%e7%bd%b2-dhcp-backend)
      - [生成公私钥对](#%e7%94%9f%e6%88%90%e5%85%ac%e7%a7%81%e9%92%a5%e5%af%b9)
      - [生成 DHCP Backend 配置](#%e7%94%9f%e6%88%90-dhcp-backend-%e9%85%8d%e7%bd%ae)
      - [添加配置到 Supervisor](#%e6%b7%bb%e5%8a%a0%e9%85%8d%e7%bd%ae%e5%88%b0-supervisor-2)
    - [启动服务](#%e5%90%af%e5%8a%a8%e6%9c%8d%e5%8a%a1)
      - [可选](#%e5%8f%af%e9%80%89)
      - [更新 Supervisor 配置](#%e6%9b%b4%e6%96%b0-supervisor-%e9%85%8d%e7%bd%ae)
    - [添加 VPN-AGENT 地址](#%e6%b7%bb%e5%8a%a0-vpn-agent-%e5%9c%b0%e5%9d%80)
    - [添加 Ticket](#%e6%b7%bb%e5%8a%a0-ticket)
  - [DNS](#dns)
    - [添加/覆盖 DNS 解析记录](#%e6%b7%bb%e5%8a%a0%e8%a6%86%e7%9b%96-dns-%e8%a7%a3%e6%9e%90%e8%ae%b0%e5%bd%95)
    - [删除 DNS 解析记录](#%e5%88%a0%e9%99%a4-dns-%e8%a7%a3%e6%9e%90%e8%ae%b0%e5%bd%95)
    - [删除 DNS 解析记录](#%e5%88%a0%e9%99%a4-dns-%e8%a7%a3%e6%9e%90%e8%ae%b0%e5%bd%95-1)
  - [设备接入](#%e8%ae%be%e5%a4%87%e6%8e%a5%e5%85%a5)
    - [使用 WorkerID 和 Ticket 换取 Key](#%e4%bd%bf%e7%94%a8-workerid-%e5%92%8c-ticket-%e6%8d%a2%e5%8f%96-key)
    - [添加 VPN Master](#%e6%b7%bb%e5%8a%a0-vpn-master)
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
```
### 编译项目

```shell
go build
```

在 Windows 上交叉编译

```shell
./build.ps1
```

## 部署

`etcd`、 `MongoDB` 可以只在一台机器上部署，不需要部署为集群。

约定

1. 项目和依赖部署在 `/data-dhcp` 目录下
2. etcd 数据持久化在 `/data-dhcp/etcd-data`
3. etcd 监听端口为 `12379`
4. 机器 DNS 地址 为 `114.114.114.114:53`
5. MongoDB 使用 `mongodb://127.0.0.1:27017/dhcp`
6. 使用 `Supervisor` 管理服务，并且配置文件为 `/etc/supervisor/conf.d/dhcp.conf`
7. 每个 VPN-AGENT 监听 `52100` 端口
8. 每个 VPN-AGENT 机器上的 `30998` 端口，都能转发到对应的 Device Manager，且注册设备的 Path 为 `/api/v1/worker/register`。对应配置为 `RegisterPath = ":30998/api/v1/worker/register"`
9. 设备已安装 `curl`、`supervisor`

### 部署 etcd

`etcd >= 3.4.3`

#### 下载 etcd

```shell
ETCD_VER=v3.4.3

# choose either URL
GOOGLE_URL=https://storage.googleapis.com/etcd
GITHUB_URL=https://github.com/etcd-io/etcd/releases/download
DOWNLOAD_URL=${GITHUB_URL}

rm -f /data-dhcp/etcd-${ETCD_VER}-linux-amd64.tar.gz
mkdir -p /data-dhcp/etcd-data

curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /data-dhcp/etcd-${ETCD_VER}-linux-amd64.tar.gz
tar xzvf /data-dhcp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /data-dhcp/ --strip-components=1
rm -f /data-dhcp/etcd-${ETCD_VER}-linux-amd64.tar.gz

/data-dhcp/etcd --version
/data-dhcp/etcdctl version
```

#### 添加配置到 `Supervisor`

```shell
echo '[program:auth-etcd]
directory=/data-dhcp
command=/data-dhcp/etcd
environment= ETCD_NAME="s1",ETCD_DATA_DIR="/data-dhcp/etcd-data",ETCD_LISTEN_CLIENT_URLS="http://0.0.0.0:12379",ETCD_LISTEN_PEER_URLS="http://0.0.0.0:12380",ETCD_INITIAL_ADVERTISE_PEER_URLS="http://0.0.0.0:12380",ETCD_ADVERTISE_CLIENT_URLS="http://0.0.0.0:12380",ETCD_INITIAL_CLUSTER="s1=http://0.0.0.0:12380",ETCD_INITIAL_CLUSTER_STATE="new",ETCD_INITIAL_CLUSTER_TOKEN="tkn",ETCD_LOGGER="zap",ETCD_LOG_OUTPUTS="stderr",ETCD_LOG_LEVEL="info"
autostart = true
startsecs = 5
autorestart = true
startretries = 3
stdout_logfile_maxbytes = 100MB
stdout_logfile_backups = 3
stderr_logfile_maxbytes = 100MB
stderr_logfile_backups = 3
stdout_logfile=/data/logs/supervisor/%(program_name)s_stdout.log
stderr_logfile=/data/logs/supervisor/%(program_name)s_stderr.log' > /data-dhcp/dhcp.conf
```

### 部署 CoreDNS

#### 下载 CoreDNS

```shell
COREDNS_VER=1.6.6

mkdir -p /data-dhcp

rm -f /data-dhcp/coredns_${COREDNS_VER}_linux_amd64.tgz

curl -L https://github.com/coredns/coredns/releases/download/v${COREDNS_VER}/coredns_${COREDNS_VER}_linux_amd64.tgz -o /data-dhcp/coredns_${COREDNS_VER}_linux_amd64.tgz
tar xzvf /data-dhcp/coredns_${COREDNS_VER}_linux_amd64.tgz -C /data-dhcp/
rm -f /data-dhcp/coredns_${COREDNS_VER}_linux_amd64.tgz

/data-dhcp/coredns --version
```

#### 生成 CoreDNS 配置

```shell
ETCD_ENDPOINTS='127.0.0.1:12379'
DNS="`cat /etc/resolv.conf |grep nameserver|head -1|awk '{print $2}'`:53"

echo "iotedge {
    etcd . {
        path /skydns
        endpoint http://$ETCD_ENDPOINTS
    }
}
worker {
    etcd . {
        path /skydns
        endpoint http://$ETCD_ENDPOINTS
	}
}
master {
    etcd . {
        path /skydns
        endpoint http://$ETCD_ENDPOINTS
    }
}
. { 
    forward . $DNS
}" > /data-dhcp/Corefile
```

#### 添加配置到 `Supervisor`

```shell
echo '[program:auth-coredns]
directory=/data-dhcp
command=/data-dhcp/coredns -conf /data-dhcp/Corefile
autostart = true
startsecs = 5
autorestart = true
startretries = 3
stdout_logfile_maxbytes = 100MB
stdout_logfile_backups = 3
stderr_logfile_maxbytes = 100MB
stderr_logfile_backups = 3
stdout_logfile=/data/logs/supervisor/%(program_name)s_stdout.log
stderr_logfile=/data/logs/supervisor/%(program_name)s_stderr.log' >> /data-dhcp/dhcp.conf
```

### 部署 MongoDB

略

### 部署 DHCP Backend

将二进制拷贝到 `/data-dhcp/dhcp-bakcend`

#### 生成公私钥对

如果已生成密钥，则将公私钥分别拷贝为 `/data-dhcp/public.pem`、`/data-dhcp/private.pem`

```shell
openssl genrsa -out /data-dhcp/private.pem 1024
openssl rsa -in /data-dhcp/private.pem -pubout -out /data-dhcp/public.pem
```

#### 生成 DHCP Backend 配置

```shell
MONGO='mongodb://127.0.0.1:27017/dhcp'

echo 'PrivateKey = """' > /data-dhcp/dhcp-backend.cfg
cat /data-dhcp/private.pem >> /data-dhcp/dhcp-backend.cfg
echo '"""' >> /data-dhcp/dhcp-backend.cfg

echo 'PublicKey = """' >> /data-dhcp/dhcp-backend.cfg
cat /data-dhcp/public.pem >> /data-dhcp/dhcp-backend.cfg
echo '"""' >> /data-dhcp/dhcp-backend.cfg

echo "MongoURI = \"${MONGO}\"
Endpoints = [
    \"${ETCD_ENDPOINTS}\",
]
DNSPrefix = \"/skydns\"
WorkerPrefix = \"/wk\"
GatewayPrefix = \"/gw\"
RegisterPath = \":30998/api/v1/worker/register\"
VPNAgentPort = 52100" >> /data-dhcp/dhcp-backend.cfg
```

#### 添加配置到 `Supervisor`

```shell
echo '[program:auth-backend]
directory=/data-dhcp
command=/data-dhcp/dhcp-backend
autostart = true
startsecs = 5
autorestart = true
startretries = 3
stdout_logfile_maxbytes = 100MB
stdout_logfile_backups = 3
stderr_logfile_maxbytes = 100MB
stderr_logfile_backups = 3
stdout_logfile=/data/logs/supervisor/%(program_name)s_stdout.log
stderr_logfile=/data/logs/supervisor/%(program_name)s_stderr.log' >> /data-dhcp/dhcp.conf
```

### 启动服务

#### 可选

把三个程序添加为一个 group

机器上部署有 etcd 时

```shell
echo '[group:auth]
programs=auth-etcd,auth-coredns,auth-backend
priority=999' >> /data-dhcp/dhcp.conf
```

机器上没有部署 etcd 时

```shell
echo '[group:auth]
programs=auth-coredns,auth-backend
priority=999' >> /data-dhcp/dhcp.conf
```

#### 更新 `Supervisor` 配置

```shell
mkdir -p /data/logs/supervisor/
ln -s /data-dhcp/dhcp.conf /etc/supervisor/conf.d/dhcp.conf
supervisorctl update
```

### 添加 VPN-AGENT 地址

- Key: `/gw/{vpn}/{ip}`
  - `vpn`:
    - `openvpn`: OpenVPN
    - `wg`: WireGuard

```shell
export ETCDCTL_ENDPOINTS="http://127.0.0.1:12379"

/data-dhcp/etcdctl put "/gw/openvpn/10.56.1.4" '{"host":"10.56.1.4","type":"openvpn","count":0}'
/data-dhcp/etcdctl put "/gw/openvpn/10.56.1.5" '{"host":"10.56.1.5","type":"openvpn","count":0}'
/data-dhcp/etcdctl put "/gw/wg/10.56.1.6" '{"host":"10.56.1.6","type":"wg","count":0}'
```

### 添加 Ticket

示例1：添加一个能使用 `9999` 次的 Key

```javascript
use dhcp;
db.boorstrapTickt.insertOne({ 
    "_id" : "jiangxing123", 
    "remainCount" : NumberInt(9999)
});
```

示例2：添加一个只能 WorkerID: `J01f594430`的设备使用，且 `2019-12-23T09:55:10.841+0000` 过期的 Key

```javascript
use dhcp;
db.boorstrapTickt.insertOne({ 
    "_id" : "j2pjvn", 
    "remainCount" : NumberInt(999), 
    "wid" : "J01f594430", 
    "deadLine" : ISODate("2019-12-23T09:55:10.841+0000")
});
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

### 添加 VPN Master

```shell
etcdctl put "/gw/openvpn/10.53.1.218" '{"host":"10.53.1.218","type":"openvpn","count":0}'
etcdctl put "/gw/wg/10.53.1.218" '{"host":"10.53.1.218","type":"wg","count":0}'
etcdctl put "/gw/wg/10.53.1.219" '{"host":"10.53.1.219","type":"wg","count":0}'
etcdctl put "/gw/openvpn/10.53.1.220" '{"host":"10.53.1.220","type":"openvpn","count":0}'
```

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
