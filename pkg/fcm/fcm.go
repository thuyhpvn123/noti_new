package fcm

import (
	"context"
	"fmt"
	"time"

	"firebase.google.com/go/v4/messaging"

	firebase "firebase.google.com/go/v4"
	"github.com/meta-node-blockchain/meta-node/pkg/logger"
	"github.com/meta-node-blockchain/noti-contract/internal/model"
	"github.com/meta-node-blockchain/noti-contract/pkg/config"
	"google.golang.org/api/option"
)

var firebaseClient *firebase.App

func NewAndroidNotificationClient(config *config.AppConfig) error {
	opt := option.WithCredentialsFile(config.NotificationCredential)

	// option.
	firebaseConf := &firebase.Config{ProjectID: config.NotificationProjectID}

	fbClient, err := firebase.NewApp(context.Background(), firebaseConf, opt)
	if err != nil {
		logger.Error("Error occured while initial firebase client", err)
		return err
	}

	firebaseClient = fbClient
	return nil
}

func getNotiClient() *firebase.App {
	return firebaseClient
}

func PushAndroidNotification(noti *model.NotiEvent,deviceToken string) {
	fmt.Println("deviceToken la:",deviceToken)
	client := getNotiClient()

	ctx := context.Background()
	oneHour := time.Duration(1) * time.Hour
	message := &messaging.Message{
		Android: &messaging.AndroidConfig{
			TTL:      &oneHour,
			Priority: "normal",
			Notification: &messaging.AndroidNotification{
				Body:  noti.Message,
				Title: noti.Title,
			},
		},
		Token: deviceToken,
		// Data: map[string]string{
		// 	"repo": noti.Dapp,
		// },
	}

	messageClient, err := client.Messaging(ctx)
	if err != nil {
		logger.Error("messageClient", err)
		return
	}

	_, err = messageClient.Send(ctx, message)
	if err != nil {
		logger.Error("messageClient.Send", err)
		return
	}

	logMsg := fmt.Sprintf("Push notification Android successfully: %s", deviceToken)
	logger.Info(logMsg)
}

func PushWebNotification(noti *model.NotiEvent,deviceToken string) {
	ctx := context.Background()
	client := getNotiClient()
	// firebaseWebClient.Messaging()

	message := &messaging.Message{
		// Data: map[string]string{
		// 	"repo": noti.Dapp,
		// },
		Notification: &messaging.Notification{
			Body:  noti.Message,
			Title: noti.Title,
		},
		Token: deviceToken,
	}

	messageClient, err := client.Messaging(ctx)
	if err != nil {
		logger.Error("PushWebNotification::client.Messaging initial client messaging", err)
		return
	}

	_, err = messageClient.Send(ctx, message)
	if err != nil {
		logger.Error("PushWebNotification::messageClient.Send: initial client messaging", err)
		return
	}
	logger.Info("Push web notification successfully")
}
