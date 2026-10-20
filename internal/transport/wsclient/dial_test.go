package wsclient

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDialRefusesPlainWSUnlessLoopback(t *testing.T) {
	cases := []struct {
		url      string
		insecure bool // refused by the transport-security rule
	}{
		{"ws://example.com/ws", true},
		{"ws://192.0.2.1:8080/ws", true},
		{"ws://[2001:db8::1]/ws", true},
		{"ws://localhost.example.com/ws", true},
		{"ws://127.0.0.1.example.com/ws", true},
		{"http://example.com/ws", true},
		{"ftp://example.com/ws", true},
		{"ws://127.0.0.1:1/ws", false},
		{"ws://127.9.9.9:1/ws", false},
		{"ws://[::1]:1/ws", false},
		{"ws://localhost:1/ws", false},
		{"ws://LOCALHOST:1/ws", false},
		{"wss://example.com:1/ws", false},
	}
	for _, tc := range cases {
		t.Run(tc.url, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			_, err := Dial(ctx, Options{URL: tc.url, Credential: "SECRETVALUE"})
			if err == nil {
				t.Fatal("dial unexpectedly succeeded")
			}
			if got := errors.Is(err, ErrInsecureTransport); got != tc.insecure {
				t.Fatalf("errors.Is(ErrInsecureTransport) = %v; want %v (err: %v)", got, tc.insecure, err)
			}
			if strings.Contains(err.Error(), "SECRETVALUE") {
				t.Fatalf("error leaks the credential: %v", err)
			}
		})
	}
}
