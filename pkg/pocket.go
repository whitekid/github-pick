package pocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/whitekid/go-utils/log"
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

func init() {
	rand.Seed(time.Now().UnixNano())
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

// Article pocket article, see https://getpocket.com/developer/docs/v3/retrieve
type Article struct {
	ItemID        string `json:"item_id"`
	ResolvedID    string `json:"resolved_id"`
	GivenURL      string `json:"given_url"`
	GivelTitle    string `json:"given_title"`
	Favorite      string `json:"favorite"`
	Status        string `json:"status"`
	ResolvedTitle string `json:"resolved_title"`
	ResolvedURL   string `json:"resolved_url"`
	Excerpt       string `json:"excerpt"`
	IsArticle     string `json:"is_article"`
	HasVideo      string `json:"has_video"`
	HasImage      string `json:"has_image"`
	WordCount     string `json:"word_count"`
	Images        map[string]struct {
		ItemID  string `json:"item_id"`
		ImageID string `json:"image_id"`
		Src     string `json:"src"`
		Width   string `json:"width"`
		Height  string `json:"height"`
		Credit  string `json:"credit"`
		Caption string `json:"caption"`
	} `json:"images"`
	Videos map[string]struct {
		ItemID  string `json:"item_id"`
		VideoID string `json:"video_id"`
		Src     string `json:"src"`
		Width   string `json:"width"`
		Height  string `json:"height"`
		Type    string `json:"type"`
		Vid     string `json:"vid"`
	} `json:"videos"`
}

func (g *GetPocketAPI) success(r *request.Response) error {
	if r.Success() {
		return nil
	}
	message := r.Header.Get("x-error")
	code := r.Header.Get("x-error-code")
	return fmt.Errorf("Error with status: %d, error=%s, code=%s", r.StatusCode, message, code)
}

// AuthorizedURL get authorizedURL
func (g *GetPocketAPI) AuthorizedURL(redirectURI string) (string, string, error) {
	resp, err := request.Post("https://getpocket.com/v3/oauth/request").
		Header("X-"+echo.HeaderAccept, echo.MIMEApplicationJSON).
		JSON(
			map[string]string{
				"consumer_key": g.consumerKey,
				"redirect_uri": redirectURI,
			},
		).Do()

	if err != nil {
		return "", "", err
	}

	if err := g.success(resp); err != nil {
		return "", "", errors.Wrap(err, "AutorizedURL failed")
	}

	var response struct {
		Code string `json:"code"`
	}

	if err := resp.JSON(&response); err != nil {
		return "", "", err
	}

	return response.Code, fmt.Sprintf("https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s", response.Code, redirectURI), nil
}

// NewAccessToken get accessToken, username from requestToken using oauth
func (g *GetPocketAPI) NewAccessToken(requestToken string) (string, string, error) {
	log.Debugf("getAccessToken with %s", requestToken)

	resp, err := g.sess.Post("https://getpocket.com/v3/oauth/authorize").
		Header("X-"+echo.HeaderAccept, echo.MIMEApplicationJSON).
		JSON(map[string]string{
			"consumer_key": g.consumerKey,
			"code":         requestToken,
		}).Do()
	if err != nil {
		return "", "", err
	}

	if err := g.success(resp); err != nil {
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
		Header("X-"+echo.HeaderAccept, echo.MIMEApplicationJSON).
		JSON(params).Do()
	if err != nil {
		return nil, err
	}

	if err := a.pocket.success(resp); err != nil {
		return nil, errors.Wrapf(err, "Get()")
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

	if err := a.pocket.success(resp); err != nil {
		return nil, errors.Wrap(err, "sendAction()")
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
