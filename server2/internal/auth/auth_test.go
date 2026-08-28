// Copyright (c) 2026 hangtiancheng
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package auth

import (
	"testing"
	"time"

	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/config"
)

func TestTokenRoundTripAndPasswordHash(t *testing.T) {
	cfg := config.Config{JWTKey: "secret", JWTIssuer: "issuer", JWTSubject: "subject", JWTExpire: time.Hour}
	token, err := NewToken(cfg, 1, "user@example.com")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	username, ok := ParseToken(cfg, token)
	if !ok || username != "user@example.com" {
		t.Fatalf("ParseToken = %q, %v", username, ok)
	}
	if _, ok := ParseToken(config.Config{JWTKey: "other"}, token); ok {
		t.Fatal("token should not verify with a different key")
	}
	if PasswordHash("pass") != "1a1dc91c907325c69271ddf0c944bc72" {
		t.Fatal("unexpected md5 hash")
	}
}

func TestBearerTokenAndNewID(t *testing.T) {
	if got := BearerToken("Bearer abc", "query"); got != "abc" {
		t.Fatalf("BearerToken = %q", got)
	}
	if got := BearerToken("", "query"); got != "query" {
		t.Fatalf("BearerToken fallback = %q", got)
	}
	id, err := NewID()
	if err != nil {
		t.Fatalf("NewID returned error: %v", err)
	}
	if len(id) != 36 {
		t.Fatalf("id length = %d, want 36", len(id))
	}
}
