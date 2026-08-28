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

package main

import (
	"context"

	swifty_cache "github.com/hangtiancheng/swifty.go/swifty_cache"
)

type cache struct {
	group *swifty_cache.Group
}

func newCache() *cache {
	return &cache{group: swifty_cache.NewGroup("swifty_chatbot", 64<<20, swifty_cache.GetterFunc(func(ctx context.Context, key string) ([]byte, error) {
		return nil, swifty_cache.ErrKeyRequired
	}))}
}

func (c *cache) Get(ctx context.Context, key string) (string, bool) {
	view, err := c.group.Get(ctx, key)
	if err != nil {
		return "", false
	}
	return view.String(), true
}

func (c *cache) Set(ctx context.Context, key string, value string) error {
	return c.group.Set(ctx, key, []byte(value))
}

func (c *cache) Delete(ctx context.Context, key string) error {
	return c.group.Delete(ctx, key)
}
