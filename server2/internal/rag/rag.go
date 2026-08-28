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

package rag

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	document_loaders "github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	text_splitter "github.com/tmc/langchaingo/textsplitter"
	vector_stores "github.com/tmc/langchaingo/vectorstores"
)

type Store struct {
	mu      sync.RWMutex
	docs    []schema.Document
	vectors [][]float32
	emb     *embeddings.EmbedderImpl
}

var _ vector_stores.VectorStore = (*Store)(nil)

func New(llm *openai.LLM) (*Store, error) {
	emb, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("create embedder: %w", err)
	}
	return &Store{emb: emb}, nil
}

func (s *Store) AddDocuments(ctx context.Context, docs []schema.Document, _ ...vector_stores.Option) ([]string, error) {
	if len(docs) == 0 {
		return nil, nil
	}
	texts := make([]string, len(docs))
	for i, d := range docs {
		texts[i] = d.PageContent
	}
	vecs, err := s.emb.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("embed documents: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, len(docs))
	for i, d := range docs {
		ids[i] = fmt.Sprintf("doc-%d", len(s.docs))
		s.docs = append(s.docs, d)
		s.vectors = append(s.vectors, vecs[i])
	}
	return ids, nil
}

func (s *Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, _ ...vector_stores.Option) ([]schema.Document, error) {
	s.mu.RLock()
	docCount := len(s.docs)
	s.mu.RUnlock()
	if docCount == 0 {
		return nil, nil
	}
	qvec, err := s.emb.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	type scored struct {
		idx   int
		score float64
	}
	scores := make([]scored, len(s.docs))
	for i, v := range s.vectors {
		scores[i] = scored{idx: i, score: cosine(qvec, v)}
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	k := min(numDocuments, len(scores))
	results := make([]schema.Document, k)
	for i := range k {
		d := s.docs[scores[i].idx]
		d.Score = float32(scores[i].score)
		results[i] = d
	}
	return results, nil
}

func (s *Store) LoadDirectory(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read dir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".md" && ext != ".txt" && ext != ".json" {
			continue
		}
		if err := s.LoadFile(ctx, filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) LoadFile(ctx context.Context, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	loader := document_loaders.NewText(f)
	splitter := text_splitter.NewRecursiveCharacter(
		text_splitter.WithChunkSize(1000),
		text_splitter.WithChunkOverlap(200),
	)
	docs, err := loader.LoadAndSplit(ctx, splitter)
	if err != nil {
		return fmt.Errorf("load and split %s: %w", path, err)
	}
	for i := range docs {
		if docs[i].Metadata == nil {
			docs[i].Metadata = make(map[string]any)
		}
		docs[i].Metadata["source"] = path
	}
	_, err = s.AddDocuments(ctx, docs)
	return err
}

func (s *Store) DocCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.docs)
}

type Registry struct {
	mu     sync.RWMutex
	stores map[string]*Store
	emb    *embeddings.EmbedderImpl
}

func NewRegistry(llm *openai.LLM) (*Registry, error) {
	emb, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("create embedder: %w", err)
	}
	return &Registry{stores: make(map[string]*Store), emb: emb}, nil
}

func (r *Registry) ForUser(ctx context.Context, username string) (*Store, error) {
	r.mu.RLock()
	s, ok := r.stores[username]
	r.mu.RUnlock()
	if ok {
		return s, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok = r.stores[username]; ok {
		return s, nil
	}
	s = &Store{emb: r.emb}
	dir := filepath.Join("uploads", username)
	if err := s.LoadDirectory(ctx, dir); err != nil {
		return nil, err
	}
	r.stores[username] = s
	return s, nil
}

func (r *Registry) IndexFile(ctx context.Context, username string, path string) error {
	r.mu.Lock()
	s, ok := r.stores[username]
	if !ok {
		s = &Store{emb: r.emb}
		r.stores[username] = s
	}
	r.mu.Unlock()
	return s.LoadFile(ctx, path)
}

func cosine(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}
