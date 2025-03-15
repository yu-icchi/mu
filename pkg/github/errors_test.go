package github

import (
	"net/http"
	"testing"

	githubv3 "github.com/google/go-github/v69/github"
	"github.com/stretchr/testify/assert"
)

func TestIsErrAlreadyExists(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{
			name:   "error is nil",
			err:    nil,
			expect: false,
		},
		{
			name:   "unknown error",
			err:    assert.AnError,
			expect: false,
		},
		{
			name: "already exists",
			err: &githubv3.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusUnprocessableEntity,
				},
				Errors: []githubv3.Error{
					{
						Code: "already_exists",
					},
				},
			},
			expect: true,
		},
		{
			name: "invalid",
			err: &githubv3.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusUnprocessableEntity,
				},
				Errors: []githubv3.Error{
					{
						Code: "invalid",
					},
				},
			},
			expect: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, IsErrAlreadyExists(tt.err))
		})
	}
}

func TestIsErrNotFound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{
			name:   "error is nil",
			err:    nil,
			expect: false,
		},
		{
			name:   "unknown error",
			err:    assert.AnError,
			expect: false,
		},
		{
			name: "not found",
			err: &githubv3.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			},
			expect: true,
		},
		{
			name: "invalid",
			err: &githubv3.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusUnprocessableEntity,
				},
				Errors: []githubv3.Error{
					{
						Code: "invalid",
					},
				},
			},
			expect: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, IsErrNotFound(tt.err))
		})
	}
}
