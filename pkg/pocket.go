package pocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	log "github.com/whitekid/go-utils/logging"
	"github.com/whitekid/go-utils/request"
)

// returns token, authorizedURL
func getAuthorizedURL(redirectURL string) (string, string, error) {
	resp, err := request.Post("https://getpocket.com/v3/oauth/request").Header("X-Accept", "application/json").JSON(
		map[string]string{
			"consumer_key": os.Getenv("CONSUMER_KEY"),
			"redirect_uri": redirectURL,
		},
	).Do()

	if err != nil {
		return "", "", err
	}

	if !resp.Success() {
		return "", "", fmt.Errorf("Error with status: %d", resp.StatusCode)
	}

	var response struct {
		Code string `json:"code"`
	}

	if err := resp.JSON(&response); err != nil {
		return "", "", err
	}

	return response.Code, fmt.Sprintf("https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s", response.Code, redirectURL), nil
}

// return accessToken, username from requestToken
func getAccessToken(requestToken string) (string, string, error) {
	log.Infof("getAccessToken with %s", requestToken)

	resp, err := request.Post("https://getpocket.com/v3/oauth/authorize").Header("X-Accept", "application/json").
		JSON(map[string]string{
			"consumer_key": os.Getenv("CONSUMER_KEY"),
			"code":         requestToken,
		}).Do()
	if err != nil {
		return "", "", err
	}

	if !resp.Success() {
		return "", "", fmt.Errorf("Failed with status: %d", resp.StatusCode)
	}

	var response struct {
		AccessToken string `json:"access_token"`
		Username    string `json:"username"`
	}
	if err := resp.JSON(&response); err != nil {
		return "", "", err
	}

	return response.AccessToken, response.Username, nil
}

func getRandomPickURL(accessToken string) (string, error) {
	resp, err := request.Post("https://getpocket.com/v3/get").Header("X-Accept", "application/json").JSON(
		map[string]interface{}{
			"consumer_key": os.Getenv("CONSUMER_KEY"),
			"access_token": accessToken,
			"state":        "all",
			"favorite":     1,
			"detailType":   "simple",
		},
	).Do()
	if err != nil {
		return "", err
	}

	if !resp.Success() {
		return "", fmt.Errorf("failed with status %d", resp.StatusCode)
	}

	var response struct {
		Status int `json:"status"`
		List   map[string]struct {
			ItemID      string `json:"item_id"`
			GivelURL    string `json:"given_url"`
			ResolvedURL string `json:"resolved_url"`
			IsArticle   string `json:"is_article"`
		} `json:"list"`
	}
	if err := resp.JSON(&response); err != nil {
		return "", err
	}

	for _, v := range response.List {
		log.Infof("article: %+v", v)
		if v.IsArticle == "1" {
			return fmt.Sprintf("https://app.getpocket.com/read/%s", v.ItemID), nil
		}
		return v.ResolvedURL, nil
	}
	return "", nil
}

func deleteArticle(accessToken, itemID string) error {
	type action struct {
		Action string  `json:"action"`
		ItemID string  `json:"item_id"`
		Time   *string `json:"time"`
	}

	actions := []action{{Action: "delete", ItemID: itemID}}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(&actions)

	log.Infof("remove item: %s", itemID)
	resp, err := request.Post("https://getpocket.com/v3/send").
		Form("consumer_key", os.Getenv("CONSUMER_KEY")).
		Form("access_token", accessToken).
		Form("actions", buf.String()).
		Do()
	if err != nil {
		return err
	}

	if !resp.Success() {
		return fmt.Errorf("failed with status %d", resp.StatusCode)
	}

	return nil
}
