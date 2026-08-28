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

package config

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	AppName        string        `json:"app_name"`
	AppHost        string        `json:"app_host"`
	AppPort        string        `json:"app_port"`
	RPCAddr        string        `json:"rpc_addr"`
	MongoURI       string        `json:"mongo_uri"`
	MongoDatabase  string        `json:"mongo_database"`
	JWTExpireHours int           `json:"jwt_expire_hours"`
	JWTExpire      time.Duration `json:"-"`
	JWTIssuer      string        `json:"jwt_issuer"`
	JWTSubject     string        `json:"jwt_subject"`
	JWTKey         string        `json:"jwt_key"`
	RAGDocsDir     string        `json:"rag_docs_dir"`
	AIModelName    string        `json:"ai_model_name"`
	AIEmbedModel   string        `json:"ai_embed_model"`
	AIBaseURL      string        `json:"ai_base_url"`
}

func Load() Config {
	return LoadFrom("config.json")
}

func LoadFrom(path string) Config {
	cfg := defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("config file %s not found, using defaults: %v", path, err)
		return cfg
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("parse config file %s: %v", path, err)
	}
	cfg.JWTExpire = time.Duration(cfg.JWTExpireHours) * time.Hour
	cfg.log(path)
	return cfg
}

func (c Config) log(path string) {
	log.Printf("config loaded from %s", path)
	log.Printf("  app_name:       %s", c.AppName)
	log.Printf("  app_host:       %s", c.AppHost)
	log.Printf("  app_port:       %s", c.AppPort)
	log.Printf("  rpc_addr:       %s", c.RPCAddr)
	log.Printf("  mongo_uri:      %s", maskURI(c.MongoURI))
	log.Printf("  mongo_database: %s", c.MongoDatabase)
	log.Printf("  jwt_expire:     %s", c.JWTExpire)
	log.Printf("  jwt_issuer:     %s", c.JWTIssuer)
	log.Printf("  jwt_subject:    %s", c.JWTSubject)
	log.Printf("  jwt_key:        %s", mask(c.JWTKey))
	log.Printf("  rag_docs_dir:   %s", c.RAGDocsDir)
	log.Printf("  ai_model_name:  %s", c.AIModelName)
	log.Printf("  ai_embed_model: %s", c.AIEmbedModel)
	log.Printf("  ai_base_url:    %s", c.AIBaseURL)
}

func mask(s string) string {
	if len(s) <= 4 {
		return strings.Repeat("*", len(s))
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}

func maskURI(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return mask(raw)
	}
	if u.User != nil && u.User.Username() != "" {
		u.User = url.UserPassword(mask(u.User.Username()), "****")
	}
	return u.String()
}

func defaults() Config {
	return Config{
		AppName:        "swifty-chatbot",
		AppHost:        "0.0.0.0",
		AppPort:        "8088",
		RPCAddr:        "127.0.0.1:19090",
		MongoURI:       "mongodb://root:pass@localhost:27017/?authSource=admin",
		MongoDatabase:  "ai_agent",
		JWTExpireHours: 8760,
		JWTExpire:      8760 * time.Hour,
		JWTIssuer:      "swifty-chatbot",
		JWTSubject:     "swifty-chatbot",
		JWTKey:         "swifty-chatbot",
		RAGDocsDir:     "./docs",
		AIModelName:    "qwen3",
		AIEmbedModel:   "nomic-embed-text",
		AIBaseURL:      "http://localhost:11434",
	}
}
