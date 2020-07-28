package pocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	log "github.com/whitekid/go-utils/logging"
	"github.com/whitekid/go-utils/request"
)

// GetPocketAPI get pocket api
// please refer https://getpocket.com/developer/docs/overview
type GetPocketAPI struct {
	consumerKey string
	accessToken string
	sess        request.Interface // common sessions

	// API interfaces
	Articles *ArticlesAPI
}

// NewGetPocketAPI create GetPocket API
func NewGetPocketAPI(consumerKey, accessToken string) *GetPocketAPI {
	api := &GetPocketAPI{
		consumerKey: consumerKey,
		accessToken: accessToken,
		sess:        request.NewSession(nil),
	}

	api.Articles = &ArticlesAPI{pocket: api}

	return api
}

// Article pocket article, simplified version, it's not full structure
type Article struct {
	ItemID      string `json:"item_id"`
	GivelURL    string `json:"given_url"`
	ResolvedURL string `json:"resolved_url"`
	IsArticle   string `json:"is_article"`
}

// AuthorizedURL get authorizedURL
func (g *GetPocketAPI) AuthorizedURL(redirectURL string) (string, string, error) {
	resp, err := g.sess.Post("https://getpocket.com/v3/oauth/request").Header("X-Accept", "application/json").JSON(
		map[string]string{
			"consumer_key": g.consumerKey,
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

// AccessToken get accessToken, username from requestToken using oauth
func (g *GetPocketAPI) AccessToken(requestToken string) (string, string, error) {
	log.Infof("getAccessToken with %s", requestToken)

	resp, err := g.sess.Post("https://getpocket.com/v3/oauth/authorize").Header("X-Accept", "application/json").
		JSON(map[string]string{
			"consumer_key": g.consumerKey,
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

type ArticlesAPI struct {
	pocket *GetPocketAPI
}

// Get Retrieving a User's Pocket Data
func (a *ArticlesAPI) Get() (map[string]Article, error) {
	resp, err := a.pocket.sess.Post("https://getpocket.com/v3/get").Header("X-Accept", "application/json").JSON(
		map[string]interface{}{
			"consumer_key": a.pocket.consumerKey,
			"access_token": a.pocket.accessToken,
			"state":        "all",
			"favorite":     1,
			"detailType":   "simple",
		},
	).Do()
	if err != nil {
		return nil, err
	}

	if !resp.Success() {
		return nil, fmt.Errorf("failed with status %d", resp.StatusCode)
	}

	var response struct {
		Status int                `json:"status"`
		List   map[string]Article `json:"list"`
	}
	if err := resp.JSON(&response); err != nil {
		return nil, err
	}

	return response.List, nil
}

// Delete delete article
func (a *ArticlesAPI) Delete(itemID string) error {
	type action struct {
		Action string  `json:"action"`
		ItemID string  `json:"item_id"`
		Time   *string `json:"time"`
	}

	actions := []action{{Action: "delete", ItemID: itemID}}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(&actions)

	log.Infof("remove item: %s", itemID)
	resp, err := a.pocket.sess.Post("https://getpocket.com/v3/send").
		Form("consumer_key", a.pocket.consumerKey).
		Form("access_token", a.pocket.accessToken).
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

func getRandomPickURL(accessToken string) (string, error) {
	api := NewGetPocketAPI(os.Getenv("CONSUMER_KEY"), accessToken)
	list, err := api.Articles.Get()
	if err != nil {
		return "", err
	}

	log.Infof("you have %d articles", len(list))
	pick := rand.Intn(len(list))

	selected := ""
	i := 0
	for k := range list {
		if i == pick-1 {
			selected = k
			break
		}
		i++
	}

	v := list[selected]
	log.Infof("article: %+v", v)
	if v.IsArticle == "1" {
		return fmt.Sprintf("https://app.getpocket.com/read/%s", v.ItemID), nil
	}
	return v.ResolvedURL, nil
}
