package model

type DeviceToken struct {
	ID             int    `gorm:"column:id;primaryKey;autoIncrement"` // Primary Key
	DAppAddress    string `gorm:"column:dapp;size:42;not null"`
	UserAddress    string `gorm:"column:user;size:42;not null"`
	PublicKey	   string `gorm:"column:public_key;type:text;not null"`
	EncryptedToken string `gorm:"column:encrypted_token;type:text;not null"` // Encrypted Token
	Platform       uint8  `gorm:"column:platform;not null"`                  // Platform field included
	CreatedAt      uint64 `gorm:"column:created_at"`                         // Created Timestamp
}
// Define a composite unique constraint
func (DeviceToken) TableName() string {
	return "device_tokens"
}
