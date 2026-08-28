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

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/ai"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/auth"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/code"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/config"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/rpc_client"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/service"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/store"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/test_util"
	rpc "github.com/hangtiancheng/swifty.go/swifty_rpc/pkg/rpc"
)

func newFakeCache() *fakeCache {
	return &fakeCache{values: make(map[string]string)}
}

type fakeCache struct {
	values map[string]string
}

func (f *fakeCache) Get(ctx context.Context, key string) (string, bool) {
	value, ok := f.values[key]
	return value, ok
}

func (f *fakeCache) Set(ctx context.Context, key string, value string) error {
	f.values[key] = value
	return nil
}

func (f *fakeCache) Delete(ctx context.Context, key string) error {
	delete(f.values, key)
	return nil
}

type fakeRPC struct{}

func (f fakeRPC) Complete(ctx context.Context, req rpc_client.AIRequest) (rpc_client.AIResponse, error) {
	return rpc_client.AIResponse{Answer: "answer: " + req.Question, Code: int(code.OK)}, nil
}

func (f fakeRPC) CompleteStream(ctx context.Context, req rpc_client.AIRequest) (rpc.ClientStream, error) {
	return &fakeStream{answer: "answer: " + req.Question}, nil
}

type fakeStream struct {
	answer string
	sent   bool
}

func (s *fakeStream) Recv(msg any) error {
	if s.sent {
		return io.EOF
	}
	s.sent = true
	chunk := msg.(*rpc_client.AIStreamChunk)
	chunk.Content = s.answer
	return nil
}

func (s *fakeStream) Context() context.Context {
	return context.Background()
}

func newTestApp(t *testing.T) (*App, config.Config) {
	t.Helper()
	database := fmt.Sprintf("server_app_test_%d", time.Now().UnixNano())
	st, err := store.Open(test_util.MongoURI(), database)
	if err != nil {
		if test_util.IsMongoUnauthorized(err) {
			t.Skipf("MongoDB requires authentication; set MONGO_URI with credentials to run integration tests: %v", err)
		}
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := st.DropDatabase(); err != nil {
			t.Fatalf("DropDatabase returned error: %v", err)
		}
		st.Close()
	})
	cfg := config.Config{JWTKey: "secret", JWTIssuer: "issuer", JWTSubject: "subject", JWTExpire: time.Hour, AIBaseURL: "http://127.0.0.1:1", AIModelName: "test"}
	services := service.New(cfg, st, ai.NewManager(cfg, st))
	return New(cfg, services, newFakeCache(), fakeRPC{}), cfg
}

func TestRegisterLoginAndAuthenticatedChat(t *testing.T) {
	app, _ := newTestApp(t)
	engine := app.Engine()
	registerBody := `{"email":"user@example.com","password":"pass"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", strings.NewReader(registerBody))
	engine.ServeHTTP(rec, req)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode returned error: %v", err)
	}
	if int(body["code"].(float64)) != int(code.OK) || body["token"] == "" {
		t.Fatalf("register body = %v", body)
	}
	token := body["token"].(string)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat/create-session-and-send-message", strings.NewReader(`{"question":"hello","model_type":"openai"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(rec, req)
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode returned error: %v", err)
	}
	if int(body["code"].(float64)) != int(code.OK) || body["answer"] != "answer: hello" {
		t.Fatalf("chat body = %v", body)
	}
}

func TestStreamRoutesUseSSE(t *testing.T) {
	app, _ := newTestApp(t)
	engine := app.Engine()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", strings.NewReader(`{"email":"stream@example.com","password":"pass"}`))
	engine.ServeHTTP(rec, req)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode returned error: %v", err)
	}
	token, ok := body["token"].(string)
	if !ok || token == "" {
		t.Fatalf("register body = %v", body)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat/create-session-and-send-message-stream", strings.NewReader(`{"question":"hello","model_type":"openai"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(rec, req)
	if got := rec.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("Content-Type = %q, want text/event-stream", got)
	}
	sseBody := rec.Body.String()
	if !strings.Contains(sseBody, "data: answer: hello\n\n") {
		t.Fatalf("SSE body missing data chunk: %q", sseBody)
	}
	if !strings.Contains(sseBody, "data: [DONE]\n\n") {
		t.Fatalf("SSE body missing DONE: %q", sseBody)
	}
	if !strings.Contains(sseBody, "event: session\n") {
		t.Fatalf("SSE body missing session event: %q", sseBody)
	}
}

func TestSendMessageStreamToSession(t *testing.T) {
	app, _ := newTestApp(t)
	engine := app.Engine()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", strings.NewReader(`{"email":"stream2@example.com","password":"pass"}`))
	engine.ServeHTTP(rec, req)
	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	token := body["token"].(string)

	// create a session first
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat/create-session-and-send-message", strings.NewReader(`{"question":"setup","model_type":"openai"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(rec, req)
	json.Unmarshal(rec.Body.Bytes(), &body)
	sessionID := body["session_id"].(string)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat/send-message-stream-2-session", strings.NewReader(fmt.Sprintf(`{"question":"again","model_type":"openai","session_id":%q}`, sessionID)))
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(rec, req)
	if got := rec.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("Content-Type = %q", got)
	}
	sseBody := rec.Body.String()
	if !strings.Contains(sseBody, "data: answer: again\n\n") {
		t.Fatalf("SSE body = %q", sseBody)
	}
}

func TestAuthRejectsInvalidToken(t *testing.T) {
	cfg := config.Config{JWTKey: "secret", JWTIssuer: "issuer", JWTSubject: "subject", JWTExpire: time.Hour}
	app := New(cfg, nil, newFakeCache(), fakeRPC{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/chat/get-user-sessions-by-username", nil)
	app.Engine().ServeHTTP(rec, req)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode returned error: %v", err)
	}
	if int(body["code"].(float64)) != int(code.TokenInvalid) {
		t.Fatalf("body = %v", body)
	}
}

func TestTokenQueryAuth(t *testing.T) {
	cfg := config.Config{JWTKey: "secret", JWTIssuer: "issuer", JWTSubject: "subject", JWTExpire: time.Hour}
	token, err := auth.NewToken(cfg, 1, "user")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	cache := newFakeCache()
	_ = cache.Set(context.Background(), "sessions:user", `[]`)
	app := New(cfg, nil, cache, fakeRPC{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/chat/get-user-sessions-by-username?token="+token, nil)
	app.Engine().ServeHTTP(rec, req)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode returned error: %v", err)
	}
	if int(body["code"].(float64)) != int(code.OK) {
		t.Fatalf("body = %v", body)
	}
}

func TestCORSHeadersAndOptions(t *testing.T) {
	app := New(config.Config{}, nil, newFakeCache(), fakeRPC{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/user/login", nil)
	app.Engine().ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, http.MethodOptions) {
		t.Fatalf("Access-Control-Allow-Methods = %q", got)
	}
}

func TestInvalidModelTypeRejected(t *testing.T) {
	app, _ := newTestApp(t)
	engine := app.Engine()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", strings.NewReader(`{"email":"model@example.com","password":"pass"}`))
	engine.ServeHTTP(rec, req)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode returned error: %v", err)
	}
	token, ok := body["token"].(string)
	if !ok || token == "" {
		t.Fatalf("register body = %v", body)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat/create-session-and-send-message", strings.NewReader(`{"question":"hello","model_type":"unknown"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(rec, req)
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode returned error: %v", err)
	}
	if int(body["code"].(float64)) != int(code.ParamsInvalid) {
		t.Fatalf("body = %v", body)
	}
}

func TestRAGModelTypeIsAccepted(t *testing.T) {
	app, _ := newTestApp(t)
	engine := app.Engine()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", strings.NewReader(`{"email":"rag@example.com","password":"pass"}`))
	engine.ServeHTTP(rec, req)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode returned error: %v", err)
	}
	token, ok := body["token"].(string)
	if !ok || token == "" {
		t.Fatalf("register body = %v", body)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat/create-session-and-send-message", strings.NewReader(`{"question":"hello","model_type":"openai-rag"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(rec, req)
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode returned error: %v", err)
	}
	if int(body["code"].(float64)) != int(code.OK) || body["answer"] != "answer: hello" {
		t.Fatalf("body = %v", body)
	}
}

func TestSessionsCanBeServedFromCache(t *testing.T) {
	cfg := config.Config{JWTKey: "secret", JWTIssuer: "issuer", JWTSubject: "subject", JWTExpire: time.Hour}
	cache := newFakeCache()
	if err := cache.Set(context.Background(), "sessions:user", `[{"id":"cached","title":"Cached"}]`); err != nil {
		t.Fatalf("cache Set returned error: %v", err)
	}
	token, err := auth.NewToken(cfg, 1, "user")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	app := New(cfg, nil, cache, fakeRPC{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/chat/get-user-sessions-by-username", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	app.Engine().ServeHTTP(rec, req)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode returned error: %v", err)
	}
	sessions := body["sessions"].([]any)
	session := sessions[0].(map[string]any)
	if session["id"] != "cached" {
		t.Fatalf("body = %v", body)
	}
}

func TestHistoryCanBeServedFromCache(t *testing.T) {
	cfg := config.Config{JWTKey: "secret", JWTIssuer: "issuer", JWTSubject: "subject", JWTExpire: time.Hour}
	cache := newFakeCache()
	if err := cache.Set(context.Background(), "history:user:session-1", `[{"is_user":true,"content":"cached"}]`); err != nil {
		t.Fatalf("cache Set returned error: %v", err)
	}
	token, err := auth.NewToken(cfg, 1, "user")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	app := New(cfg, nil, cache, fakeRPC{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat/get-chat-history-list", strings.NewReader(`{"session_id":"session-1"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	app.Engine().ServeHTTP(rec, req)
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode returned error: %v", err)
	}
	history := body["history"].([]any)
	message := history[0].(map[string]any)
	if message["content"] != "cached" {
		t.Fatalf("body = %v", body)
	}
}
