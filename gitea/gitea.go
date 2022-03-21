package gitea

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// Server Path
const (
	ServerPath = "/gitea/webhooks"
)

// parse errors
var (
	ErrEventNotSpecifiedToParse    = errors.New("no Event specified to parse")
	ErrInvalidHTTPMethod           = errors.New("invalid HTTP Method")
	ErrMissingGiteaEventHeader     = errors.New("missing X-Gitea-Event Header")
	ErrMissingGiteaSignatureHeader = errors.New("missing X-Gitea-Signature Header")
	ErrHMACVerificationFailed      = errors.New("X-Gitea-Signature is invalid")
	//ErrGiteaTokenVerificationFailed = errors.New("X-Gitea-Token validation failed")
	ErrEventNotFound        = errors.New("event not defined to be parsed")
	ErrParsingPayload       = errors.New("error parsing payload")
	ErrParsingSystemPayload = errors.New("error parsing system payload")
)

/*
 *** Every struct got from:
 - https://pkg.go.dev/code.gitea.io/sdk/gitea
 - https://github.com/go-gitea/gitea/blob/main/modules/structs/hook.go
*/

// Gitea hook types

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

const (
	CreateEvents                    Event = "create"
	DeleteEvents                    Event = "delete"
	ForkEvents                      Event = "fork"
	PushEvents                      Event = "push"
	IssuesEvents                    Event = "issues"
	IssueAssignEvents               Event = "issue_assign"
	IssueLabelEvents                Event = "issue_label"
	IssueMilestoneEvents            Event = "issue_milestone"
	IssueCommentEvents              Event = "issue_comment"
	PullRequestEvents               Event = "pull_request"
	PullRequestAssignEvents         Event = "pull_request_assign"
	PullRequestLabelEvents          Event = "pull_request_label"
	PullRequestMilestoneEvents      Event = "pull_request_milestone"
	PullRequestCommentEvents        Event = "pull_request_comment"
	PullRequestReviewApprovedEvents Event = "pull_request_review_approved"
	PullRequestReviewRejectedEvents Event = "pull_request_review_rejected"
	PullRequestReviewCommentEvents  Event = "pull_request_review_comment"
	PullRequestSyncEvents           Event = "pull_request_sync"
	RepositoryEvents                Event = "repository"
	ReleaseEvents                   Event = "release"
)

// Option is a configuration option for the webhook
type Option func(*Webhook) error

// Options is a namespace var for configuration options
var Options = WebhookOptions{}

// WebhookOptions is a namespace for configuration option methods
type WebhookOptions struct{}

// Secret registers the Gitea secret
func (WebhookOptions) Secret(secret string) Option {
	return func(hook *Webhook) error {
		hook.secret = secret
		return nil
	}
}

// Webhook instance contains all methods needed to process events
type Webhook struct {
	secret string
}

// Event defines a Gitea hook event type by the X-Gitea-Event Header
type Event string

// New creates and returns a WebHook instance denoted by the Provider type
func New(options ...Option) (*Webhook, error) {
	hook := new(Webhook)
	for _, opt := range options {
		if err := opt(hook); err != nil {
			return nil, errors.New("Error applying Option")
		}
	}
	return hook, nil
}

// Parse verifies and parses the events specified and returns the payload object or an error
func (hook Webhook) Parse(r *http.Request, events ...Event) (interface{}, error) {
	defer func() {
		_, _ = io.Copy(ioutil.Discard, r.Body)
		_ = r.Body.Close()
	}()

	fmt.Println("--- --- Hook Parse --- ---")

	if len(events) == 0 {
		return nil, ErrEventNotSpecifiedToParse
	}
	if r.Method != http.MethodPost {
		return nil, ErrInvalidHTTPMethod
	}

	event := r.Header.Get("X-Gitea-Event")
	if len(event) == 0 {
		return nil, ErrMissingGiteaEventHeader
	}

	giteaEvent := Event(event)

	var found bool
	for _, evt := range events {
		if evt == giteaEvent {
			found = true
			break
		}
	}
	// event not defined to be parsed
	if !found {
		return nil, ErrEventNotFound
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return nil, ErrParsingPayload
	}

	// If we have a Secret set, we should check the MAC
	if len(hook.secret) > 0 {
		signature := r.Header.Get("X-Gitea-Signature")
		if len(signature) == 0 {
			return nil, ErrMissingGiteaSignatureHeader
		}

		mac := hmac.New(sha256.New, []byte(hook.secret))
		_, _ = mac.Write(payload)

		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			return nil, ErrHMACVerificationFailed
		}
	}

	fmt.Println("--- --- Hook eventParsing --- ---")

	switch giteaEvent {
	case CreateEvents:
		var pl CreatePayload
		err := json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case DeleteEvents:
		var pl DeletePayload
		err := json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case ForkEvents:
		var pl ForkPayload
		err := json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case PushEvents:
		var pl PushPayload
		err := json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case IssuesEvents:
		var pl IssuePayload
		err := json.Unmarshal([]byte(payload), &pl)
		return pl, err
		// case issue_assign   //TODO
		// case issue_label   //TODO
		// case issue_milestone   //TODO
		// case issue_comment   //TODO
	case PullRequestEvents:
		var pl PullRequestPayload
		err := json.Unmarshal([]byte(payload), &pl)
		return pl, err
		// case pull_request_assign   //TODO
		// case pull_request_label   //TODO
		// case pull_request_milestone   //TODO
		// case pull_request_comment   //TODO
		// case pull_request_review_approved   //TODO
		// case pull_request_review_rejected   //TODO
		// case pull_request_review_comment   //TODO
		// case pull_request_sync   //TODO
	case RepositoryEvents:
		var pl RepositoryPayload
		err := json.Unmarshal([]byte(payload), &pl)
		return pl, err
	case ReleaseEvents:
		var pl ReleasePayload
		err := json.Unmarshal([]byte(payload), &pl)
		return pl, err

	default:
		return nil, fmt.Errorf("unknown event %s", giteaEvent)
	}
}
