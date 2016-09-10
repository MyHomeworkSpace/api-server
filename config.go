package main

import (
    "io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

var config Config

type Config struct {
    Server ServerConfig
	Database DatabaseConfig
	Email EmailConfig
	CORS CORSConfig
	Whitelist WhitelistConfig
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
    Host string
	Username string
	Password string
	Database string
}

type EmailConfig struct {
	Enabled bool
	From string
    SMTPHost string
    SMTPPort int
    SMTPSecure bool
	SMTPUsername string
	SMTPPassword string
}

type CORSConfig struct {
	Enabled bool
	Origin string
}

type WhitelistConfig struct {
	Enabled bool
	WhitelistFile string
	BlockMessage string
}

func CreateNewConfig() {
	newConfig := `# MyHomeworkSpace configuration
[server]
Port = 3000

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

[cors]
Enabled = false
Origin = "http://myhomework.space"

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
