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

package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/ai"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/code"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/config"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/store"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/test_util"
)

func newTestServices(t *testing.T) *Services {
	t.Helper()
	database := fmt.Sprintf("server_service_test_%d", time.Now().UnixNano())
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
	return New(cfg, st, ai.NewManager(cfg, st))
}

func TestRegisterAndLogin(t *testing.T) {
	srv := newTestServices(t)
	ctx := context.Background()
	token, username, result := srv.Register(ctx, "user@example.com", "pass")
	if result != code.OK || token == "" || username != "user@example.com" {
		t.Fatalf("Register = %q, %q, %d", token, username, result)
	}
	if _, _, result := srv.Register(ctx, "user@example.com", "pass"); result != code.UserExist {
		t.Fatalf("duplicate register code = %d", result)
	}
	if token, result := srv.Login(ctx, "user@example.com", "pass"); result != code.OK || token == "" {
		t.Fatalf("Login = %q, %d", token, result)
	}
	if _, result := srv.Login(ctx, "user@example.com", "bad"); result != code.PasswordError {
		t.Fatalf("bad password code = %d", result)
	}
	if _, result := srv.Login(ctx, "missing", "pass"); result != code.UserNotExist {
		t.Fatalf("missing user code = %d", result)
	}
}

func TestSessionAndHistory(t *testing.T) {
	srv := newTestServices(t)
	ctx := context.Background()
	sessionID, result := srv.CreateSession(ctx, "user", "hello")
	if result != code.OK || sessionID == "" {
		t.Fatalf("CreateSession = %q, %d", sessionID, result)
	}
	if sessions := srv.Sessions(ctx, "user"); len(sessions) != 1 || sessions[0].ID != sessionID {
		t.Fatalf("sessions = %+v", sessions)
	}
	if err := srv.Store.CreateMessage(ctx, sessionID, "user", "hello", true); err != nil {
		t.Fatalf("CreateMessage returned error: %v", err)
	}
	history, result := srv.History(ctx, "user", sessionID)
	if result != code.OK || len(history) != 1 || !history[0].IsUser {
		t.Fatalf("history = %+v, %d", history, result)
	}
}

func TestAnswerRejectsUnsupportedModelType(t *testing.T) {
	srv := newTestServices(t)
	answer, result := srv.Answer(context.Background(), "user", "session", "hello", "unknown")
	if result != code.ModelNotFound || answer != "" {
		t.Fatalf("Answer = %q, %d", answer, result)
	}
}

func TestAnswerAcceptsRAGModelType(t *testing.T) {
	srv := newTestServices(t)
	_, result := srv.Answer(context.Background(), "user", "session", "hello", ai.ModelOpenAIRAG)
	if result == code.ModelNotFound {
		t.Fatalf("RAG model type should be accepted, got ModelNotFound")
	}
}
