package config

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

var config Config

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Email    EmailConfig
	Redis    RedisConfig
	CORS     CORSConfig
	Feedback FeedbackConfig
}

type ServerConfig struct {
	Port               int
	AuthURLBase        string
	APIURLBase         string
	AppURLBase         string
	ReverseProxyHeader string
}

type DatabaseConfig struct {
	Host     string
	Username string
	Password string
	Database string
}

type EmailConfig struct {
	Enabled      bool
	From         string
	SMTPHost     string
	SMTPPort     int
	SMTPSecure   bool
	SMTPUsername string
	SMTPPassword string
}

type RedisConfig struct {
	Host string
	Port int
}

type CORSConfig struct {
	Enabled bool
	Origins []string
}

type FeedbackConfig struct {
	SlackEnabled  bool
	SlackURL      string
	SlackHostName string
}

func createNewConfig() {
	newConfig := `# MyHomeworkSpace configuration
[server]
Port = 3000
AuthURLBase = "http://myhomework.space/applicationAuth.html"
APIURLBase = "http://api-v2.myhomework.space/"
AppURLBase = "http://myhomework.space/app.html#!"
ReverseProxyHeader = ""

[database]
Host = "localhost"
Username = "myhomeworkspace"
Password = "myhomeworkspace"
Database = "myhomeworkspace"

[email]
Enabled = false
From = "Misconfigured MyHomeworkSpace <misconfigured@misconfigured.invalid>"
SMTPHost = "localhost"
SMTPPort = 465
SMTPSecure = true
SMTPUsername = "accounts@myhomework.space"
SMTPPassword = "password123"

[redis]
Host = "localhost"
Port = 6379

[cors]
Enabled = false
Origins = [ "http://myhomework.space" ]

[feedback]
SlackEnabled = false
SlackURL = ""
SlackHostName = ""`
	err := ioutil.WriteFile("config.toml", []byte(newConfig), 0644)
	if err != nil {
		panic(err)
	}
}

// GetCurrent returns a pointer to the active configuration struct
func GetCurrent() *Config {
	return &config
}

// Init loads the configuration from disk, creating it if needed
func Init() {
	if _, err := os.Stat("config.toml"); err != nil {
		createNewConfig() // create new config to be parsed
	}
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		panic(err)
	}
}
