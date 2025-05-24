package repository

import (
	"github.com/meta-node-blockchain/noti-contract/internal/model"
	"gorm.io/gorm"
)

type DeviceTokenRepository interface {
	Save(deviceToken model.DeviceToken) error
	Insert(deviceToken model.DeviceToken) error
	Update(deviceToken model.DeviceToken) error
	GetEncryptedTokensByDappAndUser(dapp, user string) ([]*model.DeviceToken, error)
	GetEncryptedTokensByUser(user string) ([]*model.DeviceToken, error)
}

type deviceTokenRepository struct {
	db *gorm.DB
}

func NewDeviceTokenRepository(db *gorm.DB) DeviceTokenRepository {
	return &deviceTokenRepository{db}
}

func (repo *deviceTokenRepository) Save(deviceToken model.DeviceToken) error {
	if err := repo.db.Save(&deviceToken).Error; err != nil {
		return err
	}
	return nil
}
func (repo *deviceTokenRepository) Insert(deviceToken model.DeviceToken) error {
	if err := repo.db.Create(&deviceToken).Error; err != nil {
		return err
	}
	return nil
}
func (repo *deviceTokenRepository) Update(deviceToken model.DeviceToken) error {
	if err := repo.db.Model(&model.DeviceToken{}).
		Where("id = ?", deviceToken.ID).
		Updates(deviceToken).Error; err != nil {
		return err
	}
	return nil
}

func (repo *deviceTokenRepository) GetEncryptedTokensByDappAndUser(dapp, user string) ([]*model.DeviceToken, error) {
	var histories []*model.DeviceToken
	result := repo.db.
	Where("`dapp` = ? AND `user` = ?", dapp, user).
	Find(&histories)
	if result.Error == gorm.ErrRecordNotFound {
		return histories, result.Error
	}
	return histories, nil
}
func (repo *deviceTokenRepository) GetEncryptedTokensByUser(user string) ([]*model.DeviceToken, error) {
	var histories []*model.DeviceToken
	result := repo.db.
	Where("`user` = ?", user).
	Find(&histories)
	if result.Error == gorm.ErrRecordNotFound {
		return histories, result.Error
	}
	return histories, nil
}
