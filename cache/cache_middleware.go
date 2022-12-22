package cache

import (
	"context"
	"encoding/json"
	"net/http"
	"github.com/openrelayxyz/cardinal-rpc"
	"github.com/openrelayxyz/cardinal-types"
)

type call struct {
	Method string
	Params []json.RawMessage
}

func (c *call) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}

type CacheMiddleware struct {
	cache *PromiseCache[string, interface{}]
	cacheable map[string]bool
}

func NewCacheMiddleware(size int, cacheable map[string]bool, feed types.Feed) (*CacheMiddleware, error) {
	c := NewPromiseCache[string, interface{}](size)
	ch := make(chan interface{}, 128)
	sub := feed.Subscribe(ch)
	go func() {
		for range ch {
			c.Purge()
		}
		sub.Unsubscribe()
	}()
	return &CacheMiddleware{c, cacheable}, nil
}

func (cm *CacheMiddleware) Connect(context.Context, *http.Request) {}
func (cm *CacheMiddleware) Enter(ctx *rpc.CallContext, method string, params []json.RawMessage) (interface{}, *rpc.RPCError) {
	_, ctxCacheable := ctx.Get("cacheable")
	if !cm.cacheable[method] || ctxCacheable) {
		return nil, nil
	}
	c := &call{method, params}
	ci, isNew := cm.cache.Lookup(c.String())
	if !isNew {
		// On all but the first entry for these parameters, release the semaphores
		// so we're not holding it while we wait
		ctx.ReleaseSemaphore()
	}
	if v, e, ok := ci.GetContext(ctx.Context()); ok {
		// ok == true means we have valid values to return
		return *v, *e
	}
	// ok == false means we need to calculate the value. Other requests for the
	// same value will block at GetContext() until this request
	ctx.Set("cacheMiddlewarePromise", ci)
	return nil, nil
}
func (cm *CacheMiddleware) Exit(ctx *rpc.CallContext, result interface{}, err *rpc.RPCError) (interface{}, *rpc.RPCError) {
	if v, ok := ctx.Get("cacheMiddlewarePromise"); ok {
		if ci, ok := v.(*CacheItem[interface{}]); ok {
			if ctxerr := ctx.Context().Err(); ctxerr != nil {
				ci.Incomplete()
			} else {
				ci.Set(result, err)
			}
		}
	}
	return result, err
}
