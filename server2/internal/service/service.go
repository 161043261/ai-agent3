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

	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/ai"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/auth"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/code"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/config"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/model"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/store"
)

type Services struct {
	Config  config.Config
	Store   *store.Store
	Manager *ai.Manager
}

func New(cfg config.Config, st *store.Store, manager *ai.Manager) *Services {
	return &Services{Config: cfg, Store: st, Manager: manager}
}

func (s *Services) Login(ctx context.Context, username string, password string) (string, code.Code) {
	user, err := s.Store.GetUserByUsernameWithPassword(ctx, username, false, 0)
	if err != nil {
		return "", code.ServerError
	}
	if user == nil {
		return "", code.UserNotExist
	}
	if user.Password != auth.PasswordHash(password) {
		return "", code.PasswordError
	}
	token, err := auth.NewToken(s.Config, user.ID, user.Username)
	if err != nil {
		return "", code.ServerError
	}
	return token, code.OK
}

func (s *Services) Register(ctx context.Context, email string, password string) (string, string, code.Code) {
	existing, err := s.Store.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", code.ServerError
	}
	if existing != nil {
		return "", "", code.UserExist
	}
	username := email
	user, err := s.Store.InsertUser(ctx, username, email, username, auth.PasswordHash(password))
	if err != nil {
		return "", "", code.ServerError
	}
	token, err := auth.NewToken(s.Config, user.ID, user.Username)
	if err != nil {
		return "", "", code.ServerError
	}
	return token, username, code.OK
}

func (s *Services) Sessions(ctx context.Context, username string) []model.SessionDTO {
	sessions, err := s.Store.GetSessionsByUsername(ctx, username)
	if err != nil {
		return nil
	}
	return sessions
}

func (s *Services) CreateSession(ctx context.Context, username string, question string) (string, code.Code) {
	id, err := auth.NewID()
	if err != nil {
		return "", code.ServerError
	}
	if err := s.Store.CreateSession(ctx, id, username, store.NowTitle(question)); err != nil {
		return "", code.ServerError
	}
	return id, code.OK
}

func (s *Services) Answer(ctx context.Context, username string, sessionID string, question string, modelType string) (string, code.Code) {
	switch modelType {
	case ai.ModelOpenAI, ai.ModelOpenAIRAG:
	default:
		return "", code.ModelNotFound
	}
	agent := s.Manager.GetOrCreate(username, sessionID, modelType)
	answer, err := agent.Response(ctx, username, question)
	if err != nil {
		return "", code.ModelError
	}
	return answer, code.OK
}

func (s *Services) AnswerStream(ctx context.Context, username string, sessionID string, question string, modelType string, cb func(string)) error {
	switch modelType {
	case ai.ModelOpenAI:
	case ai.ModelOpenAIRAG:
	default:
		return fmt.Errorf("unsupported model type: %s", modelType)
	}
	agent := s.Manager.GetOrCreate(username, sessionID, modelType)
	return agent.ResponseStream(ctx, username, question, cb)
}

func (s *Services) IndexFile(ctx context.Context, username string, path string) error {
	return s.Manager.IndexFile(ctx, username, path)
}

func (s *Services) History(ctx context.Context, username string, sessionID string) ([]model.History, code.Code) {
	agent := s.Manager.Get(username, sessionID)
	if agent == nil {
		messages, err := s.Store.GetMessagesBySessionID(ctx, sessionID)
		if err != nil {
			return nil, code.ServerError
		}
		out := make([]model.History, 0, len(messages))
		for _, msg := range messages {
			out = append(out, model.History{IsUser: msg.IsUser, Content: msg.Content})
		}
		return out, code.OK
	}
	messages := agent.Messages()
	out := make([]model.History, 0, len(messages))
	for _, msg := range messages {
		out = append(out, model.History{IsUser: msg.IsUser, Content: msg.Content})
	}
	return out, code.OK
}
