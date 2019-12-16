#!/usr/bin/env bash

apt install -y curl supervisor

COREDNS_VER=1.6.6
ETCD_ENDPOINTS='127.0.0.1:12379'
MONGO='mongodb://127.0.0.1:27017/dhcp'

mkdir -p /data-dhcp

rm -f /data-dhcp/coredns_${COREDNS_VER}_linux_amd64.tgz

curl -L https://github.com/coredns/coredns/releases/download/v${COREDNS_VER}/coredns_${COREDNS_VER}_linux_amd64.tgz -o /data-dhcp/coredns_${COREDNS_VER}_linux_amd64.tgz
tar xzvf /data-dhcp/coredns_${COREDNS_VER}_linux_amd64.tgz -C /data-dhcp/
rm -f /data-dhcp/coredns_${COREDNS_VER}_linux_amd64.tgz

/data-dhcp/coredns --version


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
stderr_logfile=/data/logs/supervisor/%(program_name)s_stderr.log' > /data-dhcp/dhcp.conf

# DHCP Backend

chmod +x /data-dhcp/dhcp-backend

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

echo '[group:auth]
programs=auth-coredns,auth-backend
priority=999' >> /data-dhcp/dhcp.conf

mkdir -p /data/logs/supervisor/
ln -s /data-dhcp/dhcp.conf /etc/supervisor/conf.d/dhcp.conf
supervisorctl update
