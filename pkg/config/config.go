package config

import (
	"fmt"

	// "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
)

type AppConfig struct {
	API_PORT  string `mapstructure:"API_PORT"`
	MYSQL_URL string
	MetaNodeVersion string
	DnsLink_                string

	PrivateKey_           string
	ParentAddress         string
	NodeConnectionAddress string
	StorageAddress        string

	NotificationProjectID  string
	NotificationCredential string

	APNSPath              string
	APNSProduction        bool
	APNSMaxConcurrentPush uint64
	APNSKeyID             string
	APNSTeamID            string
	APNSTopic             string

	PrivateKeyPemPath		  string
	NotiFactoryAddress string
	NotiFactoryABIPath string

	NotiHubABIPath string
	NotiHubAddress string

	PathLevelDB string
	
	ServerPrivateKeyPath     string
	ParentConnectionAddress  string
	ParentConnectionType     string
	ConnectionAddress_       string
	ChainId                  uint64
	StorageConnectionAddress string
	RpcURL string

}

var Config *AppConfig

func LoadConfig(path string) (*AppConfig, error) {
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config:   %w", err)
	}

	Config = &config
	return &config, nil
}

// func (c *AppConfig) Version() string {
// 	return c.MetaNodeVersion
// }
// 
// func (c *AppConfig) NodeType() string {
// 	return "explorer"
// }

// func (c *AppConfig) PrivateKey() []byte {
// 	return common.FromHex(c.WalletPrivateKey)
// }

// func (c *AppConfig) PublicConnectionAddress() string {
// 	return c.SocketPublicConnectionAddress
// }

// func (c *AppConfig) ConnectionAddress() string {
// 	return c.SocketConnectionAddress
// }

func (c *AppConfig) DnsLink() string {
	return c.DnsLink_
}
