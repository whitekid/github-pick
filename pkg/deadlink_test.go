package pocket

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/whitekid/go-utils/request"
)

func TestCheckFetchArticle(t *testing.T) {
	type args struct {
		url string
	}

	tests := [...]struct {
		name        string
		args        args
		wantErr     bool
		wantSuccess bool // true if fetch success
	}{
		// {"", args{"https://infuture.kr/1688"}, false, true},
		// {"", args{"https://m.biz.chosun.com/svc/article.html?contid=2016012201926"}, false, true},
		{"", args{"http://blog.naver.com/inno_life/162500428"}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := request.Get("https://infuture.kr/1271").
				Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36").
				Do()
			if tt.wantErr && err != nil {
				require.Fail(t, "wantErr: %s but got success", tt.wantErr)
				// require.True(t, resp.Success())
			}

			require.Equal(t, tt.wantSuccess, resp.Success(), "wantSuccess: %v but get status %d", tt.wantSuccess, resp.StatusCode)
		})
	}
}
