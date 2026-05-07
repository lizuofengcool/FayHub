package distlock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"fayhub/pkg/redisclient"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrLockFailed    = errors.New("获取分布式锁失败")
	ErrLockNotHeld   = errors.New("未持有该锁")
	ErrLockExpired   = errors.New("锁已过期")
)

type DistributedLock struct {
	key        string
	token      string
	ttl        time.Duration
	rdb        *redis.Client
	localMu    *sync.Mutex
	localHeld  bool
	stopRenew  chan struct{}
	mu         sync.Mutex
}

var (
	localLocks   = make(map[string]*sync.Mutex)
	localLocksMu sync.Mutex
)

func NewLock(key string, ttl time.Duration) *DistributedLock {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}

	tokenBytes := make([]byte, 16)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	localLocksMu.Lock()
	if _, ok := localLocks[key]; !ok {
		localLocks[key] = &sync.Mutex{}
	}
	localMu := localLocks[key]
	localLocksMu.Unlock()

	return &DistributedLock{
		key:     key,
		token:   token,
		ttl:     ttl,
		rdb:     redisclient.GetRawClient(),
		localMu: localMu,
	}
}

func (l *DistributedLock) Lock(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.rdb != nil {
		return l.lockRedis(ctx)
	}

	return l.lockLocal()
}

func (l *DistributedLock) lockRedis(ctx context.Context) error {
	ok, err := l.rdb.SetNX(ctx, l.key, l.token, l.ttl).Result()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrLockFailed, err)
	}

	if !ok {
		return ErrLockFailed
	}

	l.stopRenew = make(chan struct{})
	go l.watchdog()

	return nil
}

func (l *DistributedLock) lockLocal() error {
	l.localMu.Lock()
	l.localHeld = true
	return nil
}

func (l *DistributedLock) Unlock(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.rdb != nil {
		return l.unlockRedis(ctx)
	}

	return l.unlockLocal()
}

func (l *DistributedLock) unlockRedis(ctx context.Context) error {
	if l.stopRenew != nil {
		close(l.stopRenew)
		l.stopRenew = nil
	}

	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.rdb.Eval(ctx, script, []string{l.key}, l.token).Result()
	if err != nil {
		return fmt.Errorf("释放锁失败: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockNotHeld
	}

	return nil
}

func (l *DistributedLock) unlockLocal() error {
	if !l.localHeld {
		return ErrLockNotHeld
	}

	l.localMu.Unlock()
	l.localHeld = false
	return nil
}

func (l *DistributedLock) watchdog() {
	renewInterval := l.ttl / 3
	ticker := time.NewTicker(renewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopRenew:
			return
		case <-ticker.C:
			ctx := context.Background()
			script := `
				if redis.call("GET", KEYS[1]) == ARGV[1] then
					return redis.call("PEXPIRE", KEYS[1], ARGV[2])
				else
					return 0
				end
			`

			result, err := l.rdb.Eval(ctx, script, []string{l.key}, l.token, l.ttl.Milliseconds()).Result()
			if err != nil || result.(int64) == 0 {
				return
			}
		}
	}
}

func (l *DistributedLock) TryLock(ctx context.Context, waitTime time.Duration) error {
	deadline := time.Now().Add(waitTime)

	for {
		err := l.Lock(ctx)
		if err == nil {
			return nil
		}

		if !errors.Is(err, ErrLockFailed) {
			return err
		}

		if time.Now().After(deadline) {
			return ErrLockFailed
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	lock := NewLock(key, ttl)

	if err := lock.Lock(ctx); err != nil {
		return err
	}

	defer func() {
		if unlockErr := lock.Unlock(ctx); unlockErr != nil {
			fmt.Printf("释放锁失败 key=%s: %v\n", key, unlockErr)
		}
	}()

	return fn()
}

func WithLockResult[T any](ctx context.Context, key string, ttl time.Duration, fn func() (T, error)) (T, error) {
	var result T
	lock := NewLock(key, ttl)

	if err := lock.Lock(ctx); err != nil {
		return result, err
	}

	defer func() {
		if unlockErr := lock.Unlock(ctx); unlockErr != nil {
			fmt.Printf("释放锁失败 key=%s: %v\n", key, unlockErr)
		}
	}()

	return fn()
}
