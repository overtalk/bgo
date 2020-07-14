// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clrucache_test

import (
	"testing"
	"time"
)

type MyValue []byte

func (mv MyValue) Size() int {
	return cap(mv)
}

func BenchmarkGet(b *testing.B) {
	cache := newLRUCache(64*1024*1024, 10*time.Second)
	value := make(MyValue, 1000)
	cache.Set("stuff", value)
	for i := 0; i < b.N; i++ {
		val, ok := cache.Get("stuff")
		if !ok {
			panic("error")
		}
		_ = val
	}
}

func BenchmarkSet(b *testing.B) {
	cache := newLRUCache(64*1024*1024, 10*time.Second)
	value := make(MyValue, 1000)
	for i := 0; i < b.N; i++ {
		cache.Set("stuff", value)
	}
}
