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
	ErrorLog SlackConfig
	Feedback SlackConfig
}

type ServerConfig struct {
	Port               int
	APIURLBase         string
	AppURLBase         string
	ReverseProxyHeader string
	HostName           string
}

type DatabaseConfig struct {
	Host     string
	Username string
	Password string
	Database string
}

type EmailConfig struct {
	Enabled      bool
	FromAddress  string
	FromDisplay  string
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

type SlackConfig struct {
	SlackEnabled bool
	SlackURL     string
}

func createNewConfig() {
	newConfig := `# MyHomeworkSpace configuration
[server]
Port = 3000
APIURLBase = "http://api-v2.myhomework.space/"
AppURLBase = "http://myhomework.space/app.html#!"
ReverseProxyHeader = ""
HostName = "local"

[database]
Host = "localhost"
Username = "myhomeworkspace"
Password = "myhomeworkspace"
Database = "myhomeworkspace"

[email]
Enabled = false
FromAddress = "misconfigured@misconfigured.invalid"
FromDisplay = "Misconfigured MyHomeworkSpace <misconfigured@misconfigured.invalid>"
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

[errorlog]
SlackEnabled = false
SlackURL = ""

[feedback]
SlackEnabled = false
SlackURL = ""`
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
