package pocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"

	"github.com/allegro/bigcache"
	"github.com/pkg/errors"
	"github.com/whitekid/go-utils/log"
	"github.com/whitekid/go-utils/request"
	"github.com/whitekid/pocket-pick/pkg/config"
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
	log.Debugf("getAccessToken with %s", requestToken)

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

// ArticlesAPI ...
type ArticlesAPI struct {
	pocket *GetPocketAPI
}

const (
	UnFavorited = 1 // only return un-favorited items
	Favorited   = 2 // only return favorited items
)

// GetOpts ...
type GetOpts struct {
	Search   string // Only return items whose title or url contain the search string
	Domain   string // Only return items from a particular domain
	Favorite int    // only return favorited items
}

// ArticleGetResponse ...
type ArticleGetResponse struct {
	Status int                 `json:"status"`
	List   *map[string]Article `json:"list"`
}

// Get Retrieving a User's Pocket Data
func (a *ArticlesAPI) Get(opts GetOpts) (map[string]Article, error) {
	params := map[string]interface{}{
		"consumer_key": a.pocket.consumerKey,
		"access_token": a.pocket.accessToken,
		"state":        "all",
		"detailType":   "simple",
	}

	if opts.Favorite != 0 {
		params["favorite"] = strconv.FormatInt(int64(opts.Favorite-1), 10)
	}

	if opts.Search != "" {
		params["search"] = opts.Search
	}

	if opts.Domain != "" {
		params["domain"] = opts.Domain
	}

	resp, err := a.pocket.sess.Post("https://getpocket.com/v3/get").
		Header("X-Accept", "application/json").JSON(params).Do()
	if err != nil {
		return nil, err
	}

	if !resp.Success() {
		return nil, fmt.Errorf("failed with status %d", resp.StatusCode)
	}

	var buffer bytes.Buffer
	var buf1 bytes.Buffer
	io.Copy(&buffer, resp.Body)
	defer resp.Body.Close()

	//
	tee := io.TeeReader(&buffer, &buf1)

	// return empty list if there is no items searched
	var emptyResponse struct {
		List []string `json:"list"`
	}
	if err := json.NewDecoder(tee).Decode(&emptyResponse); err == nil {
		return nil, err
	}

	var response ArticleGetResponse
	if err := json.NewDecoder(&buf1).Decode(&response); err != nil {
		errors.Wrap(err, "JSONDecode")
		return nil, err
	}

	return *response.List, nil
}

type articleActionParam struct {
	Action string `json:"action"`
	ItemID string `json:"item_id"`
	Time   string `json:"time,omitempty"`
}

type articleActionResults struct {
	ActionResults []bool `json:"action_results"`
	Status        int    `json:"status"`
}

func (a *ArticlesAPI) sendAction(actions []articleActionParam) (*articleActionResults, error) {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(&actions)

	log.Debugf("actions: %+v", actions)
	resp, err := a.pocket.sess.Post("https://getpocket.com/v3/send").
		Form("consumer_key", a.pocket.consumerKey).
		Form("access_token", a.pocket.accessToken).
		Form("actions", buf.String()).
		Do()
	if err != nil {
		return nil, err
	}

	if !resp.Success() {
		return nil, fmt.Errorf("failed with status %d", resp.StatusCode)
	}

	var response articleActionResults
	if err := resp.JSON(&response); err != nil {
		return nil, errors.Wrapf(err, "decode response")
	}
	log.Debugf("resp: %+v", response)

	if !response.ActionResults[0] {
		return nil, fmt.Errorf("delete failed: %v, %d", response.ActionResults[0], response.Status)
	}

	return &response, nil
}

// Delete delete article by item id
// NOTE Delete action always success ㅡㅡ;
func (a *ArticlesAPI) Delete(itemIDs ...string) error {
	log.Debugf("remove item: %s", itemIDs)

	params := make([]articleActionParam, len(itemIDs))
	for i := 0; i < len(itemIDs); i++ {
		params[i].Action = "delete"
		params[i].ItemID = itemIDs[i]
	}

	_, err := a.sendAction(params)
	if err != nil {
		return errors.Wrapf(err, "delete(%s)", itemIDs)
	}

	return nil
}

// TODO(need refactor) cache를 넘기는게 좀 그렇네..
func getRandomPickArticle(cache *bigcache.BigCache, accessToken string) (*Article, error) {
	api := NewGetPocketAPI(config.ConsumerKey(), accessToken)

	key := fmt.Sprintf("%s/favorites", accessToken)
	data, err := cache.Get(key)
	var articleList map[string]Article
	if err != nil {
		if err != bigcache.ErrEntryNotFound {
			return nil, errors.Wrapf(err, "set cache: %s", key)
		}

		articleList, err = api.Articles.Get(GetOpts{Favorite: Favorited})
		if err != nil {
			return nil, errors.Wrap(err, "getArticles")
		}
		log.Debugf("you have %d articles", len(articleList))

		// write to cache
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(articleList); err != nil {
			return nil, errors.Wrap(err, "json encode")
		}
		cache.Set(key, buf.Bytes())
	} else {
		log.Debug("load articles from cache")

		articleList = make(map[string]Article)
		buf := bytes.NewBuffer(data)
		if err := json.NewDecoder(buf).Decode(&articleList); err != nil {
			return nil, errors.Wrap(err, "json decode")
		}
	}

	// random pick from articles
	pick := rand.Intn(len(articleList))

	selected := ""
	i := 0
	for k := range articleList {
		if i == pick-1 {
			selected = k
			break
		}
		i++
	}

	v := articleList[selected]
	log.Debugf("article: %+v", v)
	return &v, nil
}
