package filter

import (
	"fmt"
	"goldenglow/dataBase"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func filterPGWithGORM(db *gorm.DB) ([]dataBase.Node, error) {
	var results []dataBase.Node
	err := db.Where("content !~ ?", "^\\[\\d+\\]").
		Where("content NOT LIKE ?", "[check]%").
		Where("content NOT LIKE ?", "[P]%").
		Where("content !~ ?", "^\\[0x[-a-zA-Z0-9*]+?\\] :").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
func FilterNV() []dataBase.Node {
	res, _ := filterPGWithGORM(gormDb)
	return res
}

var (
	gormDb *gorm.DB
)

func init() {
	dsn := "host=localhost user=postgres password=gg dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("连接数据库失败: %v", err))
	}
	gormDb = db
}
