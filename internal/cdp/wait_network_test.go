package cdp

import (
	"testing"

	"github.com/chromedp/cdproto/network"
)

func TestShouldIgnoreRequestDataURL(t *testing.T) {
	e := &network.EventRequestWillBeSent{
		Request: &network.Request{
			URL: "data:text/html,<h1>hello</h1>",
		},
	}

	if !shouldIgnoreRequest(e) {
		t.Fatal("expected data: URL to be ignored")
	}
}

func TestShouldIgnoreRequestBlobURL(t *testing.T) {
	e := &network.EventRequestWillBeSent{
		Request: &network.Request{
			URL: "blob:http://localhost/abc-123",
		},
	}

	if !shouldIgnoreRequest(e) {
		t.Fatal("expected blob: URL to be ignored")
	}
}

func TestShouldIgnoreRequestWebSocket(t *testing.T) {
	e := &network.EventRequestWillBeSent{
		Request: &network.Request{
			URL: "wss://example.com/socket",
		},
		Type: network.ResourceTypeWebSocket,
	}

	if !shouldIgnoreRequest(e) {
		t.Fatal("expected WebSocket request to be ignored")
	}
}

func TestShouldNotIgnoreNormalHTTPRequest(t *testing.T) {
	e := &network.EventRequestWillBeSent{
		Request: &network.Request{
			URL: "https://cdn.example.com/style.css",
		},
		Type: network.ResourceTypeStylesheet,
	}

	if shouldIgnoreRequest(e) {
		t.Fatal("expected normal HTTP request to NOT be ignored")
	}
}

func TestShouldIgnoreNilRequest(t *testing.T) {
	e := &network.EventRequestWillBeSent{
		Request: nil,
	}

	if !shouldIgnoreRequest(e) {
		t.Fatal("expected nil request to be ignored")
	}
}
