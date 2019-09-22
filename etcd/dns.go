package etcd

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coreos/etcd/clientv3"

	dns "dhcp-backend/dns"
)

// SkyDNSRecord Skydns 格式存储的 DNS 记录
type SkyDNSRecord struct {
	Name      string `json:"name,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	msg.Service
}

// SkyDNSAgent 返回 SkyDNSAgent 格式的数据
type SkyDNSAgent interface {
	dns.Agent

	// FindSkyDNS 查找 DNS 解析记录
	FindSkyDNS(domain string, prefix bool) []SkyDNSRecord
}

// SkyDNSRecord defines a discoverable service in etcd.
// type SkyDNSRecord msg.Service

// DNSAgent 基于 Etcd
type DNSAgent struct {
	Prefix string
	Client *clientv3.Client
}

// AddRecord 添加 DNS 解析记录
func (a *DNSAgent) AddRecord(r dns.Record) error {
	if r.TTL == 0 {
		r.TTL = dns.DefaultTTL
	}
	if r.Port == 0 {
		r.Port = dns.DefaultPort
	}
	if r.Priority == 0 {
		r.Priority = dns.DefaultPriority
	}

	srv := msg.Service{
		Host: r.Host,
		TTL:  r.TTL,
		// Port:     r.Port,
		// Priority: r.Priority,
		// Text:     net.JoinHostPort(r.Host, strconv.Itoa(r.Port)),
		// Weight:   r.Weight,
		// // Mail: "",
		Key: a.DomainToKey(r.Name),
	}
	switch r.Type {
	case "A", "AAAA":
	case "SRV":
		srv.Port = r.Port
		srv.Priority = r.Priority
	case "TXT":
		srv.Host = ""
		srv.Text = net.JoinHostPort(r.Host, strconv.Itoa(r.Port))
	}

	result, err := json.Marshal(SkyDNSRecord{Service: srv, Timestamp: time.Now().Unix()})
	if err != nil {
		return err
	}

	val := string(result)
	_, err = a.Client.Put(context.Background(), srv.Key, val)

	return err
}

// RemoveRecord 移除 DNS 解析记录
func (a *DNSAgent) RemoveRecord(r dns.Record) error {
	key := a.DomainToKey(r.Name)
	_, err := a.Client.Delete(context.Background(), key, clientv3.WithPrefix())
	return err
}

// ModifyRecord 修改 DNS 解析记录
func (a *DNSAgent) ModifyRecord(o, n dns.Record) error {

	return nil
}

// ModifySubTXTRecord 修改指定的一级域名下所有的 TXT 记录
func (a *DNSAgent) ModifySubTXTRecord(root dns.Record) error {
	key := a.DomainToKey(root.Name)

	resp, err := a.Client.Get(context.Background(), key)
	if err != nil {
		return err
	}
	if resp.Count == 0 {
		a.AddRecord(root)
		return nil
	}

	it := resp.Kvs[0]
	var srv SkyDNSRecord
	if err := json.Unmarshal(it.Value, &srv); err == nil {
		srv.Name = a.KeyToDomain(string(it.Key))
		if srv.Host == root.Host {
			// IP 地址未变动
			return nil
		}
		a.AddRecord(root)
	}
	result := make([]SkyDNSRecord, 0)
	resp, _ = a.Client.Get(context.Background(), key, clientv3.WithPrefix())
	if resp != nil {
		for _, it := range resp.Kvs {
			var srv SkyDNSRecord
			if err := json.Unmarshal(it.Value, &srv); err == nil {
				if srv.Text != "" {
					continue
				}
				srv.Key = string(it.Key)
				srv.Name = a.KeyToDomain(string(it.Key))
				_, p, err := net.SplitHostPort(srv.Text)
				if err != nil {
					continue
				}
				srv.Text = net.JoinHostPort(root.Host, p)
				result = append(result, srv)
			}
		}
	}
	for _, it := range result {
		val, err := json.Marshal(it)
		if err == nil {
			continue
		}
		a.Client.Put(context.Background(), it.Key, string(val))
	}

	return nil
}

// FindSkyDNS 查找 DNS 解析记录
func (a *DNSAgent) FindSkyDNS(domain string, prefix bool) []SkyDNSRecord {
	result := make([]SkyDNSRecord, 0)
	key := a.DomainToKey(domain)
	var resp *clientv3.GetResponse

	if prefix {
		resp, _ = a.Client.Get(context.Background(), key, clientv3.WithPrefix())
	} else {
		resp, _ = a.Client.Get(context.Background(), key)
	}
	if resp != nil {
		for _, it := range resp.Kvs {
			var srv SkyDNSRecord
			if err := json.Unmarshal(it.Value, &srv); err == nil {
				srv.Name = a.KeyToDomain(string(it.Key))
				result = append(result, srv)
			}
		}
	}
	return result
}

// DomainToKey Translate Domain to Key in Etcd
func (a *DNSAgent) DomainToKey(domain string) string {
	domain = strings.ToLower(domain) // 域名必须是全小写
	keys := strings.Split(domain, ".")

	length := len(keys)

	w := bytes.NewBufferString(a.Prefix)
	for index := length - 1; index >= 0; index-- {
		s := keys[index]
		if s == "" {
			continue
		}
		w.WriteRune('/')
		w.WriteString(strings.ToLower(keys[index]))
	}
	w.WriteRune('/')
	return w.String()
}

// KeyToDomain Translate Key in Etcd to Domain
func (a *DNSAgent) KeyToDomain(key string) string {

	key = strings.Trim(key[len(a.Prefix):], "/")
	keys := strings.Split(key, "/")

	length := len(keys)

	w := bytes.NewBufferString("")
	for index := length - 1; index >= 0; index-- {
		s := keys[index]
		if s == "" {
			continue
		}
		w.WriteRune('.')
		w.WriteString(keys[index])
	}
	return strings.Trim(w.String(), ".")
}
