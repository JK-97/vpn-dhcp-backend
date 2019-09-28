package vpn

// VPN 类型
const (
	VpnNone      = "none"
	VpnOpenvpn   = "openvpn"
	VpnWireguard = "wg"
)

// WorkerIPStatus 设备 IP 的状态
type WorkerIPStatus struct {
	Addr       string `json:"addr"`
	Type       string `json:"type"`
	Gateway    string `json:"gateway"`
	Master     string `json:"master"`
	RegisterIP string `json:"regIP,omitempty"` // 注册 vpn 时使用的 IP
	Disabled   bool   `json:"disabled,omitempty"`
}

// WorkerIP 设备的 IP 地址
type WorkerIP struct {
	ID         string                     `json:"id"`
	Status     map[string]*WorkerIPStatus `json:"status"`
	Registered bool                       `json:"registered"` // 设备是否已注册
}

// WorkerClient 获取设备信息
type WorkerClient interface {
	// AddIP 添加设备的 IP
	AddIP(workerID string, ip WorkerIPStatus) error
	// GetIP 获取设备的 IP
	GetIP(workerID string) WorkerIP
	// RemoveIP 移除设备的 IP
	RemoveIP(workerID string, ip WorkerIPStatus) error
}
