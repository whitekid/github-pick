package pocket

import (
	"fmt"
	"time"

	"github.com/allegro/bigcache/v3"
)

type cacher interface {
	Set(key, value []byte, opts ...setOption) error
	Get(key []byte) ([]byte, bool)
	Has(key []byte) bool
}

func newBigCache() cacher {
	config := bigcache.DefaultConfig(time.Hour)
	config.CleanWindow = time.Minute
	cache, _ := bigcache.NewBigCache(config)
	return &bigCacher{
		cache: cache,
	}
}

type bigCacher struct {
	cache *bigcache.BigCache
}

type setOptions struct {
	ttl *time.Time
}

type setOption interface {
	apply(o *setOptions)
}

type funcSetOption struct {
	f func(o *setOptions)
}

func (f *funcSetOption) apply(o *setOptions) { f.f(o) }

func newFuncSetOption(f func(o *setOptions)) setOption {
	return &funcSetOption{f: f}
}

func withTTL(ttl time.Duration) setOption {
	return newFuncSetOption(func(o *setOptions) {
		t := time.Now().Add(ttl)
		o.ttl = &t
	})
}

func (b *bigCacher) Set(key, value []byte, opts ...setOption) error {
	var sOpts setOptions
	for _, o := range opts {
		o.apply(&sOpts)
	}

	if err := b.cache.Set(string(key), value); err != nil {
		return err
	}

	if sOpts.ttl != nil {
		timez := sOpts.ttl.Format(time.RFC3339)
		b.cache.Set(fmt.Sprintf("%s/expire", key), []byte(timez))
	}

	return nil
}

func (b *bigCacher) Get(key []byte) ([]byte, bool) {
	data, err := b.cache.Get(string(key))
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return nil, false
		}
		return nil, false
	}

	expireb, err := b.cache.Get(fmt.Sprintf("%s/expire", key))
	if err != nil {
		return data, true
	}

	expire, err := time.Parse(time.RFC3339, string(expireb))
	if err != nil {
		return data, true
	}

	if expire.Before(time.Now()) {
		return nil, false
	}
	return data, true
}

func (b *bigCacher) Has(key []byte) bool {
	_, err := b.cache.Get(string(key))
	return err != nil
}
