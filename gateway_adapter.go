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

func _onboardGateway(addr GatewayAddress, app App, token string) (string, error) {
	url := fmt.Sprintf("http://%s:%d/%s/apps/%s/gateway/onboarding", addr.Host, addr.Port, app.Site, app.ID)
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

func _onboardMasterGateway(addr GatewayAddress, app App, token string) (string, error) {
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
