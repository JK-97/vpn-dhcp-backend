package option

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/BurntSushi/toml"
)

// DNSConfig DNS 相关的配置
type DNSConfig struct {
	DNSPrefix string
	Endpoints []string
	Username  string
	Password  string
}

// DHCPConfig DHCP 服务
type DHCPConfig struct {
	DNSConfig
	WorkerPrefix  string
	GatewayPrefix string
	RegisterPath  string
}

// ApplicationConfig 应用程序配置
type ApplicationConfig struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	MongoURI   string
	DHCPConfig
}

type appConfig struct {
	PrivateKey string
	PublicKey  string
	MongoURI   string
	DHCPConfig
}

// ReadConfigFile 从文件中读取配置
func ReadConfigFile(filename string) *ApplicationConfig {
	var config appConfig
	_, err := toml.DecodeFile(filename, &config)
	if err != nil {
		return nil
	}
	result := new(ApplicationConfig)

	if config.PrivateKey != "" {
		block, _ := pem.Decode([]byte(config.PrivateKey))
		result.PrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	if config.PublicKey != "" {
		block, _ := pem.Decode([]byte(config.PublicKey))
		result.PublicKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
	}

	result.MongoURI = config.MongoURI
	result.DHCPConfig = config.DHCPConfig

	return result
}
