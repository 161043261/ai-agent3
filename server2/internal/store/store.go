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
	"errors"
	"fmt"
	"time"

	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/model"
	swifty_orm "github.com/hangtiancheng/swifty.go/swifty_orm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	engine *swifty_orm.Engine
}

type userDoc struct {
	ID        int64      `bson:"_id"`
	Name      string     `bson:"name"`
	Email     string     `bson:"email"`
	Username  string     `bson:"username"`
	Password  string     `bson:"password"`
	CreatedAt time.Time  `bson:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty"`
}

type sessionDoc struct {
	ID        string     `bson:"_id"`
	Username  string     `bson:"username"`
	Title     string     `bson:"title"`
	CreatedAt time.Time  `bson:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty"`
}

type messageDoc struct {
	ID        int64     `bson:"_id"`
	SessionID string    `bson:"session_id"`
	Username  string    `bson:"username"`
	Content   string    `bson:"content"`
	IsUser    bool      `bson:"is_user"`
	CreatedAt time.Time `bson:"created_at"`
}

func Open(uri string, database string) (*Store, error) {
	ctx, cancel := operationContext()
	defer cancel()
	engine, err := swifty_orm.NewEngine(ctx, uri, database)
	if err != nil {
		return nil, err
	}
	store := &Store{engine: engine}
	if err := store.ensureIndexes(); err != nil {
		_ = engine.Close(ctx)
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() {
	if s != nil && s.engine != nil {
		ctx, cancel := operationContext()
		defer cancel()
		_ = s.engine.Close(ctx)
	}
}

func (s *Store) DropDatabase() error {
	if s == nil || s.engine == nil {
		return nil
	}
	ctx, cancel := operationContext()
	defer cancel()
	return s.engine.DropDatabase(ctx)
}

func (s *Store) ensureIndexes() error {
	ctx, cancel := operationContext()
	defer cancel()
	if _, err := s.engine.Collection("users").EnsureIndexes(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uniq_users_username")},
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetName("idx_users_email")},
	}); err != nil {
		return err
	}
	if _, err := s.engine.Collection("sessions").EnsureIndexes(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "username", Value: 1}, {Key: "created_at", Value: 1}}, Options: options.Index().SetName("idx_sessions_username_created_at")},
	}); err != nil {
		return err
	}
	_, err := s.engine.Collection("messages").EnsureIndexes(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "session_id", Value: 1}, {Key: "created_at", Value: 1}}, Options: options.Index().SetName("idx_messages_session_created_at")},
	})
	return err
}

func (s *Store) InsertUser(ctx context.Context, name string, email string, username string, password string) (*model.User, error) {
	id, err := s.engine.NextSequence(ctx, "users")
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	doc := &userDoc{ID: id, Name: name, Email: email, Username: username, Password: password, CreatedAt: now, UpdatedAt: now}
	if _, err := s.engine.Collection("users").Insert(ctx, doc); err != nil {
		return nil, err
	}
	return s.GetUserByUsernameWithPassword(ctx, username, true, id)
}

func (s *Store) GetUserByUsernameWithPassword(ctx context.Context, username string, allowIDFallback bool, fallbackID int64) (*model.User, error) {
	var doc userDoc
	err := s.engine.Collection("users").
		Where("username", username).
		WhereNull("deleted_at").
		First(ctx, &doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		if allowIDFallback && fallbackID > 0 {
			return &model.User{ID: fallbackID, Name: username, Email: username, Username: username}, nil
		}
		return nil, err
	}
	return doc.toModel(), nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var doc userDoc
	err := s.engine.Collection("users").
		Where("email", email).
		WhereNull("deleted_at").
		First(ctx, &doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return doc.toModel(), nil
}

func (s *Store) CreateSession(ctx context.Context, id string, username string, title string) error {
	now := time.Now().UTC()
	doc := &sessionDoc{ID: id, Username: username, Title: title, CreatedAt: now, UpdatedAt: now}
	_, err := s.engine.Collection("sessions").Insert(ctx, doc)
	return err
}

func (s *Store) GetSessionsByUsername(ctx context.Context, username string) ([]model.SessionDTO, error) {
	var docs []sessionDoc
	err := s.engine.Collection("sessions").
		Where("username", username).
		WhereNull("deleted_at").
		OrderBy("created_at").
		OrderBy("_id").
		Find(ctx, &docs)
	if err != nil {
		return nil, err
	}
	out := make([]model.SessionDTO, 0, len(docs))
	for _, doc := range docs {
		out = append(out, model.SessionDTO{ID: doc.ID, Title: doc.Title})
	}
	return out, nil
}

func (s *Store) CreateMessage(ctx context.Context, sessionID string, username string, content string, isUser bool) error {
	id, err := s.engine.NextSequence(ctx, "messages")
	if err != nil {
		return err
	}
	doc := &messageDoc{ID: id, SessionID: sessionID, Username: username, Content: content, IsUser: isUser, CreatedAt: time.Now().UTC()}
	_, err = s.engine.Collection("messages").Insert(ctx, doc)
	return err
}

func (s *Store) GetMessagesBySessionID(ctx context.Context, sessionID string) ([]model.Message, error) {
	return s.getMessages(ctx, sessionID)
}

func (s *Store) GetAllMessages(ctx context.Context) ([]model.Message, error) {
	return s.getMessages(ctx, "")
}

func (s *Store) getMessages(ctx context.Context, sessionID string) ([]model.Message, error) {
	q := s.engine.Collection("messages")
	if sessionID != "" {
		q = q.Where("session_id", sessionID)
	}
	var docs []messageDoc
	err := q.OrderBy("created_at").OrderBy("_id").Find(ctx, &docs)
	if err != nil {
		return nil, err
	}
	out := make([]model.Message, 0, len(docs))
	for _, doc := range docs {
		out = append(out, doc.toModel())
	}
	return out, nil
}

func (s *Store) LoadMessagesInto(ctx context.Context, manager MessageLoader) error {
	messages, err := s.GetAllMessages(ctx)
	if err != nil {
		return err
	}
	for _, msg := range messages {
		manager.AddStoredMessage(msg.Username, msg.SessionID, msg.Content, msg.IsUser)
	}
	return nil
}

type MessageLoader interface {
	AddStoredMessage(username string, sessionID string, content string, isUser bool)
}

func NowTitle(value string) string {
	if value != "" {
		return value
	}
	return fmt.Sprintf("Session %s", time.Now().Format(time.RFC3339))
}

func operationContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func (d userDoc) toModel() *model.User {
	return &model.User{
		ID:        d.ID,
		Name:      d.Name,
		Email:     d.Email,
		Username:  d.Username,
		Password:  d.Password,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		DeletedAt: d.DeletedAt,
	}
}

func (d messageDoc) toModel() model.Message {
	return model.Message{
		ID:        d.ID,
		SessionID: d.SessionID,
		Username:  d.Username,
		Content:   d.Content,
		IsUser:    d.IsUser,
		CreatedAt: d.CreatedAt,
	}
}
