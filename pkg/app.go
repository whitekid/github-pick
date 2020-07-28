package pocket

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/whitekid/go-utils/logging"
)

const (
	keyRequestToken = "REQUEST_TOKEN"
	keyAccessToken  = "ACCESS_TOKEN"
)

// New implements service interface
func New() *Service {
	return &Service{}
}

// Service the main service
type Service struct {
	rootURL string
}

func getEnvDef(key, def string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return def
	}

	return value
}

// Serve serve the main service
func (s *Service) Serve(ctx context.Context, args ...string) error {
	e := s.setupRoute()

	s.rootURL = getEnvDef("ROOT_URL", "http://127.0.0.1:8000")

	return e.Start(":8000")
}

func (s *Service) setupRoute() *echo.Echo {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	e.GET("/", s.getIndex)
	e.GET("/auth", s.getAuth)
	e.GET("/article/:item_id", s.getArticle) // TODO 원래는 DELETE로 해야하는데, 귀찮아서..
	e.GET("/sessions", s.getSession)

	return e
}

func (s *Service) session(c echo.Context) *sessions.Session {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
	}

	return sess
}

func (s *Service) getSession(c echo.Context) error {
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

func (s *Service) getIndex(c echo.Context) error {
	sess := s.session(c)

	// if not token, try to authorize
	if _, exists := sess.Values[keyRequestToken]; !exists {
		requestToken, authorizedURL, err := getAuthorizedURL(s.rootURL)
		if err != nil {
			return err
		}
		sess.Values[keyRequestToken] = requestToken
		c.Logger().Infof("save requestToken to session: %s", requestToken)
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, authorizedURL)
	}

	if _, exists := sess.Values[keyAccessToken]; !exists {
		delete(sess.Values, keyRequestToken)
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	accessToken := sess.Values[keyAccessToken].(string)
	log.Infof("accessToken acquired, get random favorite pick: %s", accessToken)
	url, err := getRandomPickURL(accessToken)
	if err != nil {
		log.Errorf("error: %s", err)
		return err
	}

	log.Infof("move to %s", url)
	return c.Redirect(http.StatusFound, url)
}

func (s *Service) getAuth(c echo.Context) (err error) {
	sess := s.session(c)

	if _, exists := sess.Values[keyRequestToken]; !exists {
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	requestToken := sess.Values[keyRequestToken].(string)
	if _, exists := sess.Values[keyAccessToken]; !exists {
		accessToken, _, err := getAccessToken(requestToken)
		if err != nil {
			log.Errorf("fail to get access token: %s", err)
			return err
		}

		if accessToken == "" {
			sess.Values[keyAccessToken] = nil
			return c.Redirect(http.StatusFound, s.rootURL)
		}

		log.Infof("get accessToken %s", accessToken)
		sess.Values[keyAccessToken] = accessToken
		sess.Save(c.Request(), c.Response())
	}

	log.Infof("redirect to root to read a item")
	return c.Redirect(http.StatusFound, s.rootURL)
}

func (s *Service) requireAccessToken(c echo.Context, token *string) error {
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
func (s *Service) getArticle(c echo.Context) error {
	itemID := c.Param("item_id")
	if itemID == "" {
		return c.String(http.StatusBadRequest, "ItemID missed")
	}

	var accessToken string

	if err := s.requireAccessToken(c, &accessToken); err != nil {
		return c.Redirect(http.StatusFound, s.rootURL)
	}

	if err := deleteArticle(accessToken, itemID); err != nil {
		log.Errorf("failed: %s", err)
		return err
	}

	return nil
}
