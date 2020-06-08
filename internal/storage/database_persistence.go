//Package storage provides functionality for storing the data needed by the url shortener urlshortener_service.
package storage

import (
	"fmt"
	"github.com/gdgenchev/urlshortener/internal/common/config"
	"github.com/gdgenchev/urlshortener/internal/common/urldata"
	my "github.com/go-mysql/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// DatabasePersistence provides a common interface for the long term url data persistence.
type DatabasePersistence interface {
	SaveUrlData(urlData urldata.UrlData) bool
	GetUrlData(shortUrl string) (urldata.UrlData, bool)
	Exists(shortSlug string) bool
	Close()
}

// MysqlPersistence is a concrete implementation of the DatabasePersistence.
type MysqlPersistence struct {
	db *gorm.DB
}

func NewMysqlPersistence(configuration config.Configuration) *MysqlPersistence {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?loc=Local&parseTime=True", configuration.Mysql.User,
		configuration.Mysql.Password, configuration.Mysql.Host, configuration.Mysql.Port, configuration.Mysql.Database)
	db, err := gorm.Open(configuration.Mysql.DriverName, connectionString)
	if err != nil {
		panic(err)
	}

	mysqlPersistence := new(MysqlPersistence)
	mysqlPersistence.db = db

	mysqlPersistence.init()

	return mysqlPersistence
}

func (mysqlPersistence *MysqlPersistence) init() {
	mysqlPersistence.db.AutoMigrate(urldata.UrlData{})
	mysqlPersistence.db.Exec("CREATE EVENT IF NOT EXISTS expires_check ON SCHEDULE EVERY 1 DAY DO DELETE FROM url_data WHERE expires <= NOW()")
}

// SaveUrlData saves the url data in the database.
// Returns true if successful and false if the url short slug already exists
func (mysqlPersistence *MysqlPersistence) SaveUrlData(urlData urldata.UrlData) bool {
	//Workaround for an expired url, but not yet deleted by the mysql event
	mysqlPersistence.deleteUrlDataIfExpired(urlData.ShortSlug)
	err := mysqlPersistence.db.Create(&urlData).Error

	return err == my.ErrDupeKey
}

// GetRealUrlData retrieves the url data given a short slug.
// It checks only valid urls(which have not expired).
func (mysqlPersistence *MysqlPersistence) GetUrlData(shortSlug string) (urldata.UrlData, bool) {
	var urlData urldata.UrlData
	found := !mysqlPersistence.db.
		Where("short_slug = ?", shortSlug).
		Where("expires > NOW()").
		First(&urlData).
		RecordNotFound()

	return urlData, found
}

// Exists checks whether the short slug is present in the database.
// It also deletes an existent entry if it has expired.
func (mysqlPersistence *MysqlPersistence) Exists(shortSlug string) bool {
	_, found := mysqlPersistence.GetUrlData(shortSlug)
	return found == true
}

// Close closes the database client.
func (mysqlPersistence *MysqlPersistence) Close() {
	err := mysqlPersistence.db.Close()
	if err != nil {
		panic(err)
	}
}

func (mysqlPersistence *MysqlPersistence) deleteUrlDataIfExpired(shortSlug string) {
	err := mysqlPersistence.db.Where("short_slug = ?", shortSlug).Where("expires < NOW()").Delete(urldata.UrlData{}).Error
	if err != nil {
		panic(err)
	}
}
