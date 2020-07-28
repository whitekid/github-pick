package pocket

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	_ "github.com/whitekid/go-utils"
	"github.com/whitekid/go-utils/request"
)

func newTestServer() (*httptest.Server, func()) {
	s := New()
	e := s.setupRoute()

	ts := httptest.NewServer(e)
	return ts, func() { ts.Close() }
}

func TestSession(t *testing.T) {
	ts, teardown := newTestServer()
	defer teardown()

	sess := request.NewSession(nil)

	for i := 0; i < 10; i++ {
		resp, err := sess.Get("%s%s", ts.URL, "/sessions").Do()
		require.NotEqual(t, 0, len(resp.Cookies()), "cookie must be exists")
		require.NoError(t, err)
		require.True(t, resp.Success(), "status=%d", resp.StatusCode)

		var v string
		require.NoError(t, resp.JSON(&v))
		require.Equal(t, strconv.FormatInt(int64(i), 10), v, "should increase cookie foo")
	}
}

func TestIndex(t *testing.T) {
	ts, teardown := newTestServer()
	defer teardown()

	resp, err := request.Get("%s/", ts.URL).Do()
	require.NoError(t, err)

	require.Equal(t, http.StatusFound, resp.StatusCode)
}

func TestAuth(t *testing.T) {
	panic("Not Implemented")
}