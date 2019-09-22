package etcd

import (
	"bytes"
	"context"
	"dhcp-backend/go-utils/logger"
	"dhcp-backend/vpn"
	"encoding/json"

	"github.com/coreos/etcd/clientv3"
)

// VPNGatewayClient 基于 Etcd 的 VPNGatewayClient 实现
type VPNGatewayClient struct {
	Prefix string
	Client *clientv3.Client
}

func (c *VPNGatewayClient) gatewayToKey(gw *vpn.Gateway) string {
	w := bytes.NewBufferString(c.Prefix)

	w.WriteRune('/')
	w.WriteString(gw.Type)

	w.WriteRune('/')
	w.WriteString(gw.Host)

	return w.String()
}

// SelectGateway 选择合适的 VPN 网关
func (c *VPNGatewayClient) SelectGateway(vpnType string) []*vpn.Gateway {
	resp, err := c.Client.Get(context.Background(), c.Prefix+"/"+vpnType, clientv3.WithPrefix())
	if err != nil {
		logger.Info(err)
		return nil
	}
	if resp.Count > 0 {
		arr := make([]*vpn.Gateway, 0, resp.Count)

		for _, it := range resp.Kvs {
			var gw vpn.Gateway
			err := json.Unmarshal(it.Value, &gw)
			if err == nil {
				arr = append(arr, &gw)
			}
		}

		if len(arr) > 0 {
			vpn.SortGateway(arr)
			return arr
		}
	}

	return nil
}

// Storage 保存 VPN 网关信息
func (c *VPNGatewayClient) Storage(gw *vpn.Gateway) error {
	if gw.Host == "" {
		return vpn.ErrMissingHost
	} else if gw.Type == "" {
		return vpn.ErrMissingType
	}
	p, err := json.Marshal(gw)
	if err != nil {
		return err
	}
	_, err = c.Client.Put(context.Background(), c.gatewayToKey(gw), string(p))

	return err
}
