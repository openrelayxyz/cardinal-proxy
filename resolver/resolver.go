package resolver

import (
  "encoding/json"
  "github.com/openrelayxyz/cardinal-rpc"
)

type Resolver interface{
  Resolve(cctx *rpc.CallContext, method string, params []json.RawMessage) *string
}

type MetaResolver []Resolver

func (m MetaResolver) Resolve(cctx *rpc.CallContext, method string, params []json.RawMessage) *string {
  for _, r := range m {
      if b := r.Resolve(cctx, method, params); b != nil {
        return b
      }
  }
  return nil
}

type MethodResolver map[string]string

func (m MethodResolver) Resolve(cctx *rpc.CallContext, method string, params []json.RawMessage) *string {
  if backend, ok := m[method]; ok {
    return &backend
  }
  return nil
}

type ContextResolver struct{}

func (m ContextResolver) Resolve(cctx *rpc.CallContext, method string, params []json.RawMessage) *string {
  if backend, ok := cctx.Get("targetBackend"); ok {
    if b, ok := backend.(string); ok {
      return &b
    }
  }
  return nil
}
