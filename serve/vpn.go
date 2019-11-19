package serve

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab.jiangxingai.com/applications/base-modules/internal-sdk/go-utils/logger"
	"gitlab.jiangxingai.com/edgenode/dhcp-backend/chaos"
	"gitlab.jiangxingai.com/edgenode/dhcp-backend/dns"
	"gitlab.jiangxingai.com/edgenode/dhcp-backend/vpn"
)

// VirtualNetworkBackend 管理虚拟网络
type VirtualNetworkBackend struct {
	PrivateKey    *rsa.PrivateKey // 私钥
	Timeout       int64           // 请求超时时间
	Prefix        string
	RegisterPath  string // 注册设备用的路径
	Encoding      *base64.Encoding
	WorkerClient  vpn.WorkerClient
	GatewayClient vpn.GatewayClient
	HTTPClient    *http.Client
	AgentPort     int
	Agent         dns.Agent
}

// NewVirtualNetworkBackend 获取新的 VPN 管理后端
func NewVirtualNetworkBackend(encoding string) *VirtualNetworkBackend {
	var enc *base64.Encoding
	if encoding == "" {
		enc = base64.RawStdEncoding
	} else {
		enc = base64.NewEncoding(encoding).WithPadding(base64.NoPadding)
	}
	return &VirtualNetworkBackend{
		Encoding: enc,
	}
}

// Verify 验证请求是否合法
func (b *VirtualNetworkBackend) Verify(r io.Reader) (buf []byte, err error) {
	buf, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}
	if b.Encoding != nil {
		n := b.Encoding.DecodedLen(len(buf))
		dst := make([]byte, n)

		_, err = b.Encoding.Decode(dst, buf)

		if err != nil {
			return
		}
		buf = dst
	}

	return
}

type registerRequest struct {
	WorkerID string `json:"wid"`
	Key      string `json:"key"`
	Version  string `json:"version"` // 请求版本
	Nonce    int64  `json:"nonce"`
}

type vpnIPRequest struct {
	WorkerID string `json:"wid"`
	Type     string `json:"type"`
}

// agentURL 获取 Agent 的地址
func (b *VirtualNetworkBackend) agentURL(gateway string) string {
	return "http://" + net.JoinHostPort(gateway, strconv.Itoa(b.AgentPort)) + "/api/v1/vpn"
}

func (b *VirtualNetworkBackend) selectGateway(vpnType string) []string {

	gwSlice := b.GatewayClient.SelectGateway(vpnType)
	if gwSlice != nil {
		arr := make([]string, 0, len(gwSlice))
		for _, it := range gwSlice {
			arr = append(arr, it.Host)
		}
		return arr
	}
	return nil
}

// updateMasterIP 更新 DNS 记录中 worker 对应 master 的 VPN IP
func (b *VirtualNetworkBackend) updateMasterIP(workerID, masterIP string) {

	err := b.Agent.AddRecord(
		dns.Record{
			Name: workerID + ".master",
			Host: masterIP,
			TTL:  dns.DefaultTTL,
		})
	if err != nil {
		logger.Info(err)
	}

}

// updateWorkerIP 更新 DNS 记录中 worker 的 VPN IP
func (b *VirtualNetworkBackend) updateWorkerIP(workerID, clientIP string) {
	if clientIP != "" {
		workerID = strings.ToLower(workerID)
		err := b.Agent.AddRecord(
			dns.Record{
				Name: workerID + ".worker",
				Host: clientIP,
				TTL:  dns.DefaultTTL,
			})
		if err != nil {
			logger.Info(err)
		}
		err = b.Agent.ModifySubTXTRecord(
			dns.Record{
				Name: workerID + ".iotedge",
				Host: clientIP,
				TTL:  dns.DefaultTTL,
			})
		if err != nil {
			logger.Info(err)
		}
	}
}

func (b *VirtualNetworkBackend) tryConnect(gateway string, reader *bytes.Reader) (resp *http.Response, err error) {
	logger.Info("Try Connect to", gateway)
	req, err := http.NewRequest(http.MethodPost, b.agentURL(gateway), reader)
	if err != nil {
		logger.Info(err)
		return
	}

	resp, err = b.HTTPClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode >= 400 {
		logger.Error("Receive From:", gateway, resp.Status)
		err = errors.New(resp.Status)
		if resp.Body != nil {
			defer resp.Body.Close()
		}
	}
	return
}

// registerWorker 在集群中注册设备
func (b *VirtualNetworkBackend) registerWorker(gateway string, workerID, clientIP string) {
	url := "http://" + gateway + b.RegisterPath
	data := map[string]string{
		"host_ip":   clientIP,
		"worker_id": workerID,
	}
	buf, err := json.Marshal(data)
	if err != nil {
		logger.Info(err, gateway, workerID, clientIP)
		return
	}
	r := bytes.NewReader(buf)
	logger.Info("registerWorker")
	resp, err := b.HTTPClient.Post(url, mimeJSON, r)
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil || resp.StatusCode >= 400 {
		logger.Info(resp, err)
	}
}

func (b *VirtualNetworkBackend) register(workerID, vpnType, remoteAddr string, gateway *string) (buffer []byte, masterIP string, err error) {
	logger.Info("Register ", vpnType)
	var p []byte
	r := vpnIPRequest{
		WorkerID: workerID,
		Type:     vpnType,
	}
	p, err = json.Marshal(r)
	if err != nil {
		logger.Info(err)
		return
	}

	reader := bytes.NewReader(p)

	var ok bool
	var resp *http.Response
	if *gateway != "" {
		resp, err = b.tryConnect(*gateway, reader)
		if err == nil {
			ok = true
		}
	}
	if !ok {
		for _, gw := range b.selectGateway(vpnType) {
			if gw == *gateway {
				continue
			}

			resp, err = b.tryConnect(gw, reader)
			if err != nil {
				reader.Seek(0, io.SeekStart)
				continue
			}
			ok = true
			*gateway = gw
			break
		}
	}

	if !ok {
		err = errNoGatewayAvailable
		return
	}

	logger.Info(resp.Header)

	masterIP = resp.Header.Get("X-Master-IP")
	clientIP := resp.Header.Get("X-Client-IP")
	clientCount := resp.Header.Get("X-Client-Count")
	count, _ := strconv.Atoi(clientCount)

	if clientIP != "" {
		if masterIP != "" {
			status := vpn.WorkerIPStatus{
				Addr:       clientIP,
				Type:       vpnType,
				Gateway:    *gateway,
				Master:     masterIP,
				RegisterIP: remoteAddr,
			}
			err = b.WorkerClient.AddIP(workerID, status)
		}
		// 注册设备
		b.registerWorker(*gateway, workerID, clientIP)
	}

	b.updateMasterIP(workerID, masterIP)
	b.updateWorkerIP(workerID, clientIP)

	if masterIP == "" {
		ip := b.WorkerClient.GetIP(workerID)
		if status, ok := ip.Status[vpnType]; ok {
			masterIP = status.Master
		} else {
			logger.Info("Failed to get master ip", workerID, vpnType)
		}
	}

	if count > 0 {
		gw := vpn.Gateway{
			Host:        *gateway,
			Type:        vpnType,
			ClientCount: count,
		}
		b.GatewayClient.Storage(&gw)
	}

	buffer, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return
}

// verifyKey 验证设备 ID 和 设备的 Key 是否匹配
func (b *VirtualNetworkBackend) verifyKey(req *registerRequest) bool {
	if b.PrivateKey != nil {
		decoded, err := base64.StdEncoding.DecodeString(req.Key)
		if err != nil {
			return false
		}

		buf, err := rsa.DecryptPKCS1v15(rand.Reader, b.PrivateKey, decoded)

		if string(buf) != req.WorkerID {
			return false
		}
	}
	return true
}

func (b *VirtualNetworkBackend) parseRegisterRequest(w http.ResponseWriter, r *http.Request, req *registerRequest) error {
	var buf []byte
	var err error
	buf, err = b.Verify(r.Body)
	defer r.Body.Close()

	err = json.Unmarshal(buf, req)
	if err != nil {
		Error(w, err.Error()+string(buf), http.StatusBadRequest)
		return err
	}
	if !b.verifyKey(req) {
		logger.Warn("Key not match")
		err = errKeyNotMatch
		Error(w, err.Error(), http.StatusUnauthorized)
		return err
	}

	return nil
}

func (b *VirtualNetworkBackend) registerVPN(w http.ResponseWriter, r *http.Request, vpnType string, arrange bool) error {
	var req registerRequest
	err := b.parseRegisterRequest(w, r, &req)
	if err != nil {
		logger.Info("RegisterRequest Failed", err)
		return err
	}

	if b.Timeout > 0 {
		now := time.Now().Unix()
		if (now - req.Nonce) >= b.Timeout {
			ErrorWithCode(w, http.StatusRequestTimeout)
			return err
		}
	}

	var gateway string

	if !arrange {
		// 不用重新分配
		ip := b.WorkerClient.GetIP(req.WorkerID)
		if ip.ID != "" {
			status, ok := ip.Status[vpn.VpnOpenvpn]
			if ok {
				gateway = status.Gateway
			}
		}
	}

	buff, masterIP, err := b.register(req.WorkerID, vpnType, r.RemoteAddr, &gateway)

	if err != nil {
		logger.Info("Register VPN Failed", err)
		Error(w, err.Error(), http.StatusInsufficientStorage)
		return err
	}

	writer := chaos.Writer{
		Writer: w,
	}
	header := w.Header()

	header.Set("Content-Type", "application/octet-stream")

	header.Set("X-Master-IP", masterIP)
	header.Set("X-Ceph-Public-IP", gateway)
	logger.Info("Rersponse Header:", header)

	w.WriteHeader(http.StatusOK)
	writer.WriteChaos(512)
	writer.Write(buff)
	writer.WriteChaos(128)

	return nil
}

// RegisterOpenVPN 注册 OpenVPN
func (b *VirtualNetworkBackend) RegisterOpenVPN(w http.ResponseWriter, r *http.Request) {
	b.registerVPN(w, r, vpn.VpnOpenvpn, false)
}

// RegisterWireguard 注册 Wireguard
func (b *VirtualNetworkBackend) RegisterWireguard(w http.ResponseWriter, r *http.Request) {
	b.registerVPN(w, r, vpn.VpnWireguard, false)
}

// ArrangeOpenVPN 重新分配 OpenVPN IP
func (b *VirtualNetworkBackend) ArrangeOpenVPN(w http.ResponseWriter, r *http.Request) {
	b.registerVPN(w, r, vpn.VpnOpenvpn, true)
}

// ArrangeWireguard 重新分配 Wireguard IP
func (b *VirtualNetworkBackend) ArrangeWireguard(w http.ResponseWriter, r *http.Request) {
	b.registerVPN(w, r, vpn.VpnWireguard, true)
}

// ReceiveVPNAgent 从 VPN Agent 上获取各 VPN 服务器的状态
func (b *VirtualNetworkBackend) ReceiveVPNAgent(w http.ResponseWriter, r *http.Request) {

}
