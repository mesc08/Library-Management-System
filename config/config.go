package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

var ViperConfig Config

type Config struct {
	RedisHost     string
	MysqlHost     string
	MysqlPort     int
	MysqlUser     string
	MysqlPassword string
	MysqlDBName   string
	SecretKey     string
}

func InitConfig(fileName string) error {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		logrus.Errorln("Error reading json file")
		return err
	}
	err = json.Unmarshal(content, &ViperConfig)
	if err != nil {
		logrus.Errorln("Error unmarshalling data")
		return err
	}
	return nil
}
