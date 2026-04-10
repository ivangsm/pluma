package server

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	t.Run("first request allowed", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		rl := NewRateLimiter(ctx)

		if !rl.Allow("1.2.3.4", "/contact", 100*time.Millisecond) {
			t.Fatal("expected first request to be allowed")
		}
	})

	t.Run("second request within window denied", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		rl := NewRateLimiter(ctx)

		rl.Allow("1.2.3.4", "/contact", 100*time.Millisecond)
		if rl.Allow("1.2.3.4", "/contact", 100*time.Millisecond) {
			t.Fatal("expected second request within window to be denied")
		}
	})

	t.Run("request allowed after window passes", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		rl := NewRateLimiter(ctx)

		rl.Allow("1.2.3.4", "/contact", 50*time.Millisecond)
		time.Sleep(60 * time.Millisecond)

		if !rl.Allow("1.2.3.4", "/contact", 50*time.Millisecond) {
			t.Fatal("expected request after window to be allowed")
		}
	})

	t.Run("different IPs are independent", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		rl := NewRateLimiter(ctx)

		rl.Allow("1.2.3.4", "/contact", time.Minute)
		if !rl.Allow("5.6.7.8", "/contact", time.Minute) {
			t.Fatal("expected different IP to be allowed")
		}
	})

	t.Run("concurrent access no panic", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		rl := NewRateLimiter(ctx)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				rl.Allow("1.2.3.4", "/contact", time.Millisecond)
			}()
		}
		wg.Wait()
	})
}
