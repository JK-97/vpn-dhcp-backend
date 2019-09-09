package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	// mysql driver
	_ "github.com/jinzhu/gorm/dialects/mysql"

	// sqlite driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	// postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	errMissingName = errors.New("Missing `name`")
	errMissingHost = errors.New("Missing `host`")
)

// DNSBackend 在DB中添加/删除 DNS 记录
type DNSBackend struct {
	Agent DNSAgent
}

func (b *DNSBackend) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func readRecordFromRequest(w http.ResponseWriter, r *http.Request) (rec DNSRecord, err error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
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

	err = b.Agent.AddRecord(rec)
	if err != nil {
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
	case SkyDNSAgent:
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

		result := b.Agent.(SkyDNSAgent).FindSkyDNS(name, prefix)
		WriteData(w, &map[string]interface{}{"records": result})

	default:
		ErrorWithCode(w, http.StatusNotImplemented)
	}

}

// VirtualNetworkBackend 管理虚拟网络
type VirtualNetworkBackend struct {
	PrivateKey *rsa.PrivateKey // 私钥
}

// Verify 验证请求是否合法
func (b *VirtualNetworkBackend) Verify(w http.ResponseWriter, r *http.Request) (buf []byte, err error) {
	buf, err = ioutil.ReadAll(r.Body)
	if err != nil {
		Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// // b.PrivateKey.Decrypt("", )
	// b.PrivateKey.Precompute()
	// publickey := b.PrivateKey.Public()

	// msg := []byte("The secret message!")

	// encryptedmsg, err := rsa.EncryptPKCS1v15(rand.Reader, publickey.(*rsa.PublicKey), msg)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	buf, err = rsa.DecryptPKCS1v15(rand.Reader, b.PrivateKey, buf)
	if err != nil {
		// TODO
		Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	return
}
