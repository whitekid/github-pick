package pocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/allegro/bigcache"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/whitekid/go-utils/log"
	"github.com/whitekid/go-utils/service"
	"github.com/whitekid/pocket-pick/pkg/config"
)

const (
	keyRequestToken = "REQUEST_TOKEN"
	keyAccessToken  = "ACCESS_TOKEN"
)

// New return pocket-pick service object
// implements service interface
func New() service.Interface {
	rootURL := config.RootURL()
	if rootURL == "" {
		panic("ROOT_URL required")
	}

	config := bigcache.DefaultConfig(config.CacheEvictionTimeout())
	config.CleanWindow = time.Minute

	cache, _ := bigcache.NewBigCache(config)

	return &pocketService{
		cache:   cache,
		rootURL: rootURL,
	}
}

type pocketService struct {
	rootURL string
	cache   *bigcache.BigCache // for api cache
}

// Serve serve the main service
func (s *pocketService) Serve(ctx context.Context, args ...string) error {
	e := s.setupRoute()

	return e.Start(config.BindAddr())
}

func (s *pocketService) setupRoute() *echo.Echo {
	e := echo.New()

	loggerConfig := middleware.DefaultLoggerConfig
	e.Use(middleware.LoggerWithConfig(loggerConfig))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))),
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				sess, _ := session.Get("pocket-pick-session", c)
				sess.Options = &sessions.Options{
					Path:     "/",
					MaxAge:   86400,
					HttpOnly: true,
				}

				c.Set("session", sess)
				return next(c)
			}
		})

	e.GET("/", s.handleGetIndex)
	e.GET("/auth", s.handleGetAuth)
	e.GET("/article/:item_id", s.handleGetArticle) // TODO 원래는 DELETE로 해야하는데, 귀찮아서..
	e.GET("/sessions", s.handleGetSession)

	return e
}

func (s *pocketService) session(c echo.Context) *sessions.Session {
	return c.Get("session").(*sessions.Session)
}

func (s *pocketService) handleGetSession(c echo.Context) error {
	sess := s.session(c)
	if sess.Values["foo"] == nil {
		sess.Values["foo"] = "0"
	} else {
		v, err := strconv.Atoi(sess.Values["foo"].(string))
		if err != nil {
			v = 0
		}
		sess.Values["foo"] = strconv.FormatInt(int64(v)+1, 10)
	}
	sess.Save(c.Request(), c.Response())

	if err := c.JSON(http.StatusOK, sess.Values["foo"]); err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (s *pocketService) handleGetIndex(c echo.Context) error {
	sess := s.session(c)

	// if not token, try to authorize
	if _, exists := sess.Values[keyRequestToken]; !exists {
		requestToken, authorizedURL, err := NewGetPocketAPI(config.ConsumerKey(), "").AuthorizedURL(fmt.Sprintf("%s/auth", s.rootURL))
		if err != nil {
			return errors.Wrapf(err, "authorize failed")
		}

		sess.Values[keyRequestToken] = requestToken
		log.Infof("save requestToken to session: %s", requestToken)
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, authorizedURL)
	}

	if _, exists := sess.Values[keyAccessToken]; !exists {
		delete(sess.Values, keyRequestToken)
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	accessToken := sess.Values[keyAccessToken].(string)
	log.Debugf("accessToken acquired, get random favorite pick: %s", accessToken)

	key := fmt.Sprintf("%s/favorites", accessToken)
	api := NewGetPocketAPI(config.ConsumerKey(), accessToken)

	data, err := s.cache.Get(key)
	var articleList map[string]Article
	if err != nil {
		if err != bigcache.ErrEntryNotFound {
			return errors.Wrapf(err, "get cache failed: %s", key)
		}

		articleList, err = api.Articles.Get(GetOpts{Favorite: Favorited})
		if err != nil {
			return errors.Wrap(err, "get favorite artcles failed")
		}
		log.Debugf("you have %d articles", len(articleList))

		// write to cache
		buf, err := json.Marshal(articleList)
		if err != nil {
			return errors.Wrap(err, "json encode failed")
		}
		s.cache.Set(key, buf)
	} else {
		log.Debug("load articles from cache")

		articleList = make(map[string]Article)
		buf := bytes.NewBuffer(data)
		if err := json.NewDecoder(buf).Decode(&articleList); err != nil {
			return errors.Wrap(err, "json decode failed")
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

	article := articleList[selected]
	log.Debugf("article: %+v", article)

	url := fmt.Sprintf("https://app.getpocket.com/read/%s", article.ItemID)
	for _, u := range []string{"blog.naver.com"} {
		if strings.Contains(url, u) {
			url = article.ResolvedID
			break
		}
	}
	// get pocket의 article view로 보이지 않는 것들..
	// https://brunch.co.kr/@workplays/29
	// http://m.blog.naver.com/mentoru/220042812351
	//
	// x 안보이는 것은... item_id, resolved_id가 같다?
	//
	// IsArticle이 뭔 의미인지..
	// if article.IsArticle == "1" {
	// 	url = article.ResolvedURL
	// }

	// log.Infof("move to %s, resolved: %s", url, article.ResolvedURL)
	return c.Redirect(http.StatusFound, url)
}

func (s *pocketService) handleGetAuth(c echo.Context) (err error) {
	sess := s.session(c)

	if _, exists := sess.Values[keyRequestToken]; !exists {
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	requestToken := sess.Values[keyRequestToken].(string)
	if _, exists := sess.Values[keyAccessToken]; !exists {
		accessToken, _, err := NewGetPocketAPI(config.ConsumerKey(), "").NewAccessToken(requestToken)
		if err != nil {
			log.Errorf("fail to get access token: %s", err)
			return err
		}

		if accessToken == "" {
			delete(sess.Values, keyAccessToken)
			sess.Save(c.Request(), c.Response())

			return c.Redirect(http.StatusFound, s.rootURL)
		}

		log.Debugf("get accessToken %s", accessToken)
		sess.Values[keyAccessToken] = accessToken
		sess.Save(c.Request(), c.Response())
	}

	log.Debug("redirect to root to read a item")
	return c.Redirect(http.StatusFound, s.rootURL)
}

func (s *pocketService) requireAccessToken(c echo.Context, token *string) error {
	sess := s.session(c)

	if _, exists := sess.Values[keyAccessToken]; !exists {
		delete(sess.Values, keyRequestToken)
		sess.Save(c.Request(), c.Response())
		return fmt.Errorf("access token not found")
	}

	*token = sess.Values[keyAccessToken].(string)
	return nil
}

// remove given article
func (s *pocketService) handleGetArticle(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ItemID missed")
	}

	var accessToken string

	if err := s.requireAccessToken(c, &accessToken); err != nil {
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	if err := NewGetPocketAPI(config.ConsumerKey(), accessToken).Articles.Delete(itemID); err != nil {
		log.Errorf("failed: %s", err)
		return err
	}

	return nil
}
