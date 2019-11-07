//+build ignore

package rdb

import (
	"database/sql"
	"errors"
	"time"

	bk "gitlab.jiangxingai.com/edgenode/dhcp-backend"

	"github.com/jinzhu/gorm"
	pdnsmodel "github.com/lrh3321/wps/pdns/model"

	// mysql driver
	_ "github.com/jinzhu/gorm/dialects/mysql"

	// sqlite driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	// postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// DatabaseDNSAgent 在DB中添加/删除 DNS 记录
type DatabaseDNSAgent struct {
	*gorm.DB
}

// AddRecord 添加 DNS 解析记录
func (a *DatabaseDNSAgent) AddRecord(r bk.DNSRecord) error {
	if r.TTL == 0 {
		r.TTL = bk.DefaultTTL
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
func (a *DatabaseDNSAgent) RemoveRecord(r bk.DNSRecord) error {
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
func (a *DatabaseDNSAgent) ModifyRecord(o, n bk.DNSRecord) error {

	return nil
}
