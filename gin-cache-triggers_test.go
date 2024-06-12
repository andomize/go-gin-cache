package gogincache

import (
	"net/http"
	"testing"
)

func TestTriggerURI(t *testing.T) {

	tURL := TriggerURI{
		Methods: []string{
			http.MethodPost,
			http.MethodPatch,
			http.MethodPut,
		},
		URI: ".*/users.*",
	}

	if !tURL.Comparable(&http.Request{
		RequestURI: "api/v1/users",
		Method:     http.MethodPost,
	}) {
		// Must Ok
		t.Fatalf("TEST 1: FAIL")
	}

	if !tURL.Comparable(&http.Request{
		RequestURI: "https://domain.domain/api/v1/users/username",
		Method:     http.MethodPatch,
	}) {
		// Must Ok
		t.Fatalf("TEST 2: FAIL")
	}

	if tURL.Comparable(&http.Request{
		RequestURI: "https://domain.domain/api/v1/users?hello=world",
		Method:     http.MethodGet,
	}) {
		// Must Not Ok
		t.Fatalf("TEST 3: FAIL")
	}

	if tURL.Comparable(&http.Request{
		RequestURI: "https://domain.domain/api/v1/users/username?hello=world",
		Method:     http.MethodGet,
	}) {
		// Must Not Ok
		t.Fatalf("TEST 4: FAIL")
	}
}
