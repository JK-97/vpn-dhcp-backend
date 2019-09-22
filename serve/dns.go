package serve

import (
	"dhcp-backend/dns"
	"dhcp-backend/etcd"
	"dhcp-backend/go-utils/logger"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	errMissingName        = errors.New("Missing `name`")
	errMissingHost        = errors.New("Missing `host`")
	errMissingType        = errors.New("Missing `type`")
	errInvalidName        = errors.New("Invalid `name`")
	errNoGatewayAvailable = errors.New("No Gateway Available")
	errKeyNotMatch        = errors.New("Key Not Match")
)

// DNSBackend 在DB中添加/删除 DNS 记录
type DNSBackend struct {
	Agent dns.Agent
}

func (b *DNSBackend) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func readRecordFromRequest(w http.ResponseWriter, r *http.Request) (rec dns.Record, err error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	logger.Info("Body:", string(b))
	if err = json.Unmarshal(b, &rec); err != nil {
		Error(w, err.Error(), http.StatusBadRequest)
	}

	return
}

// AddRecord 添加 DNS 记录
func (b *DNSBackend) AddRecord(w http.ResponseWriter, r *http.Request) {
	rec, err := readRecordFromRequest(w, r)
	if err != nil {
		return
	}
	if rec.Host == "" {
		Error(w, errMissingHost.Error(), http.StatusBadRequest)
		return
	} else if rec.Name == "" {
		Error(w, errMissingName.Error(), http.StatusBadRequest)
		return
	}

	// 域名只接收 数字、字母、'.'、'-'
	for _, char := range rec.Name {
		if char >= '0' && char <= '9' {
			continue
		} else if char >= 'A' && char <= 'Z' {
			continue
		} else if char >= 'a' && char <= 'z' {
			continue
		} else if char == '.' || char == '-' {
			continue
		}
		Error(w, errInvalidName.Error(), http.StatusBadRequest)
		return
	}

	if strings.HasSuffix(rec.Name, ".master") || strings.HasSuffix(rec.Name, ".worker") {
		if strings.Count(rec.Name, ".") < 2 {
			Error(w, "Not Allowed", http.StatusForbidden)
			return
		}
	}

	err = b.Agent.AddRecord(rec)
	if err != nil {
		logger.Info("AddRecord", err)
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteSucess(w)
}

// RemoveRecord 删除 DNS 记录
func (b *DNSBackend) RemoveRecord(w http.ResponseWriter, r *http.Request) {
	rec, err := readRecordFromRequest(w, r)
	if err != nil {
		return
	}
	err = b.Agent.RemoveRecord(rec)
	if err != nil {
		logger.Info("RemoveRecord", err)
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ModifyRecord 修改 DNS 记录
func (b *DNSBackend) ModifyRecord(w http.ResponseWriter, r *http.Request) {

}

// FindRecord 查找 DNS 记录
func (b *DNSBackend) FindRecord(w http.ResponseWriter, r *http.Request) {

	switch b.Agent.(type) {
	case etcd.SkyDNSAgent:
		query := r.URL.Query()
		name := query.Get("name")

		var prefix bool
		if strings.HasPrefix(name, ".") {
			prefix = true
			name = name[1:]
		}
		if name == "" {
			Error(w, errMissingName.Error(), http.StatusBadRequest)
			return
		}

		result := b.Agent.(etcd.SkyDNSAgent).FindSkyDNS(name, prefix)
		WriteData(w, &map[string]interface{}{"records": result})

	default:
		ErrorWithCode(w, http.StatusNotImplemented)
	}

}
