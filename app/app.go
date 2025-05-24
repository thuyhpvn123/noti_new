package app

import (
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/meta-node-blockchain/meta-node/cmd/client"
	"github.com/meta-node-blockchain/meta-node/pkg/logger"
	// "github.com/meta-node-blockchain/meta-node/types"
	"github.com/meta-node-blockchain/noti-contract/internal/network"
	"github.com/meta-node-blockchain/noti-contract/internal/repository"
	"github.com/meta-node-blockchain/noti-contract/internal/usecase"
	"github.com/meta-node-blockchain/noti-contract/pkg/apns"
	"github.com/meta-node-blockchain/noti-contract/pkg/config"
	"github.com/meta-node-blockchain/noti-contract/pkg/dbsvc"
	"github.com/meta-node-blockchain/noti-contract/pkg/fcm"
	"github.com/meta-node-blockchain/noti-contract/internal/model"
	c_config "github.com/meta-node-blockchain/meta-node/cmd/client/pkg/config"
)

type App struct {
	Config *config.AppConfig
	ApiApp *gin.Engine

	ChainClient *client.Client
	EventChan   chan model.EventLog
	StopChan    chan bool

	NotiHandler *network.NotiHandler
}

func NewApp(
	configPath string,
	loglevel int,
) (*App, error) {
	loggerConfig := &logger.LoggerConfig{
		Flag:    loglevel,
		Outputs: []*os.File{os.Stdout},
	}
	logger.SetConfig(loggerConfig)

	config, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal("invalid configuration", err)
		return nil, err
	}
	app := &App{}

	dbsvc.StartMySQL(config)
	db := dbsvc.GetMySqlConn()
	leveldb, err :=dbsvc.Open(config.PathLevelDB)

	deviceTokenRepo := repository.NewDeviceTokenRepository(db)
	deviceTokenUsecase := usecase.NewDeviceTokenUsecase(deviceTokenRepo)
	app.ChainClient, err = client.NewClient(
		&c_config.ClientConfig{
			Version_:                config.MetaNodeVersion,
			PrivateKey_:             config.PrivateKey_,
			ParentAddress:           config.ParentAddress,
			ParentConnectionAddress: config.ParentConnectionAddress,
			DnsLink_:                config.DnsLink(),
			ConnectionAddress_:      config.ConnectionAddress_,
			ParentConnectionType:    config.ParentConnectionType,
			ChainId:                 config.ChainId,
		},
	)
	if err != nil {
		logger.Error(fmt.Sprintf("error when create chain client %v", err))
		return nil, err
	}

	app.EventChan = make(chan model.EventLog, 1000) // buffer 100 để tránh nghẽn

	readerFactory, err := os.Open(config.NotiFactoryABIPath)
	if err != nil {
		logger.Error("Error occured while read create card smart contract abi")
		return nil, err
	}
	defer readerFactory.Close()

	notiFactoryAbi, err := abi.JSON(readerFactory)
	if err != nil {
		logger.Error("Error occured while parse create noti factory smart contract abi")
		return nil, err
	}
	readerHub, err := os.Open(config.NotiHubABIPath)
	if err != nil {
		logger.Error("Error occured while read create card smart contract abi")
		return nil, err
	}
	defer readerHub.Close()

	notiHubAbi, err := abi.JSON(readerHub)
	if err != nil {
		logger.Error("Error occured while parse create noti factory smart contract abi")
		return nil, err
	}

	err = fcm.NewAndroidNotificationClient(config)
	if err != nil {
		logger.Error("Invalid configuration for android notification client")
		return nil, err
	}
	err = apns.NewIosNotificationClient(config)
	if err != nil {
		logger.Error("Invalid configuration for ios notification client")
		return nil, err
	}
	bserverPrivateKey, err := os.ReadFile(config.ServerPrivateKeyPath)
	if err != nil {
		logger.Error("Can not read private key pem file")
		return nil, err
	}
	app.NotiHandler = network.NewNotiEventHandler(
		config,
		app.ChainClient,
		common.HexToAddress(config.NotiFactoryAddress),
		&notiFactoryAbi,
		common.HexToAddress(config.NotiHubAddress),
		&notiHubAbi,
		deviceTokenUsecase,
		string(bserverPrivateKey),
		app.EventChan,
		leveldb,
	)

	app.Config = config

	return app, nil
}

func (app *App) Run() {
	app.StopChan = make(chan bool)
	go app.NotiHandler.ListenEvents() // BẮT ĐẦU LẮNG NGHE EVENT

	for {
		select {
		case <-app.StopChan:
			return
		case eventLogs := <-app.EventChan:
			logger.Info("eventLogs:",eventLogs)
			app.NotiHandler.HandleConnectSmartContract(eventLogs)
		}
	}
}

func (app *App) Stop() error {
	app.ChainClient.Close()

	logger.Warn("App Stopped")
	return nil
}

