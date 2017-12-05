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
	DB             string         `yaml:"db"`
}

type GatewayAddress struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}
type App struct {
	ID   string `yaml:"app-id"`
	Key  string `yaml:"app-key"`
	Site string `yaml:"app-site"`
	Host string `yaml:"app-host"`
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
	var configFile string
	if os.Getenv("GWM_CONFIG_PATH") != "" {
		configFile = os.Getenv("GWM_CONFIG_PATH")
	} else {
		configFile = "./config.yml"
	}
	kii.Logger = log.New(os.Stderr, "", log.LstdFlags)
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalln("can't read "+configFile+" file.", err)
	}
	err = yaml.Unmarshal(b, &gConfig)
	if err != nil {
		log.Fatalln("can't unmarshal "+configFile, err)
	}

	dbFile := gConfig.DB
	if dbFile == "" {
		dbFile = "manager.db"
	}

	db, err = bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatalln("can't open " + dbFile)
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
	app.Usage = "Sample app shows how to manage Gateway. Specify the path of config file with env variable GWM_CONFIG_PATH" +
		"when config file located in different folder with binary file"
	app.Author = "Kii Corporation"
	app.Email = "support@kii.com"
	app.Commands = Commands
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "app-name",
			Usage: "Specifiy app name configured in config file",
		},
	}

	app.Run(os.Args)
}
