package dbsvc

import (
	"fmt"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)
func Open( path string) (*leveldb.DB,error){
	db, err := leveldb.OpenFile(path, nil)
    if err != nil {
        return nil, err
    }
    // defer db.Close()
	return db,err
}
func WriteValueStorage(call map[string]interface{},db *leveldb.DB) error {
	key, ok := call["key"].(string)
	data, ok := call["data"].(string)
	if !ok || key == "" || data == "" {
		return fmt.Errorf("invalid key or data")
	}

	// securyDb.Write(key, data)
	err := db.Put([]byte(key), []byte(data), nil)
	if err!=nil{
		return err
	}
	return nil
}
 func ReadValueStorage(call map[string]interface{},db *leveldb.DB) ([]byte,error ){
	key, ok := call["key"].(string)
	if !ok || key == "" {
		return nil,fmt.Errorf("Key not found")
	}
	value,err := db.Get([]byte(key), nil)
	if err != nil {
		return nil,err
	
	} 
	return value,nil
}
func DeleteKeyStorage(call map[string]interface{},db *leveldb.DB) map[string]interface{} {
	key, ok := call["key"].(string)
	if !ok || key == "" {
	return map[string]interface{}{
	"error": "EDKS-001",
	"message": "Key is required",
	}
	}
	// securyDb.Delete(key)
	err := db.Delete([]byte(key), nil)
	if err!=nil{
		log.Fatal()
	}
	return map[string]interface{}{
	"success": true,
	}
}