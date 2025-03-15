package github

import (
	"errors"
	"net/http"

	githubv3 "github.com/google/go-github/v69/github"
)

var (
	ErrNotFound             = errors.New("not found")
	errUnexpectedStatus     = errors.New("unexpected status")
	errUnsupportedEventType = errors.New("unsupported event type")
)

func IsErrAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	var errResp *githubv3.ErrorResponse
	if !errors.As(err, &errResp) {
		return false
	}
	const alreadyExists = "already_exists"
	for _, e := range errResp.Errors {
		if e.Code == alreadyExists {
			return true
		}
	}
	return false
}

func IsErrNotFound(err error) bool {
	if err == nil {
		return false
	}
	var errResp *githubv3.ErrorResponse
	if !errors.As(err, &errResp) {
		return false
	}
	res := errResp.Response
	return res != nil && res.StatusCode == http.StatusNotFound
}
