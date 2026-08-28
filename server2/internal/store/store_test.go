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

package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/test_util"
)

type fakeLoader struct {
	items []loadedMessage
}

type loadedMessage struct {
	username string
	session  string
	content  string
	isUser   bool
}

func (f *fakeLoader) AddStoredMessage(username string, sessionID string, content string, isUser bool) {
	f.items = append(f.items, loadedMessage{
		username: username,
		session:  sessionID,
		content:  content,
		isUser:   isUser,
	})
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	database := fmt.Sprintf("server_store_test_%d", time.Now().UnixNano())
	st, err := Open(test_util.MongoURI(), database)
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
	return st
}

func TestInsertAndGetUser(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	user, err := st.InsertUser(ctx, "User A", "user@example.com", "user@example.com", "hash")
	if err != nil {
		t.Fatalf("InsertUser returned error: %v", err)
	}
	if user == nil || user.ID <= 0 || user.Username != "user@example.com" {
		t.Fatalf("InsertUser returned unexpected user: %+v", user)
	}

	byName, err := st.GetUserByUsernameWithPassword(ctx, "user@example.com", false, 0)
	if err != nil {
		t.Fatalf("GetUserByUsernameWithPassword returned error: %v", err)
	}
	if byName == nil || byName.Email != "user@example.com" {
		t.Fatalf("GetUserByUsernameWithPassword returned: %+v", byName)
	}

	byEmail, err := st.GetUserByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail returned error: %v", err)
	}
	if byEmail == nil || byEmail.Username != "user@example.com" {
		t.Fatalf("GetUserByEmail returned: %+v", byEmail)
	}

	dup, err := st.InsertUser(ctx, "User B", "user2@example.com", "user@example.com", "hash")
	if err == nil || dup != nil {
		t.Fatalf("expected duplicate username error, got user=%+v err=%v", dup, err)
	}
}

func TestSessionAndMessageFlow(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	if _, err := st.InsertUser(ctx, "User A", "user@example.com", "user", "hash"); err != nil {
		t.Fatalf("InsertUser returned error: %v", err)
	}
	if err := st.CreateSession(ctx, "session-1", "user", "hello"); err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}
	sessions, err := st.GetSessionsByUsername(ctx, "user")
	if err != nil {
		t.Fatalf("GetSessionsByUsername returned error: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "session-1" {
		t.Fatalf("GetSessionsByUsername returned: %+v", sessions)
	}

	if err := st.CreateMessage(ctx, "session-1", "user", "hi", true); err != nil {
		t.Fatalf("CreateMessage returned error: %v", err)
	}
	if err := st.CreateMessage(ctx, "session-1", "assistant", "hello", false); err != nil {
		t.Fatalf("CreateMessage returned error: %v", err)
	}
	msgs, err := st.GetMessagesBySessionID(ctx, "session-1")
	if err != nil {
		t.Fatalf("GetMessagesBySessionID returned error: %v", err)
	}
	if len(msgs) != 2 || !msgs[0].IsUser || msgs[1].IsUser {
		t.Fatalf("GetMessagesBySessionID returned: %+v", msgs)
	}

	all, err := st.GetAllMessages(ctx)
	if err != nil {
		t.Fatalf("GetAllMessages returned error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("GetAllMessages length = %d", len(all))
	}
}

func TestLoadMessagesInto(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	if err := st.CreateMessage(ctx, "session-1", "user", "q1", true); err != nil {
		t.Fatalf("CreateMessage returned error: %v", err)
	}
	if err := st.CreateMessage(ctx, "session-1", "assistant", "a1", false); err != nil {
		t.Fatalf("CreateMessage returned error: %v", err)
	}
	loader := &fakeLoader{}
	if err := st.LoadMessagesInto(ctx, loader); err != nil {
		t.Fatalf("LoadMessagesInto returned error: %v", err)
	}
	if len(loader.items) != 2 {
		t.Fatalf("loaded items length = %d", len(loader.items))
	}
}
