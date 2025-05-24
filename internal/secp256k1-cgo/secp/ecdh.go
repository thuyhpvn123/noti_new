package secp

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lsecp256k1
#include "secp256k1.h"
#include "secp256k1_ecdh.h"
#include "secp256k1_recovery.h"


my_secp256k1_context* get_context() {
	static my_secp256k1_context* ctx = NULL;
	if (ctx == NULL) {
		ctx = my_secp256k1_context_create(SECP256K1_CONTEXT_SIGN | SECP256K1_CONTEXT_VERIFY);
	}
	return ctx;
}
*/
import "C"
import (
	"encoding/hex"
	"errors"
	"unsafe"
)

func CreateECDH(privHex, pubHex string) (string, error) {
	privBytes, err := hex.DecodeString(privHex)
	if err != nil || len(privBytes) != 32 {
		return "", errors.New("private key không hợp lệ")
	}

	pubBytes, err := hex.DecodeString(pubHex)
	if err != nil || (len(pubBytes) != 33 && len(pubBytes) != 65) {
		return "", errors.New("public key không hợp lệ")
	}

	ctx := C.get_context()

	var pubkey C.my_secp256k1_pubkey
	if C.my_secp256k1_ec_pubkey_parse(ctx, &pubkey, (*C.uchar)(unsafe.Pointer(&pubBytes[0])), C.size_t(len(pubBytes))) != 1 {
		return "", errors.New("không parse được public key")
	}

	var output [32]byte
	if C.my_secp256k1_ecdh(ctx, (*C.uchar)(unsafe.Pointer(&output[0])), &pubkey, (*C.uchar)(unsafe.Pointer(&privBytes[0])), nil, nil) != 1 {
		return "", errors.New("ECDH thất bại")
	}

	return hex.EncodeToString(output[:]), nil
}

func CreatePublicKey(privHex string, compressed bool) (string, error) {
	privBytes, err := hex.DecodeString(privHex)
	if err != nil || len(privBytes) != 32 {
		return "", errors.New("private key không hợp lệ")
	}

	ctx := C.get_context()

	// Tạo public key
	var pubkey C.my_secp256k1_pubkey
	if C.my_secp256k1_ec_pubkey_create(ctx, &pubkey, (*C.uchar)(unsafe.Pointer(&privBytes[0]))) != 1 {
		return "", errors.New("không tạo được public key")
	}

	// Serialize public key
	var outputLen C.size_t
	var flags C.uint
	if compressed {
		outputLen = 33
		flags = C.SECP256K1_EC_COMPRESSED
	} else {
		outputLen = 65
		flags = C.SECP256K1_EC_UNCOMPRESSED
	}

	output := make([]byte, outputLen)
	if C.my_secp256k1_ec_pubkey_serialize(ctx,
		(*C.uchar)(unsafe.Pointer(&output[0])),
		&outputLen,
		&pubkey,
		flags) != 1 {
		return "", errors.New("không serialize được public key")
	}

	return hex.EncodeToString(output), nil
}

func RecoverPublicKey(hashHex, sigHex string) (string, error) {
	hashBytes, err := hex.DecodeString(hashHex)
	if err != nil || len(hashBytes) != 32 {
		return "", errors.New("hash không hợp lệ")
	}

	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil || len(sigBytes) != 65 {
		return "", errors.New("signature không hợp lệ")
	}

	ctx := C.get_context()

	// Tách r||s (64 bytes) và v (1 byte)
	rs := sigBytes[:64]
	v := int(sigBytes[64]) // không cần -27 nếu v đã đúng định dạng 0-3

	if v < 0 || v > 3 {
		return "", errors.New("v không hợp lệ, cần trong khoảng [0..3]")
	}

	// Parse recoverable signature
	var recSig C.my_secp256k1_ecdsa_recoverable_signature
	if C.my_secp256k1_ecdsa_recoverable_signature_parse_compact(
		ctx,
		&recSig,
		(*C.uchar)(unsafe.Pointer(&rs[0])),
		C.int(v),
	) != 1 {
		return "", errors.New("không parse được recoverable signature")
	}

	// Recover public key
	var pubkey C.my_secp256k1_pubkey
	if C.my_secp256k1_ecdsa_recover(
		ctx,
		&pubkey,
		&recSig,
		(*C.uchar)(unsafe.Pointer(&hashBytes[0])),
	) != 1 {
		return "", errors.New("không recover được public key")
	}

	// Serialize dạng uncompressed (65 bytes)
	var outputLen C.size_t = 65
	output := make([]byte, outputLen)
	if C.my_secp256k1_ec_pubkey_serialize(
		ctx,
		(*C.uchar)(unsafe.Pointer(&output[0])),
		&outputLen,
		&pubkey,
		C.SECP256K1_EC_UNCOMPRESSED,
	) != 1 {
		return "", errors.New("không serialize được public key")
	}

	return hex.EncodeToString(output), nil
}

func SignRecoverable(rawHashHex, rawPrivHex string) (string, error) {
	hashBytes, err := hex.DecodeString(rawHashHex)
	if err != nil || len(hashBytes) != 32 {
		return "", errors.New("hash không hợp lệ")
	}

	privBytes, err := hex.DecodeString(rawPrivHex)
	if err != nil || len(privBytes) != 32 {
		return "", errors.New("private key không hợp lệ")
	}

	ctx := C.get_context()

	// Ký ECDSA recoverable
	var sig C.my_secp256k1_ecdsa_recoverable_signature
	if C.my_secp256k1_ecdsa_sign_recoverable(
		ctx,
		&sig,
		(*C.uchar)(unsafe.Pointer(&hashBytes[0])),
		(*C.uchar)(unsafe.Pointer(&privBytes[0])),
		C.my_secp256k1_nonce_function_rfc6979,
		nil,
	) != 1 {
		return "", errors.New("không thể ký")
	}

	// Serialize r||s (64 byte) và recId (v)
	var rs [64]byte
	var recID C.int
	if C.my_secp256k1_ecdsa_recoverable_signature_serialize_compact(
		ctx,
		(*C.uchar)(unsafe.Pointer(&rs[0])),
		&recID,
		&sig,
	) != 1 {
		return "", errors.New("không thể serialize recoverable signature")
	}

	// Ghép r||s||v = 65 byte
	sig65 := append(rs[:], byte(recID)) // Ethereum: v = recID hoặc (recID + 27)

	return hex.EncodeToString(sig65), nil
}
