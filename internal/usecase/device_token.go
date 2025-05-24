package usecase

import (
	"github.com/meta-node-blockchain/noti-contract/internal/model"
	"github.com/meta-node-blockchain/noti-contract/internal/repository"
)

type DeviceTokenUsecase interface {
	Save(deviceToken model.DeviceToken) error
	Insert(deviceToken model.DeviceToken) error
	Update(deviceToken model.DeviceToken) error
	GetEncryptedTokensByDappAndUser(dapp, user string) ([]*model.DeviceToken, error)
	GetEncryptedTokensByUser(user string) ([]*model.DeviceToken, error)
}

type deviceTokenUsecase struct {
	repo repository.DeviceTokenRepository
}

func NewDeviceTokenUsecase(
	repo repository.DeviceTokenRepository,
) DeviceTokenUsecase {
	return &deviceTokenUsecase{repo}
}

func (usecase *deviceTokenUsecase) Save(deviceToken model.DeviceToken) error {
	return usecase.repo.Save(deviceToken)
}
func (usecase *deviceTokenUsecase) Insert(deviceToken model.DeviceToken) error {
	return usecase.repo.Insert(deviceToken)
}
func (usecase *deviceTokenUsecase) Update(deviceToken model.DeviceToken) error {
	return usecase.repo.Update(deviceToken)
}

func (usecase *deviceTokenUsecase) GetEncryptedTokensByDappAndUser(dapp, user string) ([]*model.DeviceToken, error) {
	return usecase.repo.GetEncryptedTokensByDappAndUser(dapp, user)
}
func (usecase *deviceTokenUsecase) GetEncryptedTokensByUser(user string) ([]*model.DeviceToken, error) {
	return usecase.repo.GetEncryptedTokensByUser(user)
}
