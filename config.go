package main

import (
	_ "github.com/caarlos0/env/v6"
	"net/http"
)

type Config struct {
	Port              int    `env:"PORT" envDefault:"80"`
	DiscordWebhookUrl string `env:"DISCORD_WEBHOOK_URL" envDefault:"https://discordapp.com/api/webhooks/TOKEN/ID"`
	SecurityToken     string `env:"SECURITY_TOKEN" envDefault:"PUT_YOUR_SECURITY_TOKEN_HERE"`
}

var config = Config{}
var client = &http.Client{}
