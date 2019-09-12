package etcd

import (
	"context"
	"dhcp-backend/vpn"
	"encoding/json"

	"github.com/coreos/etcd/clientv3"
)

// WorkerClient 基于 Etcd 的 WorkerClient 实现
type WorkerClient struct {
	Prefix string
	Client *clientv3.Client
}

func (w *WorkerClient) idToKey(workerID string) string {
	return w.Prefix + "/" + workerID
}

// AddIP 添加设备的 IP
func (w *WorkerClient) AddIP(workerID string, ip vpn.WorkerIPStatus) error {
	key := w.idToKey(workerID)
	resp, err := w.Client.Get(context.Background(), key)
	if err != nil {
		return err
	}
	var worker vpn.WorkerIP
	if resp.Count > 0 {
		err = json.Unmarshal(resp.Kvs[0].Value, &worker)
		if err != nil {
			return err
		}
	}
	if worker.Status == nil {
		worker.Status = make(map[string]*vpn.WorkerIPStatus)
	}
	worker.ID = workerID
	worker.Status[ip.Type] = &ip

	b, err := json.Marshal(worker)
	if err != nil {
		return err
	}
	_, err = w.Client.Put(context.Background(), key, string(b))

	return err
}

// GetIP 获取设备的 IP
func (w *WorkerClient) GetIP(workerID string) vpn.WorkerIP {
	var ip vpn.WorkerIP
	key := w.idToKey(workerID)
	resp, err := w.Client.Get(context.Background(), key)
	if err != nil || resp.Count == 0 {
		return ip
	}
	json.Unmarshal(resp.Kvs[0].Value, &resp)
	return ip
}

// RemoveIP 移除设备的 IP
func (w *WorkerClient) RemoveIP(workerID string, ip vpn.WorkerIPStatus) error {
	key := w.idToKey(workerID)
	resp, err := w.Client.Get(context.Background(), key)
	if err != nil {
		return err
	}
	var worker vpn.WorkerIP
	if resp.Count == 0 {
		return nil
	}
	err = json.Unmarshal(resp.Kvs[0].Value, &worker)
	if err != nil {
		return err
	}
	if worker.Status == nil {
		return nil
	}

	delete(worker.Status, ip.Type)

	b, err := json.Marshal(worker)
	if err != nil {
		return err
	}
	_, err = w.Client.Put(context.Background(), key, string(b))

	return err
}
