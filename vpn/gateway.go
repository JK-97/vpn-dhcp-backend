package vpn

import (
	"errors"
	"sort"
)

// 常见错误
var (
	ErrMissingHost        = errors.New("Missing `host`")
	ErrMissingType        = errors.New("Missing `type`")
	ErrNoGatewayAvailable = errors.New("No Gateway Available")
)

// Gateway VPN 网关的信息
type Gateway struct {
	Host        string `json:"host"`
	Type        string `json:"type"`
	ClientCount int    `json:"count"`
}

// SimpleGatewaySlice 按客户端数量进行简单比较的切片
type SimpleGatewaySlice []*Gateway

func (a SimpleGatewaySlice) Len() int           { return len(a) }
func (a SimpleGatewaySlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SimpleGatewaySlice) Less(i, j int) bool { return a[i].ClientCount < a[j].ClientCount }

// GatewaySelector 选择合适的 VPN 网关
type GatewaySelector interface {
	// SelectGateway 选择合适的 VPN 网关
	SelectGateway(vpnType string) []*Gateway
}

// GatewayStorage 保存 VPN 网关信息
type GatewayStorage interface {
	// Storage 保存 VPN 网关信息
	Storage(*Gateway) error
}

// GatewayClient 操作 VPN Gateway 信息
type GatewayClient interface {
	GatewaySelector
	GatewayStorage
}

// SortGateway 按客户端数量从小到大排序
func SortGateway(gwSlice []*Gateway) {
	sort.Sort(SimpleGatewaySlice(gwSlice))
}
