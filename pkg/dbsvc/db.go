package dbsvc

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/meta-node-blockchain/noti-contract/internal/model"
	"github.com/meta-node-blockchain/noti-contract/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var MySQLDB *gorm.DB

func StartMySQL(config *config.AppConfig) {
	db, err := sql.Open("mysql", config.MYSQL_URL)
	if err != nil {
		log.Fatal("Connect DB mysql Error", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Connect DB Gorm Error", err)
	}

	gormdb, err := gormDB.DB()
	if err != nil {
		log.Fatal("Connect DB Gorm Error", err)
	}

	migrations := []interface{}{
		&model.DeviceToken{},
	}

	// gormDB.Migrator().DropTable(migrations)
	gormDB.AutoMigrate(migrations...)
	// gormDB.Migrator().
	gormdb.SetMaxIdleConns(10)
	gormdb.SetMaxOpenConns(100)
	gormdb.SetConnMaxLifetime(time.Minute * 5)

	MySQLDB = gormDB

	fmt.Print("MYSQL DATABASE CONNECTED!\n")
}

func GetMySqlConn() *gorm.DB {
	return MySQLDB
}
