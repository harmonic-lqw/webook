package dao

import "gorm.io/gorm"

// InitTables 使用GORM自带的建表功能
// 这是种不太好的做法
func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{},
		&AsyncSms{},
	)
}
