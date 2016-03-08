package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/KiiPlatform/kii_go"
	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Apps           map[string]App `yaml:"apps"`
	GatewayAddress GatewayAddress `yaml:"gateway-address"`
}

type GatewayAddress struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}
type App struct {
	ID   string `yaml:"app-id"`
	Key  string `yaml:"app-key"`
	Site string `yaml:"app-site"`
}

type User struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type Node struct {
	ID  string `json:"id"`
	VID string `json:"vid"`
}

// Global variables. :(
var gConfig Config
var db *bolt.DB

func main() {

	kii.Logger = log.New(os.Stderr, "", log.LstdFlags)
	b, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		log.Fatalln("can't read ./config.yml file.")
	}
	err = yaml.Unmarshal(b, &gConfig)
	if err != nil {
		log.Fatalln("can't unmarshal ./config.yml")
	}

	db, err = bolt.Open("manager.db", 0600, nil)
	if err != nil {
		log.Fatalln("can't open db")
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("tokens"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("gateway-ids"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatalln("can't create bucket: ", err)
	}
	app := cli.NewApp()
	app.Name = "gw-manager"
	app.Version = "1.0.0"
	app.Usage = "Sample app shows how to manage Gateway"
	app.Author = "Kii Corporation"
	app.Email = "support@kii.com"
	app.Commands = Commands
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "app-name",
			Usage: "Specifiy app name configured in config.yml",
		},
	}

	app.Run(os.Args)
}
