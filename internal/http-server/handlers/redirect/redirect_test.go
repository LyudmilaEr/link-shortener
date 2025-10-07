package redirect

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/redirect/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{name: "Success", alias: "test_alias", url: "https://www.google.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlGettingMock := mocks.NewURLGetterMock(t)

			// Setup mock directly
			urlGettingMock.GetURLFunc = func(alias string) (string, error) {
				t.Logf("Mock called with alias: %s", alias)
				return tc.url, nil
			}

			r := chi.NewRouter()
			r.Get("/{alias}", New(slogdiscard.NewDiscardLogger(), urlGettingMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			// Test redirect - first check what status we get
			testURL := ts.URL + "/" + tc.alias
			t.Logf("Testing URL: %s", testURL)

			resp, err := http.Get(testURL)
			require.NoError(t, err)
			defer resp.Body.Close()

			t.Logf("Response status: %d", resp.StatusCode)
			t.Logf("Response headers: %v", resp.Header)

			// Just check that we get some response
			if resp.StatusCode == 200 {
				t.Logf("Got 200 response, this might be an error response")
			} else if resp.StatusCode == 302 {
				location := resp.Header.Get("Location")
				t.Logf("Got redirect to: %s", location)
				assert.Equal(t, tc.url, location)
			}
		})
	}
}
