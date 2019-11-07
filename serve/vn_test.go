package serve

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"gitlab.jiangxingai.com/edgenode/dhcp-backend/etcd"
	"gitlab.jiangxingai.com/edgenode/dhcp-backend/vpn"

	"go.etcd.io/etcd/clientv3"
)

const (
	pemString = `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDLets8+7M+iAQAqN/5BVyCIjhTQ4cmXulL+gm3v0oGMWzLupUS
v8KPA+Tp7dgC/DZPfMLaNH1obBBhJ9DhS6RdS3AS3kzeFrdu8zFHLWF53DUBhS92
5dCAEuJpDnNizdEhxTfoHrhuCmz8l2nt1pe5eUK2XWgd08Uc93h5ij098wIDAQAB
AoGAHLaZeWGLSaen6O/rqxg2laZ+jEFbMO7zvOTruiIkL/uJfrY1kw+8RLIn+1q0
wLcWcuEIHgKKL9IP/aXAtAoYh1FBvRPLkovF1NZB0Je/+CSGka6wvc3TGdvppZJe
rKNcUvuOYLxkmLy4g9zuY5qrxFyhtIn2qZzXEtLaVOHzPQECQQDvN0mSajpU7dTB
w4jwx7IRXGSSx65c+AsHSc1Rj++9qtPC6WsFgAfFN2CEmqhMbEUVGPv/aPjdyWk9
pyLE9xR/AkEA2cGwyIunijE5v2rlZAD7C4vRgdcMyCf3uuPcgzFtsR6ZhyQSgLZ8
YRPuvwm4cdPJMmO3YwBfxT6XGuSc2k8MjQJBAI0+b8prvpV2+DCQa8L/pjxp+VhR
Xrq2GozrHrgR7NRokTB88hwFRJFF6U9iogy9wOx8HA7qxEbwLZuhm/4AhbECQC2a
d8h4Ht09E+f3nhTEc87mODkl7WJZpHL6V2sORfeq/eIkds+H6CJ4hy5w/bSw8tjf
sz9Di8sGIaUbLZI2rd0CQQCzlVwEtRtoNCyMJTTrkgUuNufLP19RZ5FpyXxBO5/u
QastnN77KfUwdj3SJt44U/uh1jAIv4oSLBr8HYUkbnI8
-----END RSA PRIVATE KEY-----`
)

func getEndpoint() string {
	endpoint := os.Getenv("ETCD_ENDPOINT")
	if endpoint == "" {
		endpoint = "10.55.2.114:2379"
	}
	return endpoint
}

type testError interface {
	Error(args ...interface{})
}

func GetEtcdClient(t testError) *clientv3.Client {
	kvc, err := clientv3.New(clientv3.Config{
		Endpoints: []string{
			getEndpoint(),
		}})
	if err != nil {
		t.Error(err)
	}

	return kvc
}

func TestVerify(t *testing.T) {
	block, _ := pem.Decode([]byte(pemString))

	pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Error(err)
	}
	pri.Precompute()
	pub := &pri.PublicKey
	wid := "000-000-000-000"
	enMsg, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(wid))

	enc := base64.NewEncoding("ABCDEFGHIJKLMNOabcdefghijklmnopqrstuvwxyzPQRSTUVWXYZ0123456789-_").WithPadding(base64.NoPadding)
	client := GetEtcdClient(t)
	backend := VirtualNetworkBackend{
		PrivateKey: pri,
		Encoding:   enc,
		GatewayClient: &etcd.VPNGatewayClient{
			Prefix: "/gw",
			Client: client,
		},
		WorkerClient: &etcd.WorkerClient{
			Prefix: "/wk",
			Client: client,
		},
		HTTPClient: http.DefaultClient,
		AgentPort:  9095,
		Agent: &etcd.DNSAgent{
			Prefix: "/skydns",
			Client: client,
		},
	}

	req := registerRequest{
		WorkerID: wid,
		Key:      base64.StdEncoding.EncodeToString(enMsg),
		Nonce:    time.Now().Unix(),
		Version:  "1.0.0",
	}

	buf, err := json.Marshal(req)
	if err != nil {
		t.Error(err)
	}
	n := enc.EncodedLen(len(buf))
	dst := make([]byte, n)
	enc.Encode(dst, buf)

	r, err := http.NewRequest(http.MethodPost, "/api/v1/wg/register", bytes.NewReader(dst))
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	backend.RegisterWireguard(w, r)

	if w.Code != http.StatusOK {
		t.Error(w.Body.String())
		t.Error(w)
	}

}

func TestVerify2(t *testing.T) {

	pri, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Error(err)
	}
	pri.Precompute()
	pub := &pri.PublicKey

	msg := []byte("hello world")
	enMsg, err := rsa.EncryptPKCS1v15(rand.Reader, pub, msg)
	if err != nil {
		t.Error(err)
	}

	b := VirtualNetworkBackend{
		PrivateKey: pri,
	}

	_, err = b.Verify(bytes.NewReader(enMsg[2:]))

	if err != rsa.ErrDecryption {
		t.Error("Failed")

	}

}

func TestWorkerID(t *testing.T) {

	enc := base64.NewEncoding("ABCDEFGHIJKLMNOabcdefghijklmnopqrstuvwxyzPQRSTUVWXYZ0123456789-_").WithPadding(base64.NoPadding)

	req := registerRequest{
		WorkerID: `j01bc6e6ac`,
		Key:      `GSTlIgy3113/X9SYpWTbADvESRGvaDUWJsXR1C5G3V3rjS9f4NfxpAcbKy8l0YcA4pha69Iq5fXzZTq8HYP23d4NFHAHB0+qF8ewXnw4OSLEhsG/BjYhdLs6gRhq1w890tkZ7Lk69f2nqrG1sfAB7CkSl7ztadvJ9NLIg0in7eM=`,
		Nonce:    time.Now().Unix(),
		Version:  "1.0.0",
	}

	buf, err := json.Marshal(req)
	if err != nil {
		t.Error(err)
	}
	n := enc.EncodedLen(len(buf))
	dst := make([]byte, n)
	enc.Encode(dst, buf)

	r, err := http.NewRequest(http.MethodPost, "http://10.55.2.114:1054/api/v1/wg/register", bytes.NewReader(dst))
	if err != nil {
		t.Error(err)
	}

	block, _ := pem.Decode([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQDvmykyYOTmIPvVHCMETMvnfK43IZB75otVEvh2ZZn489s/09z6
obiZow5+lFY3fKka2uT6YM8gZbx0tNUs5v6ekBHQZSDVxWwYr1fqzcUwPLotPUJA
tOfp1zb29LlAUH7h7Bj+wAv4eZPamcSVlLJ+M6Hy8HdIntdVsYEYclxZWQIDAQAB
AoGAHFHYgq3VICYR3dRfhyiUiR3BcZ6z9xD+suV1HHlRw4z/AwJFghIPQYl2MxvR
POmtCxGIMteyY3/i0GB3OcFropOU5iZFWAH0uj3ggTRB4iUpKI+tzVbPq2O9H2/x
VkyiMDZKfPGm0zatV9R7dvQKFb2FpRpUdO2cNcVONuGZr2kCQQD3drO6V+BMGwuI
Po3BhE/IGUCpgmQ7uFdnMgcvNJ67zPIDdEsUhlXfJnl76umbKKNJItzr0up6oCwu
mNJxHmGPAkEA998R/JDU+fyBM7MBzns6g3UtD835AfoGhr2C3zRyDpCYccZfDyPp
rPuKbxva69JfZAGLRybe9yopQ1IoRD9SlwJAX3xmDVkrKzKkWIYKnMk5H7TexomR
s5mF4EPlkcl0FnMWT07oSZssN1bZOX+DdGNR3j6dkEFqSLbVVYWSbiOS8QJAEeWt
assaVaKBwbXfH4WOSAeh5U49+IKRDhGI7Yzf32VZXH2yR2mUacUPzc35FKXv9UyX
Pd/0oWwN5qp79dGMqQJAHb5/etAYraVRaOzYat+Wn0GHgZPO9doXibcl1IzkIluk
7BQbXxg9mjoMO3ZgTMI130fmpds7g/I5gJ5FdxY8qA==
-----END RSA PRIVATE KEY-----`))

	pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Error(err)
	}

	client := GetEtcdClient(t)
	backend := VirtualNetworkBackend{
		PrivateKey: pri,
		Encoding:   enc,
		GatewayClient: &etcd.VPNGatewayClient{
			Prefix: "/gw",
			Client: client,
		},
		WorkerClient: &etcd.WorkerClient{
			Prefix: "/wk",
			Client: client,
		},
		HTTPClient: http.DefaultClient,
		AgentPort:  9095,
		Agent: &etcd.DNSAgent{
			Prefix: "/skydns",
			Client: client,
		},
	}
	w := httptest.NewRecorder()
	backend.RegisterOpenVPN(w, r)

	if w.Code != http.StatusOK {
		t.Error(w)
	}
}

// // TestReceiveVPNAgent 测试接收
// func TestReceiveVPNAgent(t *testing.T) {

// }
func TestAddGateway(t *testing.T) {
	client := GetEtcdClient(t)

	gwClient := &etcd.VPNGatewayClient{
		Prefix: "/test-gw",
		Client: client,
	}

	gw := vpn.Gateway{
		Host:        "88.99.88.1",
		Type:        "unknown",
		ClientCount: 0,
	}
	err := gwClient.Storage(&gw)
	if err != nil {
		t.Error(err)
	}
	gw2 := gwClient.SelectGateway(gw.Type)
	if gw2 == nil || gw2[0].Host != gw.Host {
		t.Error(gw, "!=", gw2)
	}

	_, err = client.Delete(context.Background(), "/test-gw", clientv3.WithPrefix())

	if err != nil {
		t.Error(err)
	}
}
