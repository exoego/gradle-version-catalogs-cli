package cmd

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// Responses in testdata were captured from search.maven.org.
func stubMavenSearch(t *testing.T, statuses ...int) *atomic.Int32 {
	found, err := os.ReadFile(filepath.Join("testdata", "maven-search-log4j-core.json"))
	assert.NoError(t, err)
	notFound, err := os.ReadFile(filepath.Join("testdata", "maven-search-not-found.json"))
	assert.NoError(t, err)

	calls := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(calls.Add(1))
		status := statuses[min(n, len(statuses))-1]
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("q") == "g:org.apache.logging.log4j AND a:log4j-core" {
			_, _ = w.Write(found)
		} else {
			_, _ = w.Write(notFound)
		}
	}))
	t.Cleanup(server.Close)

	origURL, origDelay := mavenSearchURL, retryBaseDelay
	mavenSearchURL, retryBaseDelay = server.URL, time.Millisecond
	t.Cleanup(func() {
		mavenSearchURL, retryBaseDelay = origURL, origDelay
	})
	return calls
}

func TestSearchMavenRetriesOnServerError(t *testing.T) {
	calls := stubMavenSearch(t, http.StatusServiceUnavailable, http.StatusTooManyRequests, http.StatusOK)
	assert.Equal(t, "2.26.0", searchMaven("org.apache.logging.log4j", "log4j-core"))
	assert.Equal(t, int32(3), calls.Load())
}

func TestSearchMavenDoesNotRetryWhenNotFound(t *testing.T) {
	calls := stubMavenSearch(t, http.StatusOK)
	assert.Equal(t, "FIXME", searchMaven("a.b", "c"))
	assert.Equal(t, int32(1), calls.Load())
}

func TestSearchMavenGivesUpAfterMaxRetries(t *testing.T) {
	calls := stubMavenSearch(t, http.StatusServiceUnavailable)
	assert.Equal(t, "FIXME", searchMaven("a.b", "c"))
	assert.Equal(t, int32(maxRetries+1), calls.Load())
}

func TestSearchMavenDoesNotRetryOnClientError(t *testing.T) {
	calls := stubMavenSearch(t, http.StatusNotFound)
	assert.Equal(t, "FIXME", searchMaven("a.b", "c"))
	assert.Equal(t, int32(1), calls.Load())
}
