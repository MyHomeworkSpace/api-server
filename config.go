package main

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

var config Config

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Email     EmailConfig
	Redis     RedisConfig
	CORS      CORSConfig
	Feedback  FeedbackConfig
	Whitelist WhitelistConfig
}

type ServerConfig struct {
	Port               int
	AuthURLBase        string
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

type WhitelistConfig struct {
	Enabled       bool
	WhitelistFile string
	BlockMessage  string
}

func CreateNewConfig() {
	newConfig := `# MyHomeworkSpace configuration
[server]
Port = 3000
AuthURLBase = "http://myhomework.space/applicationAuth.html"
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
SlackHostName = ""

[whitelist]
Enabled = false
WhitelistFile = "whitelist.txt"
BlockMessage = "You aren't allowed in"`
	err := ioutil.WriteFile("config.toml", []byte(newConfig), 0644)
	if err != nil {
		panic(err)
	}
}

func InitConfig() {
	if _, err := os.Stat("config.toml"); err != nil {
		CreateNewConfig() // create new config to be parsed
	}
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		panic(err)
	}
}
