package dns

// import "github.com/miekg/dns"
// 默认设置
const (
	DefaultTTL      uint32 = 360
	DefaultPort     int    = 80
	DefaultPriority int    = 10
)

// Record DNS 记录
type Record struct {
	ID        uint   `json:"id,omitempty"`
	DomainID  int64  `json:"domain_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	Host      string `json:"host,omitempty"`
	Port      int    `json:"port,omitempty"`
	TTL       uint32 `json:"ttl,omitempty"`
	Priority  int    `json:"priority,omitempty"`
	Weight    int    `json:"weight,omitempty"`
	ChangDate int    `json:"chang_date,omitempty"`
}

// Agent 处理 DNS 记录
type Agent interface {
	// AddRecord 添加 DNS 解析记录
	AddRecord(r Record) error
	// RemoveRecord 移除 DNS 解析记录
	RemoveRecord(r Record) error
	// ModifyRecord 修改 DNS 解析记录
	ModifyRecord(o, n Record) error
}
