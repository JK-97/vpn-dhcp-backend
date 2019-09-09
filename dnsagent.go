package main

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	pdnsmodel "github.com/lrh3321/wps/pdns/model"
)

// import "github.com/miekg/dns"
const (
	defaultTTL      uint32 = 360
	defaultPort     int    = 80
	defaultPriority int    = 10
)

// DNSRecord DNS 记录
type DNSRecord struct {
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

// DNSAgent 处理 DNS 记录
type DNSAgent interface {
	// AddRecord 添加 DNS 解析记录
	AddRecord(r DNSRecord) error
	// RemoveRecord 移除 DNS 解析记录
	RemoveRecord(r DNSRecord) error
	// ModifyRecord 修改 DNS 解析记录
	ModifyRecord(o, n DNSRecord) error
}

// DatabaseDNSAgent 在DB中添加/删除 DNS 记录
type DatabaseDNSAgent struct {
	*gorm.DB
}

// AddRecord 添加 DNS 解析记录
func (a *DatabaseDNSAgent) AddRecord(r DNSRecord) error {
	if r.TTL == 0 {
		r.TTL = defaultTTL
	}

	model := pdnsmodel.Record{
		// ID:        r.ID,
		DomainID:  sql.NullInt64{Int64: r.DomainID, Valid: r.DomainID > 0},
		Name:      r.Name,
		Type:      r.Type,
		Content:   r.Host,
		TTL:       r.TTL,
		Priority:  r.Priority,
		ChangDate: int(time.Now().Unix()),
	}
	a.DB.NewRecord(&model)

	return nil
}

// RemoveRecord 移除 DNS 解析记录
func (a *DatabaseDNSAgent) RemoveRecord(r DNSRecord) error {
	var errs []error

	if r.ID > 0 {
		errs = a.DB.Delete(pdnsmodel.Record{}, "id = ?", r.ID).GetErrors()
	} else {
		if r.Name == "" {
			return errors.New("Missing `name`")
		}
		errs = a.DB.Delete(pdnsmodel.Record{}, "name = ?", r.Name).GetErrors()
	}

	for _, err := range errs {
		return err
	}

	return nil
}

// ModifyRecord 修改 DNS 解析记录
func (a *DatabaseDNSAgent) ModifyRecord(o, n DNSRecord) error {

	return nil
}
