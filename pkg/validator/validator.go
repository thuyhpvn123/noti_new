package validator

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	v "github.com/go-playground/validator/v10"
)

var validate *v.Validate

func ethAddressValidator(fl v.FieldLevel) bool {
	address := fl.Field().String()
	return common.IsHexAddress(address)
}

// Validate ProductID: bytes32 in Solidity
func productIdValidator(fl v.FieldLevel) bool {
	productId := fl.Field().String()

	decoded, err := hex.DecodeString(productId)
	if err != nil {
		return false
	}

	return len(decoded) == 32
}

// InitValidator initializes the validator
func InitValidator() {
	validate = v.New()
}

func ValidateStruct(data interface{}) error {
	if validate == nil {
		InitValidator() // Initialize the validator if not already initialized
	}

	err := validate.Struct(data)
	if err != nil {
		// If there are validation errors, you can process them as needed
		// For example, you can extract and return specific error messages
		// based on field names using the Field() function.
		// Example: err.(v.ValidationErrors).Field("Email").Tag()
		// You can also customize the error messages by setting custom tags on your struct fields.
		return err
	}

	return nil
}
