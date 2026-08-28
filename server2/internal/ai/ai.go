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

package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/config"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/model"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/rag"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/store"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	vector_stores "github.com/tmc/langchaingo/vectorstores"
)

const (
	ModelOpenAI    = "openai"
	ModelOpenAIRAG = "openai-rag"
)

type Manager struct {
	mu          sync.RWMutex
	agents      map[string]map[string]*Agent
	store       *store.Store
	cfg         config.Config
	llm         *openai.LLM
	ragRegistry *rag.Registry
}

type Agent struct {
	SessionID   string
	ModelType   string
	Username    string
	mu          sync.Mutex
	messages    []model.Message
	llm         *openai.LLM
	ragRegistry *rag.Registry
	store       *store.Store
	cfg         config.Config
}

func NewManager(cfg config.Config, st *store.Store) *Manager {
	llm, err := openai.New(
		openai.WithModel(cfg.AIModelName),
		openai.WithBaseURL(cfg.AIBaseURL),
	)
	if err != nil {
		log.Fatalf("create openai llm: %v", err)
	}
	embedLLM, err := openai.New(
		openai.WithModel(cfg.AIEmbedModel),
		openai.WithBaseURL(cfg.AIBaseURL),
	)
	if err != nil {
		log.Printf("create embed llm failed (RAG disabled): %v", err)
	}
	var ragRegistry *rag.Registry
	if embedLLM != nil {
		ragRegistry, err = rag.NewRegistry(embedLLM)
		if err != nil {
			log.Printf("create rag registry failed (RAG disabled): %v", err)
		}
	}
	return &Manager{
		agents:      make(map[string]map[string]*Agent),
		store:       st,
		cfg:         cfg,
		llm:         llm,
		ragRegistry: ragRegistry,
	}
}

func (m *Manager) AddStoredMessage(username string, sessionID string, content string, isUser bool) {
	agent := m.GetOrCreate(username, sessionID, ModelOpenAI)
	agent.messages = append(agent.messages, model.Message{SessionID: sessionID, Username: username, Content: content, IsUser: isUser})
}

func (m *Manager) GetOrCreate(username string, sessionID string, modelType string) *Agent {
	m.mu.Lock()
	defer m.mu.Unlock()
	bySession := m.agents[username]
	if bySession == nil {
		bySession = make(map[string]*Agent)
		m.agents[username] = bySession
	}
	if agent := bySession[sessionID]; agent != nil {
		agent.ModelType = modelType
		return agent
	}
	agent := &Agent{
		SessionID:   sessionID,
		ModelType:   modelType,
		Username:    username,
		llm:         m.llm,
		ragRegistry: m.ragRegistry,
		store:       m.store,
		cfg:         m.cfg,
	}
	bySession[sessionID] = agent
	return agent
}

func (m *Manager) Get(username string, sessionID string) *Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if bySession := m.agents[username]; bySession != nil {
		return bySession[sessionID]
	}
	return nil
}

func (m *Manager) SessionIDs(username string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	bySession := m.agents[username]
	ids := make([]string, 0, len(bySession))
	for id := range bySession {
		ids = append(ids, id)
	}
	return ids
}

func (m *Manager) IndexFile(ctx context.Context, username string, path string) error {
	if m.ragRegistry == nil {
		return fmt.Errorf("rag registry not initialized")
	}
	return m.ragRegistry.IndexFile(ctx, username, path)
}

func (a *Agent) Response(ctx context.Context, username string, userMessage string) (string, error) {
	a.mu.Lock()
	a.appendMessage(username, userMessage, true)
	messages := a.buildLLMMessages(ctx, userMessage)
	a.mu.Unlock()

	resp, err := a.llm.GenerateContent(ctx, messages)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from llm")
	}
	content := resp.Choices[0].Content

	a.mu.Lock()
	a.appendMessage(username, content, false)
	a.mu.Unlock()

	if a.store != nil {
		_ = a.store.CreateMessage(ctx, a.SessionID, username, userMessage, true)
		_ = a.store.CreateMessage(ctx, a.SessionID, username, content, false)
	}
	return content, nil
}

func (a *Agent) ResponseStream(ctx context.Context, username string, question string, cb func(chunk string)) error {
	a.mu.Lock()
	a.appendMessage(username, question, true)
	messages := a.buildLLMMessages(ctx, question)
	a.mu.Unlock()

	var fullContent strings.Builder
	resp, err := a.llm.GenerateContent(ctx, messages,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			s := string(chunk)
			fullContent.WriteString(s)
			cb(s)
			return nil
		}),
	)
	if err != nil {
		return err
	}

	content := fullContent.String()
	if content == "" && resp != nil && len(resp.Choices) > 0 {
		content = resp.Choices[0].Content
	}

	a.mu.Lock()
	a.appendMessage(username, content, false)
	a.mu.Unlock()

	if a.store != nil {
		_ = a.store.CreateMessage(ctx, a.SessionID, username, question, true)
		_ = a.store.CreateMessage(ctx, a.SessionID, username, content, false)
	}
	return nil
}

func (a *Agent) Messages() []model.Message {
	a.mu.Lock()
	out := make([]model.Message, len(a.messages))
	copy(out, a.messages)
	a.mu.Unlock()
	return out
}

func (a *Agent) appendMessage(username string, content string, isUser bool) {
	a.messages = append(a.messages, model.Message{
		SessionID: a.SessionID,
		Username:  username,
		Content:   content,
		IsUser:    isUser,
	})
}

func (a *Agent) buildLLMMessages(ctx context.Context, currentQuestion string) []llms.MessageContent {
	var msgs []llms.MessageContent

	for _, msg := range a.messages {
		if msg.IsUser {
			msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, msg.Content))
		} else {
			msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeAI, msg.Content))
		}
	}

	if a.ModelType == ModelOpenAIRAG && a.ragRegistry != nil && len(msgs) > 0 {
		userStore, err := a.ragRegistry.ForUser(ctx, a.Username)
		if err != nil {
			log.Printf("load rag store for user %s: %v", a.Username, err)
			return msgs
		}
		retriever := vector_stores.ToRetriever(userStore, 5)
		docs, err := retriever.GetRelevantDocuments(ctx, currentQuestion)
		if err == nil && len(docs) > 0 {
			ragPrompt := buildRAGPrompt(currentQuestion, docs)
			msgs[len(msgs)-1] = llms.TextParts(llms.ChatMessageTypeHuman, ragPrompt)
		}
	}

	return msgs
}

func buildRAGPrompt(userMessage string, docs []schema.Document) string {
	if len(docs) == 0 {
		return userMessage
	}
	var b strings.Builder
	b.WriteString("\nAnswer the user's question based on the following reference document. If the document does not contain the relevant information, please state that the information could not be found.\n\nReference Document:\n")
	for i, doc := range docs {
		fmt.Fprintf(&b, "[Document %d]: %s\n\n", i+1, doc.PageContent)
	}
	fmt.Fprintf(&b, "User Question: %s\n\nPlease provide an accurate and complete answer:", userMessage)
	return b.String()
}
