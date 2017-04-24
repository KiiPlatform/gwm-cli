package main

import (
	"encoding/json"

	"github.com/KiiPlatform/kii_go"
)

func location(app App) string {
	if len(app.Host) > 0 {
		return app.Host
	}
	return app.Site
}

func _updateVID(app App, user User, currentID string, newVID string, password string) error {
	author := kii.APIAuthor{
		App: kii.App{
			AppID:    app.ID,
			AppKey:   app.Key,
			Location: location(app),
		},
	}
	author.Token = user.Token
	req := kii.UpdateVendorThingIDRequest{
		VendorThingID: newVID,
		Password:      password,
	}
	return author.UpdateVendorThingID(currentID, req)
}

func _postCommand(app App, user User, nodeID string, command []byte) (resp *kii.PostCommandResponse, err error) {
	author := kii.APIAuthor{
		App: kii.App{
			AppID:    app.ID,
			AppKey:   app.Key,
			Location: location(app),
		},
	}
	author.Token = user.Token

	var req kii.PostCommandRequest
	err = json.Unmarshal(command, &req)
	// overwrite issuer.
	req.Issuer = "user:" + user.ID
	if err != nil {
		return
	}
	return author.PostCommand(nodeID, req)
}

func _onboardNode(app App, user User, gatewayID string, nodeVID string, nodePass string, thingType string, firmwareVersion string) (string, error) {
	author := kii.APIAuthor{
		App: kii.App{
			AppID:    app.ID,
			AppKey:   app.Key,
			Location: location(app),
		},
	}
	author.Token = user.Token
	req := kii.OnboardEndnodeWithGatewayThingIDRequest{
		GatewayThingID: gatewayID,
		OnboardEndnodeRequestCommon: kii.OnboardEndnodeRequestCommon{
			EndNodeVendorThingID:   nodeVID,
			EndNodePassword:        nodePass,
			Owner:                  "user:" + user.ID,
			EndNodeThingType:       thingType,
			EndNodeFirmwareVersion: firmwareVersion,
		},
	}
	resp, err := author.OnboardEndnodeWithGatewayThingID(req)
	if err != nil {
		return "", err
	}
	nodeID := resp.EndNodeThingID
	return nodeID, err
}

func _userLogin(app App, username string, password string) (id string, token string, err error) {
	author := kii.APIAuthor{
		App: kii.App{
			AppID:    app.ID,
			AppKey:   app.Key,
			Location: location(app),
		},
	}
	req := kii.UserRegisterRequest{
		LoginName: username,
		Password:  password,
	}
	author.RegisterKiiUser(req)
	req2 := kii.UserLoginRequest{
		UserName: username,
		Password: password,
	}
	resp, err := author.LoginAsKiiUser(req2)
	if err != nil {
		return
	}
	id = resp.ID
	token = resp.AccessToken
	return
}

func _addOwner(app App, userID string, userToken string, gatewayID string, gatewayPassword string) error {
	author := kii.APIAuthor{
		App: kii.App{
			AppID:    app.ID,
			AppKey:   app.Key,
			Location: location(app),
		},
	}
	author.Token = userToken
	req3 := kii.OnboardByOwnerRequest{
		ThingID:       gatewayID,
		ThingPassword: gatewayPassword,
		Owner:         "user:" + userID,
	}
	_, err := author.OnboardThingByOwner(req3)
	return err
}
