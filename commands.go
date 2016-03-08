package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"log"
)

var Commands = []cli.Command{
	userLogin,
	auth,
	onboardGateway,
	addOwner,
	listPendingNodes,
	onboardNode,
	postCommand,
	restore,
	replaceNode,
	showDB,
}

var userLogin = cli.Command{
	Name:      "user-login",
	Usage:     "user-login --username <user name> --password <password> --app-name <app name>",
	UsageText: "Kii Cloud User Login. This user will be an owner of the Gateway",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "username",
			Usage: "Gateway owner user name (Kii Cloud User)",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "Gateway owner password (Kii Cloud User)",
		},
		cli.StringFlag{
			Name: "app-name",
		},
	},
	Action: func(c *cli.Context) {
		username := c.String("username")
		password := c.String("password")
		appName := c.String("app-name")
		if username == "" {
			log.Fatalln("no username is specified")
		}
		if password == "" {
			log.Fatalln("no password is specified")
		}
		if appName == "" {
			log.Fatalln("no app-name is specified")
		}
		app := gConfig.Apps[appName]
		userID, userToken, err := _userLogin(app, username, password)
		if err != nil {
			log.Fatalln("failed to login with the user: ", err)
		}
		user := User{
			ID:    userID,
			Token: userToken,
		}
		j, _ := json.Marshal(user)
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("users"))
			err = b.Put([]byte(appName), j)
			if err != nil {
				log.Println(err)
			}
			return err
		})
		if err != nil {
			log.Fatalln("failed to store user: ", err)
		}
	},
}

var auth = cli.Command{
	Name:      "auth",
	Usage:     "auth --username <user name> --password <password> --app-name <app name>",
	UsageText: "gateway local rest api authentication",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "username",
			Usage: "Gateway admin user name",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "Gateway admin password",
		},
		cli.StringFlag{
			Name: "app-name",
		},
	},
	Action: func(c *cli.Context) {
		username := c.String("username")
		password := c.String("password")
		appName := c.String("app-name")
		if username == "" {
			log.Fatalln("no username is specified")
		}
		if password == "" {
			log.Fatalln("no password is specified")
		}
		if appName == "" {
			log.Fatalln("no app-name is specified")
		}

		app := gConfig.Apps[appName]
		addr := gConfig.GatewayAddress
		token, err := localAuth(addr, app, username, password)
		if err != nil {
			log.Fatalln("local rest api authenticatoin error: %+v", err)
		}
		log.Println("token: ", token)
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("tokens"))
			err = b.Put([]byte(appName), []byte(token))
			if err != nil {
				log.Println(err)
			}
			return err
		})
		if err != nil {
			log.Fatalln("failed to store token")
		}
	},
}

var onboardGateway = cli.Command{
	Name:  "onboard-gateway",
	Usage: "onboard-gateway --app-name <app name>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "app-name",
		},
	},
	Action: func(c *cli.Context) {
		appName := c.String("app-name")
		if appName == "" {
			log.Fatalln("no app-name is specified")
		}
		var token string
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("tokens"))
			v := b.Get([]byte(appName))
			token = string(v[:])
			return nil
		})
		log.Printf("token %s\n", token)
		if token == "" {
			log.Fatalln("no auth token is stored for the specified app.")
		}
		app := gConfig.Apps[appName]
		addr := gConfig.GatewayAddress
		id, err := _onboardGateway(addr, app, token)
		if err != nil {
			log.Fatalln("local rest api authenticatoin error: %+v", err)
		}
		log.Printf("id %s\n", id)
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("gateway-ids"))
			err = b.Put([]byte(appName), []byte(id))
			if err != nil {
				log.Println(err)
			}
			return err
		})
		if err != nil {
			log.Fatalln("failed to store id")
		}
	},
}

var addOwner = cli.Command{
	Name:  "add-owner",
	Usage: "add-owner --gateway-password <gateway password> --app-name <app name>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "gateway-password",
			Usage: "Password of the gateway. It is configured in coonfig file of Gateway Agent.",
		},
		cli.StringFlag{
			Name: "app-name",
		},
	},
	Action: func(c *cli.Context) {
		gatewayPassword := c.String("gateway-password")
		appName := c.String("app-name")
		if gatewayPassword == "" {
			log.Fatalln("no gateway-password specified.")
		}
		if appName == "" {
			log.Fatalln("no app-name specified.")
		}
		var id string
		var user User
		err := db.View(func(tx *bolt.Tx) error {
			// Retrieve gateway id.
			b := tx.Bucket([]byte("gateway-ids"))
			v := b.Get([]byte(appName))
			id = string(v[:])

			// Retrieve user.
			b2 := tx.Bucket([]byte("users"))
			v2 := b2.Get([]byte(appName))
			err := json.Unmarshal(v2, &user)
			return err
		})
		if id == "" {
			log.Fatalln("no gateway-id is stored. please execute onboard-gateway.")
		}
		if err != nil {
			log.Fatalln("no login user is stored. please execute login-user.: ", err)
		}
		log.Println("gateway thing id: ", id)
		log.Println("user: ", user)
		app := gConfig.Apps[appName]

		err = _addOwner(app, user.ID, user.Token, id, gatewayPassword)
		if err != nil {
			log.Fatalln("failed to store id. ", err)
		}
	},
}

var listPendingNodes = cli.Command{
	Name:      "list-pending-nodes",
	Usage:     "list-pending-nodes --app-name <app name>",
	Aliases:   []string{"l"},
	UsageText: "List end-nodes connected to the gateway but haven't been onboarded to the cloud.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "app-name",
		},
	},
	Action: func(c *cli.Context) {
		appName := c.String("app-name")
		var token string
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("tokens"))
			v := b.Get([]byte(appName))
			token = string(v[:])
			return nil
		})
		if err != nil || token == "" {
			log.Fatalln("token is not stored for the specified app")
		}
		app := gConfig.Apps[appName]
		addr := gConfig.GatewayAddress
		l, err := _listPendingNodes(addr, app, token)
		if err != nil {
			log.Fatalln("can not list pending nodes: ", err)
		}
		log.Printf("pending nodes: \n%v", l)
	},
}

var onboardNode = cli.Command{
	Name:      "onboard-node",
	Usage:     "onboard-node --node-vid <end-node vendor thing id> --node-password <end-node password> --app-name <app name>",
	Aliases:   []string{"on"},
	UsageText: "Execute onboard for specified end-node",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "node-vid",
			Usage: "end node vendor thing id",
		},
		cli.StringFlag{
			Name:  "node-password",
			Usage: "end node password",
		},
		cli.StringFlag{
			Name: "app-name",
		},
	},
	Action: func(c *cli.Context) {
		nodeVID := c.String("node-vid")
		nodePass := c.String("node-password")
		appName := c.String("app-name")
		var gatewayID string
		var user User
		var token string
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("gateway-ids"))
			v := b.Get([]byte(appName))
			gatewayID = string(v[:])

			b2 := tx.Bucket([]byte("users"))
			v2 := b2.Get([]byte(appName))
			err := json.Unmarshal(v2, &user)
			if err != nil {
				return err
			}

			b3 := tx.Bucket([]byte("tokens"))
			v3 := b3.Get([]byte(appName))
			token = string(v3[:])
			return nil
		})
		if gatewayID == "" {
			log.Fatalln("gateway id is not stored for the specified app. execute onboard-gateway.")
		}
		if err != nil {
			log.Fatalln("no login user. execute user-login.")
		}
		if token == "" {
			log.Fatalln("token is not stored for the specified app. execute auth.")
		}
		app := gConfig.Apps[appName]
		nodeID, err := _onboardNode(app, user, gatewayID, nodeVID, nodePass)
		node := Node{
			ID:  nodeID,
			VID: nodeVID,
		}

		// Store end-node mapping.
		err = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("nodes:" + appName))
			if err != nil {
				return err
			}
			err = b.Put([]byte(node.VID), []byte(node.ID))
			return err
		})
		if err != nil {
			log.Fatalln("failed to store end-node: ", err)
		}

		// Tell End Node mapping to Gateway Agent.
		addr := gConfig.GatewayAddress
		err = _mapNode(addr, app, node, token)
		if err != nil {
			log.Fatalln("failed to map end-node: ", err)
		}
	},
}

var postCommand = cli.Command{
	Name:      "post-command",
	Usage:     "post-command --node-vid <end-node vendor thing id> --command-file <filename> --app-name <app name>",
	Aliases:   []string{"s"},
	UsageText: "Post command to the specified end-node.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "node-vid",
			Usage: "end node vendor thing id",
		},
		cli.StringFlag{
			Name:  "command-file",
			Value: "command.json",
			Usage: "file path describes command in json format.",
		},
		cli.StringFlag{
			Name: "app-name",
		},
	},
	Action: func(c *cli.Context) {
		nodeVID := c.String("node-vid")
		path := c.String("command-file")
		appName := c.String("app-name")

		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalln("can not read command-file: ", err)
		}

		var user User
		var nodeID string
		err = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("users"))
			v := b.Get([]byte(appName))
			err := json.Unmarshal(v, &user)
			if err != nil {
				return err
			}

			b2 := tx.Bucket([]byte("nodes:" + appName))
			v2 := b2.Get([]byte(nodeVID))
			nodeID = string(v2[:])
			return nil
		})
		node := Node{
			ID:  nodeID,
			VID: nodeVID,
		}
		if err != nil {
			log.Fatalln("can not find user in specified app. execute user-login. ", err)
		}
		if nodeID == "" {
			log.Fatalln("can not find end-node. execute onboard-node")
		}
		app := gConfig.Apps[appName]
		resp, err := _postCommand(app, user, node.ID, b)
		if err != nil {
			log.Fatalln("failed to post command: ", err)
		}
		log.Printf("post command resp: %v", resp)
	},
}

var restore = cli.Command{
	Name:      "restore",
	Usage:     "restore --app-name <app name>",
	Aliases:   []string{"r"},
	UsageText: "Restore the gateway. Gateway Agent should be started in restore mode.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "app-name",
		},
	},
	Action: func(c *cli.Context) {
		var token string
		appName := c.String("app-name")
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("tokens"))
			v := b.Get([]byte(appName))
			token = string(v[:])
			return nil
		})
		if token == "" {
			log.Fatalln("token is not stored for the specified app. execute auth.")
		}
		app := gConfig.Apps[appName]
		addr := gConfig.GatewayAddress
		err := _restore(addr, app, token)
		if err != nil {
			log.Fatalln("failed to restore: ", err)
		}
	},
}

var replaceNode = cli.Command{
	Name:      "replace-node",
	Usage:     "replace-node --node-vid <old end-node vendor thing id> --new-vid <new end-node vendor thing id> --node-password <end-node password> --app <app name>",
	Aliases:   []string{"rp"},
	UsageText: "Replace end-node hardware with new one.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "node-vid",
			Usage: "end node vendor thing id to be replaced.",
		},
		cli.StringFlag{
			Name:  "new-vid",
			Usage: "new end node vendor thing id.",
		},
		cli.StringFlag{
			Name:  "node-password",
			Usage: "end node password",
		},
		cli.StringFlag{
			Name: "app-name",
		},
	},
	Action: func(c *cli.Context) {
		nodeVID := c.String("node-vid")
		newVID := c.String("new-vid")
		nodePass := c.String("node-password")
		appName := c.String("app-name")
		var user User
		var nodeID string
		var token string
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("users"))
			v := b.Get([]byte(appName))
			err := json.Unmarshal(v, &user)
			if err != nil {
				return err
			}

			b2 := tx.Bucket([]byte("nodes:" + appName))
			v2 := b2.Get([]byte(nodeVID))
			nodeID = string(v2[:])

			b3 := tx.Bucket([]byte("tokens"))
			v3 := b3.Get([]byte(appName))
			token = string(v3[:])
			return nil
		})
		if err != nil {
			log.Fatalln("no user stored for the specified app. execute user-login.")
		}
		if nodeID == "" {
			log.Fatalln("no end-node is onboard with the specified VID. execute onboard-endnode.")
		}
		if token == "" {
			log.Fatalln("token is not stored for the specified app. execute auth.")
		}
		app := gConfig.Apps[appName]
		err = _updateVID(app, user, nodeID, newVID, nodePass)
		if err != nil {
			log.Fatalln("failed to update vendor thing id on Kii Cloud: ", err)
		}
		addr := gConfig.GatewayAddress
		node := Node{
			ID:  nodeID,
			VID: newVID,
		}
		err = _replaceNode(addr, app, node, token)
		if err != nil {
			log.Fatalln("failed to replace end-node on Gateway Agent: ", err)
		}
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("nodes:" + appName))
			// Remove the old entry if exit.
			b.Delete([]byte(nodeVID))
			// Put new entry.
			return b.Put([]byte(node.VID), []byte(node.ID))
		})
		if err != nil {
			log.Fatalln("failed to store end-node in db: ", err)
		}
	},
}

var showDB = cli.Command{
	Name:  "show-db",
	Usage: "show-db --bucket <bucket name>",
	UsageText: `show entries stored in the db.
	available bucket name:
	tokens - show stored tokens. `,
	Aliases: []string{"db"},

	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bucket",
			Usage: "bucket name of the DB",
		},
	},
	Action: func(c *cli.Context) {
		bucketName := c.String("bucket")
		if bucketName == "" {
			log.Fatalln("no bucket is specified")
		}
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucketName))
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				log.Printf("key: %s, value: %s\n", k, v)
			}
			return nil
		})
	},
}
