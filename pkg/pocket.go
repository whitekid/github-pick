package pocket

import (
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
			ItemID string `json:"item_id"`
		} `json:"list"`
	}
	if err := resp.JSON(&response); err != nil {
		return "", err
	}

	for _, v := range response.List {
		log.Infof("get pick: %s", v.ItemID)
		return fmt.Sprintf("https://app.getpocket.com/read/%s", v.ItemID), nil
	}
	return "", nil
}
