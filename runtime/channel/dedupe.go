package channel

import (
	"strings"
	"sync"
	"time"
)

type DedupeReservation struct {
	Key string
	TTL time.Duration
}

type dedupeEntry struct {
	state     string
	expiresAt time.Time
}

type Deduper struct {
	ttl time.Duration
	now func() time.Time

	mu    sync.Mutex
	items map[string]dedupeEntry
}

func NewDeduper(ttl time.Duration) *Deduper {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &Deduper{
		ttl:   ttl,
		now:   time.Now,
		items: map[string]dedupeEntry{},
	}
}

func (d *Deduper) Begin(key string) bool {
	return d.BeginFor(key, 0)
}

func (d *Deduper) BeginFor(key string, ttl time.Duration) bool {
	if d == nil || strings.TrimSpace(key) == "" {
		return true
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.pruneLocked()
	if _, ok := d.items[key]; ok {
		return false
	}
	d.items[key] = dedupeEntry{state: "pending", expiresAt: d.now().Add(d.resolveTTL(ttl))}
	return true
}

func (d *Deduper) Complete(key string) {
	d.CompleteFor(key, 0)
}

func (d *Deduper) CompleteFor(key string, ttl time.Duration) {
	if d == nil || strings.TrimSpace(key) == "" {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.items[key] = dedupeEntry{state: "done", expiresAt: d.now().Add(d.resolveTTL(ttl))}
}

func (d *Deduper) Forget(key string) {
	if d == nil || strings.TrimSpace(key) == "" {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.items, key)
}

func (d *Deduper) resolveTTL(ttl time.Duration) time.Duration {
	if ttl > 0 {
		return ttl
	}
	return d.ttl
}

func (d *Deduper) pruneLocked() {
	now := d.now()
	for key, item := range d.items {
		if now.After(item.expiresAt) {
			delete(d.items, key)
		}
	}
}

func BuildContentDedupeKey(message Message) string {
	text := strings.ToLower(strings.TrimSpace(message.Text))
	if text == "" {
		return ""
	}
	text = strings.Join(strings.Fields(text), " ")
	scope := firstNonEmpty(strings.TrimSpace(message.SessionID), strings.TrimSpace(message.ChatID))
	if scope == "" {
		return ""
	}
	return "content:" + scope + ":" + strings.TrimSpace(message.UserID) + ":" + text
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
