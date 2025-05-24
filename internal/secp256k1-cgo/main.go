// package main

// import (
// 	"fmt"
// 	"log"
// 	"secp256k1-cgo/secp"
// )

// func main() {
// 	priv := "726b5bf7e48b4fdf8e0423dcf30a962e8695eaa35a45e4a4c4aa55e392a0a9c7"
// 	pub := "04274132e9021a5d6103260ad397cd82c5f5bc16c8d627f8789a3a7258a7fe735ac19bac9055dd91b76076bcd0fc3593dfd51d9697ea8cfdc756e84039c19bfe5f"

// 	// ✅ Tạo public key từ private key (compressed)
// 	pubCompressed, err := secp.CreatePublicKey(priv, true)
// 	if err != nil {
// 		log.Fatalf("❌ CreatePublicKey (compressed) lỗi: %v", err)
// 	}
// 	fmt.Println("🔑 Public Key (compressed):", pubCompressed)

// 	// ✅ Tạo public key từ private key (uncompressed)
// 	pubUncompressed, err := secp.CreatePublicKey(priv, false)
// 	if err != nil {
// 		log.Fatalf("❌ CreatePublicKey (uncompressed) lỗi: %v", err)
// 	}
// 	fmt.Println("🔓 Public Key (uncompressed):", pubUncompressed)

// 	// ✅ Tạo ECDH shared secret
// 	shared, err := secp.CreateECDH(priv, pub)
// 	if err != nil {
// 		log.Fatal("❌ CreateECDH lỗi:", err)
// 	}
// 	fmt.Println("🧩 Shared secret (hex):", shared)

// 	// ✅ Hash (message đã hash SHA256 32 byte)
// 	hash := "5b11ecb3a1a93beee88b59708a24c46dc31f3d6220e2f65d3e0f82dc1ebd4c4c"

// 	// ✅ Bước 1: Tạo chữ ký recoverable
// 	sig65, err := secp.SignRecoverable(hash, priv)
// 	if err != nil {
// 		log.Fatal("❌ Ký lỗi:", err)
// 	}
// 	fmt.Println("✍️ Signature (r||s||v):", sig65)

// 	// ✅ Bước 2: Recover public key từ chữ ký
// 	recoveredPub, err := secp.RecoverPublicKey(hash, sig65)
// 	if err != nil {
// 		log.Fatal("❌ Recover lỗi:", err)
// 	}
// 	fmt.Println("🔄 Recovered pubkey:", recoveredPub)

// 	// ✅ Bước 3: So sánh với pubkey gốc tạo từ private key
// 	originalPub, err := secp.CreatePublicKey(priv, false)
// 	if err != nil {
// 		log.Fatal("❌ Tạo pubkey lỗi:", err)
// 	}
// 	fmt.Println("🎯 Original pubkey:  ", originalPub)

// 	// ✅ So sánh
// 	if originalPub == recoveredPub {
// 		fmt.Println("✅ MATCH: public key recovered chính xác!")
// 	} else {
// 		fmt.Println("❌ KHÔNG TRÙNG: kiểm tra lại logic ký")
// 	}

// }
