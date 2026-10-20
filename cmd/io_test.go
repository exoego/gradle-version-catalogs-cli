package cmd

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func stubMavenSearch(t *testing.T, statuses ...int) *atomic.Int32 {
	calls := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(calls.Add(1))
		status := statuses[min(n, len(statuses))-1]
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(`{"response":{"numFound":1,"docs":[{"v":"1.2.3"}]}}`))
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
	assert.Equal(t, "1.2.3", searchMaven("a.b", "c"))
	assert.Equal(t, int32(3), calls.Load())
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
