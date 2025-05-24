package apns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/meta-node-blockchain/meta-node/pkg/logger"
	"github.com/meta-node-blockchain/noti-contract/pkg/config"
	"github.com/mitchellh/mapstructure"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

var (
	ApnsClient             *apns2.Client
	MaxConcurrentIOSPushes chan struct{}
	doOnce                 sync.Once
)

func NewIosNotificationClient(config *config.AppConfig) error {
	if config.APNSPath == "" {
		logger.Error("Invalid APNS path")
		return errors.New("invalid APNS path")
	}
	if config.APNSKeyID == "" || config.APNSTeamID == "" {
		msg := "you should provide KeyID and TeamID for p8 token"
		logger.Error(msg)
		return errors.New(msg)
	}
	authKey, err := token.AuthKeyFromFile(config.APNSPath)
	if err != nil {
		logger.Error("Cert Error:", err)
		return errors.New(err.Error())
	}

	token := &token.Token{
		AuthKey: authKey,
		KeyID:   config.APNSKeyID,
		TeamID:  config.APNSTeamID,
	}

	ApnsClient = apns2.NewTokenClient(token)

	if config.APNSProduction {
		ApnsClient = ApnsClient.Production()
	} else {
		ApnsClient = ApnsClient.Development()
	}

	doOnce.Do(func() {
		MaxConcurrentIOSPushes = make(chan struct{}, config.APNSMaxConcurrentPush)
	})

	return nil
}

func getIosNotificationClient() *apns2.Client {
	return ApnsClient
}

type Sound struct {
	Critical int     `json:"critical,omitempty"`
	Name     string  `json:"name,omitempty"`
	Volume   float32 `json:"volume,omitempty"`
}

func iosAlertDictionary(notificationPayload *payload.Payload, req *PushNotification) *payload.Payload {
	// Alert dictionary

	if len(req.Title) > 0 {
		notificationPayload.AlertTitle(req.Title)
	}

	if len(req.InterruptionLevel) > 0 {
		notificationPayload.InterruptionLevel(payload.EInterruptionLevel(req.InterruptionLevel))
	}

	if len(req.Message) > 0 && len(req.Title) > 0 {
		notificationPayload.AlertBody(req.Message)
	}

	if len(req.Alert.Title) > 0 {
		notificationPayload.AlertTitle(req.Alert.Title)
	}

	// Apple Watch & Safari display this string as part of the notification interface.
	if len(req.Alert.Subtitle) > 0 {
		notificationPayload.AlertSubtitle(req.Alert.Subtitle)
	}

	if len(req.Alert.TitleLocKey) > 0 {
		notificationPayload.AlertTitleLocKey(req.Alert.TitleLocKey)
	}

	if len(req.Alert.LocArgs) > 0 {
		notificationPayload.AlertLocArgs(req.Alert.LocArgs)
	}

	if len(req.Alert.TitleLocArgs) > 0 {
		notificationPayload.AlertTitleLocArgs(req.Alert.TitleLocArgs)
	}

	if len(req.Alert.Body) > 0 {
		notificationPayload.AlertBody(req.Alert.Body)
	}

	if len(req.Alert.LaunchImage) > 0 {
		notificationPayload.AlertLaunchImage(req.Alert.LaunchImage)
	}

	if len(req.Alert.LocKey) > 0 {
		notificationPayload.AlertLocKey(req.Alert.LocKey)
	}

	if len(req.Alert.Action) > 0 {
		notificationPayload.AlertAction(req.Alert.Action)
	}

	if len(req.Alert.ActionLocKey) > 0 {
		notificationPayload.AlertActionLocKey(req.Alert.ActionLocKey)
	}

	// General
	if len(req.Category) > 0 {
		notificationPayload.Category(req.Category)
	}

	if len(req.Alert.SummaryArg) > 0 {
		notificationPayload.AlertSummaryArg(req.Alert.SummaryArg)
	}

	if req.Alert.SummaryArgCount > 0 {
		notificationPayload.AlertSummaryArgCount(req.Alert.SummaryArgCount)
	}

	return notificationPayload
}

// ref: https://github.com/appleboy/gorush/blob/master/notify/notification_apns.go
func getIOSNotification(req *PushNotification) *apns2.Notification {
	notification := &apns2.Notification{
		ApnsID:     req.ApnsID,
		Topic:      req.Topic,
		CollapseID: req.CollapseID,
	}

	if req.Expiration != nil {
		notification.Expiration = time.Unix(*req.Expiration, 0)
	}

	if len(req.PushType) > 0 {
		notification.PushType = apns2.EPushType(req.PushType)
	}

	payload := payload.NewPayload()

	// add alert object if message length > 0 and title is empty
	if len(req.Message) > 0 && req.Title == "" {
		payload.Alert(req.Message)
	}

	// zero value for clear the badge on the app icon.
	if req.Badge != nil && *req.Badge >= 0 {
		payload.Badge(*req.Badge)
	}

	if req.MutableContent {
		payload.MutableContent()
	}

	switch req.Sound.(type) {
	// from http request binding
	case map[string]interface{}:
		result := &Sound{}
		_ = mapstructure.Decode(req.Sound, &result)
		payload.Sound(result)
	// from http request binding for non critical alerts
	case string:
		payload.Sound(&req.Sound)
	case Sound:
		payload.Sound(&req.Sound)
	}

	if len(req.SoundName) > 0 {
		payload.SoundName(req.SoundName)
	}

	if req.SoundVolume > 0 {
		payload.SoundVolume(req.SoundVolume)
	}

	if req.ContentAvailable {
		payload.ContentAvailable()
	}

	if len(req.URLArgs) > 0 {
		payload.URLArgs(req.URLArgs)
	}

	if len(req.ThreadID) > 0 {
		payload.ThreadID(req.ThreadID)
	}

	for k, v := range req.Data {
		payload.Custom(k, v)
	}

	payload = iosAlertDictionary(payload, req)

	notification.Payload = payload

	return notification
}

func PushIosNotification(ctx context.Context, config *config.AppConfig, req *PushNotification) {
	client := getIosNotificationClient()
	noti := getIOSNotification(req)
	for _, token := range req.Tokens {
		go func(notification apns2.Notification, token string) {
			notification.DeviceToken = token
			res, err := client.Push(&notification)
			if err != nil || (res != nil && res.StatusCode != http.StatusOK) {
				if err == nil {
					// error message:
					// ref: https://github.com/sideshow/apns2/blob/master/response.go#L14-L65
					err = errors.New(res.Reason)
				}
				logger.Error(err)
			}
			if res != nil && res.Sent() {
				logger.Info(fmt.Sprintf("Success push Ios notification to device: %s", token))
			}
		}(*noti, token)
		logMsg := fmt.Sprintf("Push notification IOS successfully to device: %s", token)
		logger.Info(logMsg)
	}

}
