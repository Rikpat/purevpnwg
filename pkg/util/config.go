package util

import (
	"os"

	"github.com/go-rod/rod"
	"gopkg.in/yaml.v3"
)

const CONFIG_FILE = "config.yml"

type SubscriptionAuth struct {
	Username string `flag:"" yaml:"username" env:"ID" help:"PureVPN subscription ID"`
	Password string `flag:"" yaml:"password" env:"PASSWORD" help:"PureVPN subscription password (not necessary)"`
}

type ServerConfig struct {
	Country string `flag:"" yaml:"country" env:"COUNTRY" help:"PureVPN server country (example: US)."`
	City    string `flag:"" yaml:"city" env:"CITY" help:"PureVPN server city (example: 8778 for New York)."`
}

type Config struct {
	Debug         bool              `flag:"" yaml:"debug" help:"Enable debug mode."`
	Username      string            `flag:"" yaml:"username" required:"" env:"USERNAME" help:"PureVPN username (email)."`
	Password      string            `flag:"" yaml:"password" required:"" env:"PASSWORD" help:"PureVPN password."`
	UUID          string            `yaml:"uuid" kong:"-"`
	Server        *ServerConfig     `yaml:"server" embed:"" prefix:"server." envprefix:"SERVER_"`
	Subscription  *SubscriptionAuth `embed:"" prefix:"subscription." envprefix:"SUBSCRIPTION_"`
	Device        string            `yaml:"device" env:"DEVICE" default:"linux"`
	WireguardFile string            `flag:"" yaml:"wireguardFile" env:"WIREGUARD_FILE" default:"wg0.conf"`
	OvpnPath      string            `flag:"" yaml:"ovpnPath" env:"OVPN_PATH" default:"vpn.conf" help:"Path to save the OpenVPN configuration file."`
}

func (sub *SubscriptionAuth) GetEncryptPassword(page *rod.Page, token string) error {
	res, err := page.Eval(`
		async (username, authorization) => {
			const res = await fetch(
				"/v2/api/wireguard/get-encrypt-password",
				{
					method: "POST",
					body: new URLSearchParams({username}).toString(),
					headers: {
						'content-type': 'application/x-www-form-urlencoded',
						accept: 'application/json',
						authorization
					}
				}
			)
		if (!res.ok) {
			throw Error(await res.text())
			return
		}
		const json = await res.json()
		if (!json.status) {
			throw Error(json.message)
		}
		return json.body.encrypPass
	}`, sub.Username, "Bearer "+token)
	if err == nil {
		sub.Password = res.Value.Str()
	}
	return err
}

func (conf *Config) Save() error {
	return WriteConfig(conf)
}

func WriteConfig(conf *Config) error {
	data, err := yaml.Marshal(&conf)
	if err == nil {
		err = os.WriteFile(CONFIG_FILE, data, 0644)
	}
	return err
}

func ReadConfig() (*Config, error) {
	conf := Config{}

	data, err := os.ReadFile(CONFIG_FILE)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal([]byte(data), &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func WriteServerConfig(filePath, config string) error {
	return os.WriteFile(filePath, []byte(config), 0644)
}
