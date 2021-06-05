package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"time"
)

var (
	appName            = "web"
	configuration      *Configuration
	prodConfigFileName = "config_prod.yaml"
	devConfigFileName  = "config_dev.yaml"
	configType         = "yaml"
	configDir          = "."
)

type Configuration struct {
	IsProduction bool
	Server       ServerConfig
	Database     DbConfig
	Redis        RedisConfig
}

type ServerConfig struct {
	Domain       string `default:"http://localhost"`
	Port         string `default:"3000"`
	Environment  string `default:"development"` // development,test,production
	JwtSecret    string `default:"bisi-bisi-bisi"`
	ReadTimeout  int    `default:"5"`
	WriteTimeout int    `default:"10"`
	IdleTimeout  int    `default:"120"`
	LogPath      string `default:"stdout"`
	LogLevel     string `default:"info"`
	UploadDir    string `default:"./upload"`
}

type DbConfig struct {
	Name        string `default:"web"`
	Username    string `default:"web"`
	Password    string `default:"web?????"`
	Host        string `default:"localhost"`
	Port        string `default:"5432"`
	LogMode     bool   `default:"false"`
	MaxPoolSize int    `default:"10"`
	MaxIdleConn int    `default:"1"`
}

type RedisConfig struct {
	Host         string
	Password     string
	Db           int
	WriteTimeout time.Duration
}

func setDefaults() {
	viper.SetDefault("isProduction", os.Getenv("APP_ENV") == "PRODUCTION")
	viper.SetDefault("server.domain", "http://localhost")
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.logPath", "stdout")
	viper.SetDefault("server.logLevel", "debug")
	viper.SetDefault("server.secret", "bisi-bisi-bisi")
	viper.SetDefault("server.readTimeout", 5)
	viper.SetDefault("server.writeTimeout", 10)
	viper.SetDefault("server.IdleTimeout", 120)
	viper.SetDefault("server.uploadDir", "./upload")

	viper.SetDefault("database.name", "web")
	viper.SetDefault("database.username", "web")
	viper.SetDefault("database.password", "web?????")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.logmode", false)
	viper.SetDefault("database.maxPoolSize", 10)
	viper.SetDefault("database.maxIdleConn", 1)

	viper.SetDefault("redis.host", "web.com:6379")
	viper.SetDefault("redis.password", "asdf")
	viper.SetDefault("redis.db", 9)
	viper.SetDefault("redis.writeTimeout", time.Second*5)
}

// os environment'ından okumak için. şimdilik çokda lazım değil..
func bindEnvs() {
	viper.BindEnv("server.domain", "DOMAIN")
	viper.BindEnv("server.port", "PORT")
	viper.BindEnv("server.environment", "ENVIRONMENT")
	viper.BindEnv("server.logPath", "LOG_PATH")
	viper.BindEnv("server.secret", "SERVER_SECRET")
	viper.BindEnv("server.timeout", "SERVER_TIMEOUT")

	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.username", "DB_USERNAME")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.logmode", "DB_LOG_MODE")
}

// read configuration from file
func readConfiguration(path, filename string) error {
	dir, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return errors.Wrapf(err, "Abssulte path oluşturulamadı.")
	}
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		// if file does not exist, simply create one

		if _, err := os.Stat(dir + "/" + filename); os.IsNotExist(err) {
			os.Create(dir + "/" + filename)
		} else if err != nil {
			return errors.Wrapf(err, "Config dosyası olşuturulamadı.")
		}
		// let's write defaults
		if err2 := viper.WriteConfig(); err2 != nil {
			return errors.Wrapf(err2, "Yeni oluşturulan Config dosyasına yazılamadı.")
		}
	}
	return nil
}

func Setup() (*Configuration, error) {
	var configFilename string
	if os.Getenv("APP_ENV") == "PRODUCTION" {
		configFilename = prodConfigFileName
	} else {
		configFilename = devConfigFileName
	}

	viper.SetConfigName(configFilename)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configDir)

	//bindEnvs()
	setDefaults()

	if err := readConfiguration(configDir, configFilename); err != nil {
		return nil, err
	}

	// Auto read env variables
	viper.AutomaticEnv()

	// Unmarshal config file to struct
	if err := viper.Unmarshal(&configuration); err != nil {
		return nil, err
	}

	return configuration, nil
}

func Get() *Configuration {
	return configuration
}
