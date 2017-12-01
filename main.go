package main

import (
	"errors"
	"fmt"
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

	kii.Logger = log.New(os.Stderr, "", log.LstdFlags)
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
			Usage: "Specifiy app name configured in config file",
		},
		cli.StringFlag{
			Name:  "config",
			Usage: "Specify path of yml format config file",
		},
	}

	app.Run(os.Args)
}

func initWithConfig(configFile string) error {
	if configFile == "" {
		configFile = "./config.yml"
	}
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return errors.New("can't read " + configFile + " file.")
	}
	err = yaml.Unmarshal(b, &gConfig)
	if err != nil {
		return errors.New("can't unmarshal " + configFile + ".")
	}

	dbFile := gConfig.DB
	if dbFile == "" {
		dbFile = "manager.db"
	}
	fmt.Println(dbFile)
	db, err = bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return errors.New("can't open db: " + dbFile)
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
		return errors.New(fmt.Sprintf("can't create bucket: %s", err))
	}
	return nil
}
