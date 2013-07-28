package cache

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCacheHit(t *testing.T) {
	c := New(func(key string) ([]byte, time.Duration) {
		t.Errorf("getter should not have been called")
		return nil, 1000
	})
	val := new(item)
	val.data = []byte("asd")
	c.data["hit"] = val
	data, cached := c.Get("hit")
	if string(data) != "asd" {
		t.Errorf("wrong cache value for key 'hit'")
	}
	if !cached {
		t.Errorf("should have been cached")
	}
}

func TestCacheMiss(t *testing.T) {
	var called bool
	c := New(func(key string) ([]byte, time.Duration) {
		called = true
		return []byte("asd"), 1000
	})
	val, cached := c.Get("miss")
	if cached {
		t.Errorf("should have not been cached")
	}
	if string(val) != "asd" {
		t.Errorf("wrong value returned on cache miss")
	}
}

func TestConcurrentCacheMiss(t *testing.T) {
	var called int32
	c := New(func(key string) ([]byte, time.Duration) {
		atomic.AddInt32(&called, 1)
		time.Sleep(1 * time.Millisecond)
		return []byte("asd"), 1 << 9
	})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			c.Get("miss")
			wg.Done()
		}()
	}
	wg.Wait()
	// Shouldn't call upstream more than once unnecessarily
	if called > 1 {
		t.Errorf("upstream getter was called %d times", called)
	}
}

func TestCacheTTL(t *testing.T) {
	c := New(func(key string) ([]byte, time.Duration) {
		return []byte("old data"), 1 * time.Millisecond
	})
	val, _ := c.Get("key")
	if string(val) != "old data" {
		t.Errorf("value expired too early!")
	}
	time.Sleep(2 * time.Millisecond)
	val2 := c.data["key"]
	if val2 != nil {
		t.Errorf("value didn't expire!")
	}
}
