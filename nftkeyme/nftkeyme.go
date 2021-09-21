package nftkeyme

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
	// NftkeymeClient struct to hold client
	NftkeymeClient struct {
		HttpClient http.Client
		BaseUrl    string
	}

	// Asset struct to hold returned asset data
	Asset struct {
		PolicyId        string                 `json:"policy_id"`
		AssetName       string                 `json:"asset_name"`
		Quantity        string                 `json:"quantity"`
		OnChainMetadata map[string]interface{} `json:"onchain_metadata"`
	}
)

//NewClientFromEnvironment create new nftkeyme client using env vars
func NewClientFromEnvironment() NftkeymeClient {
	httpClient := &http.Client{
		Timeout: time.Second * 300,
	}

	baseURL := os.Getenv("NFTKEYME_URL")

	client := NftkeymeClient{
		HttpClient: *httpClient,
		BaseUrl:    baseURL,
	}

	return client
}

//GetAssetsForUser gets all the assets for the provided token/user
func (client NftkeymeClient) GetAssetsForUser(token string) ([]Asset, error) {
	logrus.Info("Getting asset info")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/assets", client.BaseUrl), nil)
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.HttpClient.Do(req)
	if err != nil {
		logrus.Errorf("Error posting request", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logrus.Errorf("Error getting asset info %d", resp.StatusCode)
		return nil, fmt.Errorf("Error getting asset info %d", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	assets := make([]Asset, 0)
	err = json.Unmarshal(bytes, &assets)
	if err != nil {
		return nil, err
	}

	return assets, nil
}
