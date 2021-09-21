package discord

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type (
	//Client store client info
	Client struct {
		HTTPClient http.Client
		BaseURL    string
	}

	//UserInfo store user id and email
	UserInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
)

//NewClientFromEnvironment create new client from discord client
func NewClientFromEnvironment() Client {
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	baseURL := os.Getenv("DISCORD_URL")

	client := Client{
		HTTPClient: *httpClient,
		BaseURL:    baseURL,
	}

	return client
}

//GetUserInfo get user info for the given token
func (client Client) GetUserInfo(token string) (*UserInfo, error) {
	logrus.Info("Getting discord user info")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/@me", client.BaseURL), nil)
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logrus.Errorf("Error posting request", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logrus.Errorf("Error getting user info %d", resp.StatusCode)
		return nil, fmt.Errorf("Error getting user info %d", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{}
	err = json.Unmarshal(bytes, &userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}
