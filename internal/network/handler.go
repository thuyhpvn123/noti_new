package network

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	e_common "github.com/ethereum/go-ethereum/common"
	"github.com/meta-node-blockchain/meta-node/cmd/client"
	"github.com/meta-node-blockchain/meta-node/pkg/logger"
	"github.com/meta-node-blockchain/noti-contract/pkg/dbsvc"
	"github.com/syndtr/goleveldb/leveldb"

	// "github.com/meta-node-blockchain/meta-node/types"
	"github.com/meta-node-blockchain/noti-contract/internal/model"
	"github.com/meta-node-blockchain/noti-contract/internal/usecase"
	"github.com/meta-node-blockchain/noti-contract/internal/utils"
	"github.com/meta-node-blockchain/noti-contract/pkg/apns"
	"github.com/meta-node-blockchain/noti-contract/pkg/config"
	"github.com/meta-node-blockchain/noti-contract/pkg/fcm"
)

type NotiHandler struct {
	config             *config.AppConfig
	chainClient        *client.Client
	notiFactoryAddress e_common.Address
	notiFactoryABI     *abi.ABI
	notiHubAddress     e_common.Address
	notiHubABI         *abi.ABI
	deviceTokenUsecase usecase.DeviceTokenUsecase
	serverPrivateKey   string
	eventChan          chan model.EventLog
	DB                 *leveldb.DB
}

func NewNotiEventHandler(
	config *config.AppConfig,
	chainClient *client.Client,
	notiFactoryAddress e_common.Address,
	notiFactoryABI *abi.ABI,
	notiHubAddress e_common.Address,
	notiHubABI *abi.ABI,
	deviceTokenUsecase usecase.DeviceTokenUsecase,
	serverPrivateKey string,
	eventChan chan model.EventLog,
	DB *leveldb.DB,

) *NotiHandler {
	return &NotiHandler{
		config:             config,
		chainClient:        chainClient,
		notiFactoryAddress: notiFactoryAddress,
		notiFactoryABI:     notiFactoryABI,
		notiHubAddress:     notiHubAddress,
		notiHubABI:         notiHubABI,
		deviceTokenUsecase: deviceTokenUsecase,
		serverPrivateKey:   serverPrivateKey,
		eventChan:          eventChan,
		DB:                 DB,
	}
}
func (h *NotiHandler) ListenEvents() {
	go func() {
		logger.Info("⏳ Start listening for new events...")

		rpcURL := h.config.RpcURL

		// Đọc ABI
		abiBytes, err := os.ReadFile(h.config.NotiFactoryABIPath)
		if err != nil {
			logger.Error("Error reading ABI file:", err)
			return
		}
		abiNotiFactoryJSON := string(abiBytes)
		abiBytes1, err := os.ReadFile(h.config.NotiHubABIPath)
		if err != nil {
			logger.Error("Error reading ABI file:", err)
			return
		}
		abiNotiHubJSON := string(abiBytes1)
		// Lấy topic0 cho UserSubscribed và GlobalNotification
		userSubscribedTopic, err := utils.GetTopic0FromABI(abiNotiFactoryJSON, "UserSubscribed")
		if err != nil {
			logger.Error("Error getting UserSubscribed topic0:", err)
			return
		}
		globalNotificationTopic, err := utils.GetTopic0FromABI(abiNotiHubJSON, "GlobalNotification")
		if err != nil {
			logger.Error("Error getting GlobalNotification topic0:", err)
			return
		}

		// Lấy last block từ DB (nếu có)
		var fromBlock uint64 = 0
		callmap := map[string]interface{}{
			"key": "lastBlock",
		}
		lastBlockBytes, err := dbsvc.ReadValueStorage(callmap, h.DB)

		if err == nil {
			fromBlock, _ = strconv.ParseUint(string(lastBlockBytes), 0, 64)
		}
		for {
			select {
			default:
				// Lấy latest block
				latestBlock, err := utils.GetLatestBlockNumber(rpcURL)
				if err != nil {
					logger.Error("Failed to get latest block:", err)
					time.Sleep(2 * time.Second)
					continue
				}
				latestBlockUint, _ := strconv.ParseUint(latestBlock, 0, 64)
				if latestBlockUint <= fromBlock {
					time.Sleep(1 * time.Second)
					continue
				}
				// Lấy log cho từng contract và topic riêng biệt
				eventSources := []struct {
					Address string
					Topic   string
				}{
					{Address: h.config.NotiFactoryAddress, Topic: userSubscribedTopic},
					{Address: h.config.NotiHubAddress, Topic: globalNotificationTopic},
				}

				for _, source := range eventSources {
					logs, err := utils.GetLogs(
						rpcURL,
						fmt.Sprintf("0x%x", fromBlock+1),
						fmt.Sprintf("0x%x", latestBlockUint),
						source.Address,
						source.Topic,
					)
					if err != nil {
						logger.Error("Error fetching logs for topic", source.Topic, ":", err)
						time.Sleep(1 * time.Second)
						continue
					}
					for _, raw := range logs {

						var log model.EventLog
						if err := json.Unmarshal(raw, &log); err != nil {
							logger.Warn("Cannot decode event log:", err)
							continue
						}
						h.eventChan <- log
					}
				}
				// Lưu latestBlock vào DB
				callmap := map[string]interface{}{
					"key":  "lastBlock",
					"data": (strconv.FormatUint(latestBlockUint, 10)),
				}
				err = dbsvc.WriteValueStorage(callmap, h.DB)

				if err != nil {
					logger.Error("Failed to save lastBlock to DB:", err)
				}
				fromBlock = latestBlockUint

				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (h *NotiHandler) HandleConnectSmartContract(event model.EventLog) {
	switch event.Topics[0] {
	case h.notiFactoryABI.Events["UserSubscribed"].ID.String():
		h.handleUserSubscribed(event)
	case h.notiHubABI.Events["GlobalNotification"].ID.String():
		h.handlePublishNotification(event)
	}
}
func (h *NotiHandler) handleUserSubscribed(event model.EventLog) {
	fmt.Println("UserSubscribed")
	if len(event.Topics) < 3 {
		logger.Warn("Not enough topics in UserRef event")
		return
	}
	// Lấy user và dapp từ indexed topics
	user := common.HexToAddress(event.Topics[1]).Hex()
	dapp := common.HexToAddress(event.Topics[2]).Hex()

	result := make(map[string]interface{})
	err := h.notiFactoryABI.UnpackIntoMap(result, "UserSubscribed", e_common.FromHex(event.Data))
	if err != nil {
		logger.Error("can't unpack to map", err)
		return
	}
	encryptedDeviceToken, ok := result["encryptedDeviceToken"].([]byte)
	if !ok {
		logger.Error("fail in parse encryptedDeviceToken :", err)
		return
	}
	fmt.Println("encryptedDeviceToken:", hex.EncodeToString(encryptedDeviceToken[16:]))
	iv := encryptedDeviceToken[:16]
	fmt.Println("iv la:", hex.EncodeToString(iv))
	deviceToken := model.DeviceToken{
		DAppAddress:    dapp,
		UserAddress:    user,
		PublicKey:      hex.EncodeToString(result["publicKey"].([]uint8)),
		EncryptedToken: hex.EncodeToString(encryptedDeviceToken[16:]),
		Platform:       result["platform"].(uint8),
		CreatedAt:      uint64(time.Now().Unix()),
	}
	histories, err := h.deviceTokenUsecase.GetEncryptedTokensByDappAndUser(dapp, user)
	if err != nil {
		logger.Error("fail in get Encrypted Token by dapp and user:", err)
		return
	}
	if len(histories) == 0 {
		err = h.deviceTokenUsecase.Insert(deviceToken)
		if err != nil {
			msg := fmt.Sprintf("Unable to store device token from user %s, %v", user, err)
			logger.Error(msg)
		}
	} else {
		var count uint
		for _, v := range histories {
			if v.Platform == result["platform"].(uint8) {
				count++
				if v.EncryptedToken != hex.EncodeToString(encryptedDeviceToken[16:]) {
					deviceToken.ID = v.ID
					err = h.deviceTokenUsecase.Update(deviceToken)
					if err != nil {
						msg := fmt.Sprintf("Unable to update device token from user %s, %v", user, err)
						logger.Error(msg)
					}

				} else {
					logger.Error("same token device existed")
				}
				break
			}
		}
		if count == 0 {
			err = h.deviceTokenUsecase.Insert(deviceToken)
			if err != nil {
				msg := fmt.Sprintf("Unable to insert device token from user %s, %v", user, err)
				logger.Error(msg)
			}
		}
	}

}
func (h *NotiHandler) handlePublishNotification(event model.EventLog) {
	fmt.Println("handlePublishNotification")
	var result model.NotiEvent
	var histories []*model.DeviceToken
	err := h.notiHubABI.UnpackIntoInterface(&result, "GlobalNotification", e_common.FromHex(event.Data))
	if err != nil {
		logger.Error("can't unpack to interface handlePublishNotification", err)
		return
	}
	result.DappOwner = common.HexToAddress(event.Topics[1])
	fmt.Println("result.DappOwner:",result.DappOwner)
	result.User = common.HexToAddress(event.Topics[2])
	fmt.Println("result.User:",result.User)
	histories, err = h.deviceTokenUsecase.GetEncryptedTokensByDappAndUser(result.DappOwner.Hex(), result.User.Hex())
	if err != nil {
		logger.Error("fail in get Encrypted Token by dapp and user:", err)
		return
	}
	for _, v := range histories {
		if v.EncryptedToken != "" {
			iv := make([]byte, 16)
			encyptedToken,err := hex.DecodeString(v.EncryptedToken)
			serverPrivateKey,err := hex.DecodeString(h.serverPrivateKey)
			publicKey ,err := hex.DecodeString(v.PublicKey)
			deviceToken, err := utils.DecryptAESCBC(encyptedToken, serverPrivateKey, publicKey, iv)

			if err != nil {
				logger.Error("fail in decrypt token:", err)
				return
			}
			deviceTokenKq := strings.Trim(string(deviceToken), "\"")

			h.handlePublish(result, deviceTokenKq, v.Platform)
		} else {
			logger.Error(fmt.Printf("can not find encryptedToken in db with user %v \n", result.User.Hex()))
		}
	}
}
func (h *NotiHandler) handlePublish(result model.NotiEvent, deviceToken string, platform uint8) {
	switch platform {
	case uint8(0): //ANDROID
		fcm.PushAndroidNotification(&result, deviceToken)
	case uint8(1): //IOS
		apns.PushIosNotification(context.Background(), h.config, &apns.PushNotification{
			// ID:      result.EventId.String(),
			Topic:   h.config.APNSTopic,
			Tokens:  []string{deviceToken},
			Message: result.Message,
			Title:   result.Title,
		})
	case uint8(2): //WEB
		fcm.PushWebNotification(&result, deviceToken)
	default:
		logger.Error("Invalid device type")
	}
}
