package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"strconv"
	"strings"

	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coreos/etcd/clientv3"
)

// SkyDNSRecord Skydns 格式存储的 DNS 记录
type SkyDNSRecord struct {
	Name string `json:"name"`
	msg.Service
}

// SkyDNSAgent 返回 SkyDNSAgent 格式的数据
type SkyDNSAgent interface {
	DNSAgent

	// FindSkyDNS 查找 DNS 解析记录
	FindSkyDNS(domain string, prefix bool) []SkyDNSRecord
}

// SkyDNSRecord defines a discoverable service in etcd.
// type SkyDNSRecord msg.Service

// EtcdDNSAgent 基于 Etcd
type EtcdDNSAgent struct {
	Prefix string
	cli    *clientv3.Client
}

// AddRecord 添加 DNS 解析记录
func (a *EtcdDNSAgent) AddRecord(r DNSRecord) error {
	if r.TTL == 0 {
		r.TTL = defaultTTL
	}
	if r.Port == 0 {
		r.Port = defaultPort
	}
	if r.Priority == 0 {
		r.Priority = defaultPriority
	}

	srv := msg.Service{
		Host:     r.Host,
		Port:     r.Port,
		Priority: r.Priority,
		TTL:      r.TTL,
		Text:     net.JoinHostPort(r.Host, strconv.Itoa(r.Port)),
		Weight:   r.Weight,
		// Mail: "",
		Key: a.DomainToKey(r.Name),
	}

	result, err := json.Marshal(srv)
	if err != nil {
		return err
	}

	val := string(result)
	_, err = a.cli.Put(context.Background(), srv.Key, val)

	return err
}

// RemoveRecord 移除 DNS 解析记录
func (a *EtcdDNSAgent) RemoveRecord(r DNSRecord) error {
	key := a.DomainToKey(r.Name)
	_, err := a.cli.Delete(context.Background(), key, clientv3.WithPrefix())
	return err
}

// ModifyRecord 修改 DNS 解析记录
func (a *EtcdDNSAgent) ModifyRecord(o, n DNSRecord) error {

	return nil
}

// FindSkyDNS 查找 DNS 解析记录
func (a *EtcdDNSAgent) FindSkyDNS(domain string, prefix bool) []SkyDNSRecord {
	result := make([]SkyDNSRecord, 0)
	key := a.DomainToKey(domain)
	var resp *clientv3.GetResponse

	if prefix {
		resp, _ = a.cli.Get(context.Background(), key, clientv3.WithPrefix())
	} else {
		resp, _ = a.cli.Get(context.Background(), key)
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
func (a *EtcdDNSAgent) DomainToKey(domain string) string {
	keys := strings.Split(domain, ".")

	length := len(keys)

	w := bytes.NewBufferString(a.Prefix)
	for index := length - 1; index >= 0; index-- {
		s := keys[index]
		if s == "" {
			continue
		}
		w.WriteRune('/')
		w.WriteString(keys[index])
	}
	w.WriteRune('/')
	return w.String()
}

// KeyToDomain Translate Key in Etcd to Domain
func (a *EtcdDNSAgent) KeyToDomain(key string) string {

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
