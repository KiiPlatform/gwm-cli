package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/KiiPlatform/kii_go"
	"github.com/koron/go-dproxy"
)

func localAuth(addr GatewayAddress, app App, username string, password string) (string, error) {
	if username == "" || password == "" {
		return "", errors.New("username or password is not given")
	}
	url := fmt.Sprintf("http://%s:%d/%s/token", addr.Host, addr.Port, app.Site)
	payload := fmt.Sprintf("{\"username\":\"%s\",\"password\":\"%s\"}", username, password)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(app.ID + ":" + app.Key))
	req, _ := http.NewRequest("POST", url, strings.NewReader(payload))

	req.Header.Add("authorization", "Basic "+basicAuth)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return "", err
	}
	t, err := dproxy.New(v).M("accessToken").String()
	if err != nil {
		return "", err
	}
	return t, nil
}

func _updateVID(app App, user User, currentID string, newVID string, password string) error {
	author := kii.APIAuthor{
		App: kii.App{
			AppID:    app.ID,
			AppKey:   app.Key,
			Location: app.Site,
		},
	}
	author.Token = user.Token
	req := kii.UpdateVendorThingIDRequest{
		VendorThingID: newVID,
		Password:      password,
	}
	return author.UpdateVendorThingID(currentID, req)
}

func _replaceNode(addr GatewayAddress, app App, node Node, token string) error {
	url := fmt.Sprintf("http://%s:%d/%s/apps/%s/gateway/end-nodes/%s",
		addr.Host, addr.Port, app.Site, app.ID, node.ID)

	payload := fmt.Sprintf("{\"vendorThingID\":\"%s\"}", node.VID)

	req, err := http.NewRequest("PUT", url, strings.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Add("authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || 400 <= res.StatusCode {
		return errors.New(fmt.Sprintf("failed to replace end-node. (%d)", res.StatusCode))
	}
	return nil
}

func _restore(addr GatewayAddress, app App, token string) error {
	url := fmt.Sprintf("http://%s:%d/gateway-app/gateway/restore", addr.Host, addr.Port)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || 400 <= res.StatusCode {
		return errors.New(fmt.Sprintf("failed to resotore. (%d)", res.StatusCode))
	}
	return nil
}

func _mapNode(addr GatewayAddress, app App, node Node, token string) error {
	url := fmt.Sprintf("http://%s:%d/%s/apps/%s/gateway/end-nodes/VENDOR_THING_ID:%s",
		addr.Host, addr.Port, app.Site, app.ID, node.VID)

	payload := fmt.Sprintf("{\"thingID\":\"%s\"}", node.ID)

	req, err := http.NewRequest("PUT", url, strings.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Add("authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || 400 <= res.StatusCode {
		return errors.New(fmt.Sprintf("failed to map end-node. (%d)", res.StatusCode))
	}
	return nil
}

func _postCommand(app App, user User, nodeID string, command []byte) (resp *kii.PostCommandResponse, err error) {
	author := kii.APIAuthor{
		App: kii.App{
			AppID:    app.ID,
			AppKey:   app.Key,
			Location: app.Site,
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

func _onboardNode(app App, user User, gatewayID string, nodeVID string, nodePass string) (string, error) {
	author := kii.APIAuthor{
		App: kii.App{
			AppID:    app.ID,
			AppKey:   app.Key,
			Location: app.Site,
		},
	}
	author.Token = user.Token
	req := kii.OnboardEndnodeWithGatewayThingIDRequest{
		GatewayThingID: gatewayID,
		OnboardEndnodeRequestCommon: kii.OnboardEndnodeRequestCommon{
			EndNodeVendorThingID: nodeVID,
			EndNodePassword:      nodePass,
			Owner:                "user:" + user.ID,
		},
	}
	resp, err := author.OnboardEndnodeWithGatewayThingID(req)
	if err != nil {
		return "", err
	}
	nodeID := resp.EndNodeThingID
	return nodeID, err
}

func _listPendingNodes(addr GatewayAddress, app App, token string) ([]interface{}, error) {
	url := fmt.Sprintf("http://%s:%d/%s/apps/%s/gateway/end-nodes/pending",
		addr.Host, addr.Port, app.Site, app.ID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("authorization", "Bearer "+token)
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var v interface{}
	err := json.Unmarshal(body, &v)
	if err != nil {
		fmt.Printf("parse body error:%v", err)
		return nil, err
	}
	return dproxy.New(v).Array()
}

func _userLogin(app App, username string, password string) (id string, token string, err error) {
	author := kii.APIAuthor{
		App: kii.App{
			AppID:    app.ID,
			AppKey:   app.Key,
			Location: app.Site,
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
			Location: app.Site,
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

func _onboardGateway(addr GatewayAddress, app App, token string) (string, error) {
	if token == "" {
		return "", errors.New("token is not given")
	}
	url := fmt.Sprintf("http://%s:%d/gateway-app/gateway/onboarding", addr.Host, addr.Port)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	log.Printf("resp body:%s\n", string(body))
	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return "", err
	}
	t, err := dproxy.New(v).M("thingID").String()
	if err != nil {
		return "", err
	}
	return t, nil
}
