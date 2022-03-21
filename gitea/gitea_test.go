package gitea

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

/*
-- Possibles Events on Gitea:
	"create",
	"delete",
	"fork",
	"push",
	"issues",
	"issue_assign",
	"issue_label",
	"issue_milestone",
	"issue_comment",
	"pull_request",
	"pull_request_assign",
	"pull_request_label",
	"pull_request_milestone",
	"pull_request_comment",
	"pull_request_review_approved",
	"pull_request_review_rejected",
	"pull_request_review_comment",
	"pull_request_sync",
	"repository",
	"release"
*/

var hook *Webhook

func TestMain(m *testing.M) {

	// setup
	var err error
	hook, err = New(Options.Secret("sampleToken"))
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())

	// teardown
}

func newServer(handler http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(ServerPath, handler)
	return httptest.NewServer(mux)
}

/*
func TestBadRequests(t *testing.T) {
	assert := require.New(t)
	tests := []struct {
		name    string
		event   Event
		payload io.Reader
		headers http.Header
	}{
		{
			name:    "NoEvent",
			event:   PushEvents,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{},
		},


		{
			name:    "BadNoEventHeader",
			event:   PushEvents,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{},
		},
		{
			name:    "UnsubscribedEvent",
			event:   PushEvents,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{
				"X-Gitea-Event": []string{"noneexistant_event"},
			},
		},
		{
			name:    "BadBody",
			event:   PushEvents,
			payload: bytes.NewBuffer([]byte("")),
			headers: http.Header{
				"X-Gitea-Event": []string{"Push Hook"},
				"X-Gitea-Token": []string{"sampleToken"},
			},
		},
		{
			name:    "TokenMismatch",
			event:   PushEvents,
			payload: bytes.NewBuffer([]byte("{}")),
			headers: http.Header{
				"X-Gitea-Event": []string{"Push Hook"},
				"X-Gitea-Token": []string{"badsampleToken!"},
			},
		},
	}

	for _, tt := range tests {
		tc := tt
		client := &http.Client{}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var parseError error
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				_, parseError = hook.Parse(r, tc.event)
			})
			defer server.Close()
			req, err := http.NewRequest(http.MethodPost, server.URL+ServerPath, tc.payload)
			assert.NoError(err)
			req.Header = tc.headers
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			assert.NoError(err)
			assert.Equal(http.StatusOK, resp.StatusCode)
			assert.Error(parseError)
		})
	}
}
*/

func TestWebhooks(t *testing.T) {
	assert := require.New(t)
	tests := []struct {
		name     string
		event    Event
		typ      interface{}
		filename string
		headers  http.Header
	}{
		{
			name:     "PushEvent",
			event:    PushEvents,
			typ:      PushPayload{},
			filename: "../testdata/gitea/push-event.json",
			headers: http.Header{
				"X-Gitea-Event": []string{"push"},
			},
		},
		// .. TODO
	}

	for _, tt := range tests {
		tc := tt
		client := &http.Client{}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload, err := os.Open(tc.filename)
			assert.NoError(err)
			defer func() {
				_ = payload.Close()
			}()

			var parseError error
			var results interface{}
			server := newServer(func(w http.ResponseWriter, r *http.Request) {
				results, parseError = hook.Parse(r, tc.event)
			})
			defer server.Close()
			req, err := http.NewRequest(http.MethodPost, server.URL+ServerPath, payload)
			assert.NoError(err)
			req.Header = tc.headers
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Gitea-Event", "push")
			req.Header.Set("X-Gitea-Signature", "dc0bc1ab4a1e8f93a933fbfac3df3182072a02899436b26224eeb3cf28392037") // sampleToken

			resp, err := client.Do(req)
			assert.NoError(err)
			assert.Equal(http.StatusOK, resp.StatusCode)
			assert.NoError(parseError)
			assert.Equal(reflect.TypeOf(tc.typ), reflect.TypeOf(results))
		})
	}
}
