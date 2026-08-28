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
	"hash/crc32"
	"io"
	"log"
	"maps"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/ai"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/auth"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/cache"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/code"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/config"
	rpc_client "github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/rpc_client"
	"github.com/hangtiancheng/swifty.go/swifty_chatbot/internal/service"
	"github.com/hangtiancheng/swifty.go/swifty_http"
	rpc "github.com/hangtiancheng/swifty.go/swifty_rpc/pkg/rpc"
)

type RPCClient interface {
	Complete(ctx context.Context, req rpc_client.AIRequest) (rpc_client.AIResponse, error)
	CompleteStream(ctx context.Context, req rpc_client.AIRequest) (rpc.ClientStream, error)
}

type App struct {
	cfg      config.Config
	services *service.Services
	cache    cache.Cache
	rpc      RPCClient
}

func New(cfg config.Config, services *service.Services, cache cache.Cache, rpc RPCClient) *App {
	return &App{cfg: cfg, services: services, cache: cache, rpc: rpc}
}

func (a *App) Engine() *swifty_http.Application {
	app := swifty_http.Default()
	app.Use(a.corsMiddleware)
	api := app.Router("/api/v1")
	user := api.Router("/user")
	user.Post("/login", a.login)
	user.Post("/register", a.register)
	aiGroup := api.Router("/ai")
	aiGroup.Use(a.authMiddleware)
	aiGroup.Get("/chat/get-user-sessions-by-username", a.getUserSessions)
	aiGroup.Post("/chat/create-session-and-send-message", a.createSessionAndSendMessage)
	aiGroup.Post("/chat/create-session-and-send-message-stream", a.createSessionAndSendMessageStream)
	aiGroup.Post("/chat/send-message-2-session", a.sendMessageToSession)
	aiGroup.Post("/chat/send-message-stream-2-session", a.sendMessageStreamToSession)
	aiGroup.Post("/chat/get-chat-history-list", a.getChatHistoryList)
	file := api.Router("/file")
	file.Use(a.authMiddleware)
	file.Post("/upload", a.uploadFile)
	return app
}

func (a *App) corsMiddleware(ctx *swifty_http.Context, next func()) {
	ctx.Set("Access-Control-Allow-Origin", "*")
	ctx.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
	ctx.Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
	if ctx.Method == http.MethodOptions {
		ctx.Status = http.StatusNoContent
		return
	}
	next()
}

func (a *App) authMiddleware(ctx *swifty_http.Context, next func()) {
	token := auth.BearerToken(ctx.Get("Authorization"), ctx.Query("token"))
	username, ok := auth.ParseToken(a.cfg, token)
	if !ok {
		writeCode(ctx, code.TokenInvalid)
		return
	}
	ctx.State["username"] = username
	next()
}

func usernameFrom(ctx *swifty_http.Context) string {
	username, _ := ctx.State["username"].(string)
	return username
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type questionRequest struct {
	Question  string `json:"question"`
	ModelType string `json:"model_type"`
	SessionID string `json:"session_id"`
}

type sessionRequest struct {
	SessionID string `json:"session_id"`
}

func (a *App) login(ctx *swifty_http.Context, next func()) {
	var req loginRequest
	if err := ctx.BindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		writeCode(ctx, code.ParamsInvalid)
		return
	}
	token, result := a.services.Login(ctx.Request.Context(), req.Username, req.Password)
	if result != code.OK {
		writeCode(ctx, result)
		return
	}
	writeSuccess(ctx, swifty_http.H{"token": token})
}

func (a *App) register(ctx *swifty_http.Context, next func()) {
	var req registerRequest
	if err := ctx.BindJSON(&req); err != nil || req.Email == "" || req.Password == "" {
		writeCode(ctx, code.ParamsInvalid)
		return
	}
	token, username, result := a.services.Register(ctx.Request.Context(), req.Email, req.Password)
	if result != code.OK {
		writeCode(ctx, result)
		return
	}
	writeSuccess(ctx, swifty_http.H{"token": token, "username": username})
}

func (a *App) getUserSessions(ctx *swifty_http.Context, next func()) {
	username := usernameFrom(ctx)
	key := sessionsCacheKey(username)
	var sessions any
	if a.cacheGetJSON(ctx.Request.Context(), key, &sessions) {
		writeSuccess(ctx, swifty_http.H{"sessions": sessions})
		return
	}
	typedSessions := a.services.Sessions(ctx.Request.Context(), username)
	a.cacheSetJSON(ctx.Request.Context(), key, typedSessions)
	writeSuccess(ctx, swifty_http.H{"sessions": typedSessions})
}

func (a *App) createSessionAndSendMessage(ctx *swifty_http.Context, next func()) {
	var req questionRequest
	if err := ctx.BindJSON(&req); err != nil || !isValidQuestion(req) {
		writeCode(ctx, code.ParamsInvalid)
		return
	}
	username := usernameFrom(ctx)
	sessionID, result := a.services.CreateSession(ctx.Request.Context(), username, req.Question)
	if result != code.OK {
		writeCode(ctx, result)
		return
	}
	a.invalidateSessionCaches(ctx.Request.Context(), username, sessionID)
	resp, err := a.rpc.Complete(ctx.Request.Context(), rpc_client.AIRequest{Username: username, SessionID: sessionID, Question: req.Question, ModelType: req.ModelType})
	if err != nil || code.Code(resp.Code) != code.OK {
		writeCode(ctx, code.ModelError)
		return
	}
	a.cacheDelete(ctx.Request.Context(), historyCacheKey(username, sessionID))
	writeSuccess(ctx, swifty_http.H{"session_id": sessionID, "answer": resp.Answer})
}

func (a *App) createSessionAndSendMessageStream(ctx *swifty_http.Context, next func()) {
	var req questionRequest
	if err := ctx.BindJSON(&req); err != nil || !isValidQuestion(req) {
		writeCode(ctx, code.ParamsInvalid)
		return
	}
	username := usernameFrom(ctx)
	sessionID, result := a.services.CreateSession(ctx.Request.Context(), username, req.Question)
	if result != code.OK {
		writeCode(ctx, result)
		return
	}
	a.invalidateSessionCaches(ctx.Request.Context(), username, sessionID)
	sse := ctx.SSE()
	sse.JSON("", swifty_http.H{"session_id": sessionID})
	stream, err := a.rpc.CompleteStream(ctx.Request.Context(), rpc_client.AIRequest{Username: username, SessionID: sessionID, Question: req.Question, ModelType: req.ModelType})
	if err != nil {
		sse.Event("error", err.Error())
		sse.Done()
		return
	}
	for {
		var chunk rpc_client.AIStreamChunk
		if err := stream.Recv(&chunk); err != nil {
			if err != io.EOF {
				sse.Event("error", err.Error())
			}
			break
		}
		sse.Data(chunk.Content)
	}
	a.cacheDelete(ctx.Request.Context(), historyCacheKey(username, sessionID))
	sse.Done()
}

func (a *App) sendMessageToSession(ctx *swifty_http.Context, next func()) {
	var req questionRequest
	if err := ctx.BindJSON(&req); err != nil || !isValidQuestionWithSession(req) {
		writeCode(ctx, code.ParamsInvalid)
		return
	}
	username := usernameFrom(ctx)
	resp, err := a.rpc.Complete(ctx.Request.Context(), rpc_client.AIRequest{Username: username, SessionID: req.SessionID, Question: req.Question, ModelType: req.ModelType})
	if err != nil || code.Code(resp.Code) != code.OK {
		writeCode(ctx, code.ModelError)
		return
	}
	a.cacheDelete(ctx.Request.Context(), historyCacheKey(username, req.SessionID))
	writeSuccess(ctx, swifty_http.H{"answer": resp.Answer})
}

func (a *App) sendMessageStreamToSession(ctx *swifty_http.Context, next func()) {
	var req questionRequest
	if err := ctx.BindJSON(&req); err != nil || !isValidQuestionWithSession(req) {
		writeCode(ctx, code.ParamsInvalid)
		return
	}
	username := usernameFrom(ctx)
	sse := ctx.SSE()
	stream, err := a.rpc.CompleteStream(ctx.Request.Context(), rpc_client.AIRequest{Username: username, SessionID: req.SessionID, Question: req.Question, ModelType: req.ModelType})
	if err != nil {
		sse.Event("error", err.Error())
		sse.Done()
		return
	}
	for {
		var chunk rpc_client.AIStreamChunk
		if err := stream.Recv(&chunk); err != nil {
			if err != io.EOF {
				sse.Event("error", err.Error())
			}
			break
		}
		sse.Data(chunk.Content)
	}
	a.cacheDelete(ctx.Request.Context(), historyCacheKey(username, req.SessionID))
	sse.Done()
}

func (a *App) getChatHistoryList(ctx *swifty_http.Context, next func()) {
	var req sessionRequest
	if err := ctx.BindJSON(&req); err != nil || req.SessionID == "" {
		writeCode(ctx, code.ParamsInvalid)
		return
	}
	username := usernameFrom(ctx)
	key := historyCacheKey(username, req.SessionID)
	var history any
	if a.cacheGetJSON(ctx.Request.Context(), key, &history) {
		writeSuccess(ctx, swifty_http.H{"history": history})
		return
	}
	typedHistory, result := a.services.History(ctx.Request.Context(), username, req.SessionID)
	if result != code.OK {
		writeCode(ctx, result)
		return
	}
	a.cacheSetJSON(ctx.Request.Context(), key, typedHistory)
	writeSuccess(ctx, swifty_http.H{"history": typedHistory})
}

func (a *App) uploadFile(ctx *swifty_http.Context, next func()) {
	username := usernameFrom(ctx)
	if username == "" {
		writeCode(ctx, code.TokenInvalid)
		return
	}
	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
		writeCode(ctx, code.ParamsInvalid)
		return
	}
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		writeCode(ctx, code.ParamsInvalid)
		return
	}
	defer file.Close()
	path, err := saveUpload(username, header, file)
	if err != nil {
		writeCode(ctx, code.ServerError)
		return
	}
	if indexErr := a.services.IndexFile(ctx.Request.Context(), username, path); indexErr != nil {
		log.Printf("index file for rag failed: %v", indexErr)
	}
	writeSuccess(ctx, swifty_http.H{"filepath": path})
}

func saveUpload(username string, header *multipart.FileHeader, file multipart.File) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".md" && ext != ".txt" && ext != ".json" {
		return "", fmt.Errorf("unsupported file type")
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	userDir := filepath.Join("uploads", username)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%08x%s", crc32.ChecksumIEEE(data), ext)
	path := filepath.Join(userDir, filename)
	return path, os.WriteFile(path, data, 0o644)
}

func isValidQuestion(req questionRequest) bool {
	return req.Question != "" && isSupportedModelType(req.ModelType)
}

func isValidQuestionWithSession(req questionRequest) bool {
	return isValidQuestion(req) && req.SessionID != ""
}

func isSupportedModelType(modelType string) bool {
	return modelType == ai.ModelOpenAI || modelType == ai.ModelOpenAIRAG
}

func sessionsCacheKey(username string) string {
	return "sessions:" + username
}

func historyCacheKey(username string, sessionID string) string {
	return "history:" + username + ":" + sessionID
}

func (a *App) cacheGetJSON(ctx context.Context, key string, out any) bool {
	if a.cache == nil {
		return false
	}
	value, ok := a.cache.Get(ctx, key)
	if !ok {
		return false
	}
	return json.Unmarshal([]byte(value), out) == nil
}

func (a *App) cacheSetJSON(ctx context.Context, key string, value any) {
	if a.cache == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	_ = a.cache.Set(ctx, key, string(data))
}

func (a *App) cacheDelete(ctx context.Context, key string) {
	if a.cache != nil {
		_ = a.cache.Delete(ctx, key)
	}
}

func (a *App) invalidateSessionCaches(ctx context.Context, username string, sessionID string) {
	a.cacheDelete(ctx, sessionsCacheKey(username))
	a.cacheDelete(ctx, historyCacheKey(username, sessionID))
}

func writeCode(ctx *swifty_http.Context, result code.Code) {
	ctx.Status = http.StatusOK
	ctx.JSON(codeBody(result))
}

func writeSuccess(ctx *swifty_http.Context, extra swifty_http.H) {
	body := codeBody(code.OK)
	maps.Copy(body, extra)
	ctx.Status = http.StatusOK
	ctx.JSON(body)
}

func codeBody(result code.Code) swifty_http.H {
	return swifty_http.H{"code": result, "message": code.Message(result)}
}
