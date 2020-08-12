package pocket

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/whitekid/pocket-pick/pkg/config"
)

func TestGetAuthorizedURL(t *testing.T) {
	api := NewGetPocketAPI("", "")

	token, url, err := api.AuthorizedURL("http://127.0.0.1")
	require.NoError(t, err)
	require.NotEqual(t, "", token)
	require.NotEqual(t, "", url)
}

func TestAuthorize(t *testing.T) {
	// need to mock web site
}

func TestArticleSearch(t *testing.T) {
	type args struct {
		url string
	}

	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{"url", args{"http://dalinaum-kr.tumblr.com/post/15516936704/git-work-flow"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := NewGetPocketAPI(config.ConsumerKey(), config.AccessToken())
			items, err := api.Articles.Get(GetOpts{
				Search: tt.args.url,
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(items))

			for _, item := range items {
				require.Equal(t, tt.args.url, item.ResolvedURL)
			}
		})
	}

}

func TestArticleDelete(t *testing.T) {
	api := NewGetPocketAPI(config.ConsumerKey(), config.AccessToken())
	require.NoError(t, api.Articles.Delete("567640688"))
}
