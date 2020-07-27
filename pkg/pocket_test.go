package pocket

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAuthorizedURL(t *testing.T) {
	token, url, err := getAuthorizedURL("http://127.0.0.1")
	require.NoError(t, err)
	require.NotEqual(t, "", token)
	require.NotEqual(t, "", url)
}

func TestAuthorize(t *testing.T) {
	// need to mock web site
}
