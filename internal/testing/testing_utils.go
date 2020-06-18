package testing_utils

import (
	"context"
	"fmt"
	"github.com/gdgenchev/urlshortener/internal/model"
	"github.com/gdgenchev/urlshortener/internal/util"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	"strconv"
)

const testingConfigFilePath = "../../config/config.testing.json"

type TestPersistence struct {
	configuration util.Configuration
	db            *gorm.DB
	redisClient   *redis.Client
}

func NewTestPersistence() *TestPersistence {
	configuration := util.ReadConfiguration(testingConfigFilePath)

	testPersistence := new(TestPersistence)
	testPersistence.configuration = configuration
	testPersistence.initTestDatabase()
	testPersistence.initTestCache()

	return testPersistence
}

func (testPersistence *TestPersistence) GetTestConfiguration() util.Configuration {
	return testPersistence.configuration
}

func (testPersistence *TestPersistence) initTestDatabase() {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/?loc=Local&parseTime=True",
		testPersistence.configuration.Mysql.User, testPersistence.configuration.Mysql.Password,
		testPersistence.configuration.Mysql.Host, testPersistence.configuration.Mysql.Port)

	var err error
	testPersistence.db, err = gorm.Open(testPersistence.configuration.Mysql.DriverName, connectionString)
	if err != nil {
		panic(err)
	}

	err = testPersistence.db.Exec("CREATE DATABASE IF NOT EXISTS " + testPersistence.configuration.Mysql.Database).Error
	if err != nil {
		panic(err)
	}

	err = testPersistence.db.Exec("USE " + testPersistence.configuration.Mysql.Database).Error
	if err != nil {
		panic(err)
	}

	err = testPersistence.db.DropTableIfExists(model.UrlData{}).Error
	if err != nil {
		panic(err)
	}
}

func (testPersistence *TestPersistence) initTestCache() {
	testPersistence.redisClient = redis.NewClient(&redis.Options{
		Addr:     testPersistence.configuration.Redis.Host + ":" + strconv.Itoa(testPersistence.configuration.Redis.Port),
		Password: testPersistence.configuration.Redis.Password,
		DB:       testPersistence.configuration.Redis.DB,
	})

	testPersistence.redisClient.FlushAll(context.Background())
}

func (testPersistence *TestPersistence) FlushTestPersistence() error {
	err := testPersistence.db.DropTableIfExists(&model.UrlData{}).Error
	if err != nil {
		return err
	}

	err = testPersistence.db.AutoMigrate(&model.UrlData{}).Error
	if err != nil {
		return err
	}

	testPersistence.FlushTestCache()

	return nil
}

func (testPersistence *TestPersistence) CleanUp() {
	testPersistence.db.Exec("DROP DATABASE " + testPersistence.configuration.Mysql.Database)
	testPersistence.redisClient.FlushAll(context.Background())
	testPersistence.db.Close()
	testPersistence.redisClient.Close()
}

func (testPersistence *TestPersistence) FlushTestCache() {
	testPersistence.redisClient.FlushAll(context.Background())
}

func (testPersistence *TestPersistence) ExistsInTestCache(shortSlug string) bool {
	exists, err := testPersistence.redisClient.Exists(context.Background(), shortSlug).Result()
	if err != nil {
		panic(err)
	}

	return exists == 1
}

func (testPersistence *TestPersistence) Close() {
	testPersistence.db.Close()
	testPersistence.redisClient.Close()
}
