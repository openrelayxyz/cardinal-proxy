package cache

import (
  "context"
  "github.com/hashicorp/golang-lru/v2"
  "github.com/openrelayxyz/cardinal-rpc"
  "errors"
)


type PromiseCache[k comparable, v any] struct {
  cache *lru.Cache[k, *CacheItem[v]]
}

func NewPromiseCache[k comparable, v any](size int) *PromiseCache[k, v] {
  cache, _ := lru.New[k, *CacheItem[v]](size)
  return &PromiseCache[k, v]{
    cache: cache,
  }
}

func (pc *PromiseCache[k,v]) Lookup(key k) (*CacheItem[v], bool) {
  ci := &CacheItem[v]{
    ch: make(chan struct{}),
    trigger: make(chan struct{}, 1),
    value: new(v),
    err: new(*rpc.RPCError),
  }
  ci.trigger <- struct{}{}
  isNew := true
  if val, ok, _ := pc.cache.PeekOrAdd(key, ci); ok {
    ci = val
    isNew = false
  }
  return ci, isNew
}

func (pc *PromiseCache[k,v]) Purge() {
  pc.cache.Purge()
}

type CacheItem[v any] struct {
  ch chan struct{}
  trigger chan struct{}
  value *v
  err **rpc.RPCError
  set bool
}

func (c *CacheItem[v]) SetValue(value v) error {
  if c.set {
    return errors.New("cache already set")
  }
  *c.value = value
  c.set = true
  close(c.ch)
  return nil
}
func (c *CacheItem[v]) Set(value v, err *rpc.RPCError) error {
  if c.set {
    return errors.New("cache already set")
  }
  *c.value = value
  if err != nil {
    *c.err = err
  }
  c.set = true
  close(c.ch)
  return nil
}

func (c *CacheItem[v]) SetError(err *rpc.RPCError) error {
  if c.set {
    return errors.New("cache already set")
  }
  *c.err = err
  c.set = true
  close(c.ch)
  return nil
}

func (c *CacheItem[v]) Incomplete() {
  select {
  case c.trigger <- struct{}{}:
  default:
  }
}

func (c *CacheItem[v]) GetContext(ctx context.Context) (*v, **rpc.RPCError, bool) {
  select{
  case <-ctx.Done():
    err := ctx.Err()
    rpcErr := rpc.NewRPCError(-1, err.Error())
    return nil, &rpcErr, true
  case <-c.ch:
    return c.value, c.err, true
  case <-c.trigger:
    go func() {
      select {
      case <-c.ch:
      case <-ctx.Done():
        c.Incomplete()
      }
    }()
    return c.value, c.err, false
  }
}

func (c *CacheItem[v]) Get() (*v, **rpc.RPCError, bool) {
  return c.GetContext(context.Background())
}
