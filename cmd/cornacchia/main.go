package main

import (
	"flag"
	"github.com/piger/cornacchia"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var (
	configFile = flag.String("config", "bot.yml", "Path to the configuration file")
)

type config struct {
	Listen        string `yaml:"listen"`
	MattermostURL string `yaml:"mattermost_url"`
}

func main() {
	flag.Parse()
	var cfg config

	configData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(configData, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting server on %s\n", cfg.Listen)
	cornacchia.StartServer(cfg.Listen, cfg.MattermostURL)
}
