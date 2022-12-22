package cache

import (
  "context"
  "testing"
  "sync"
  // "log"
)

func TestHappyPath(t *testing.T) {
  pc := NewPromiseCache[string, int](12)
  var wg sync.WaitGroup
  q := 0
  for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(wg *sync.WaitGroup, i int) {
      p, _ := pc.Lookup("x")
      v, err, ok := p.Get()
      if !ok {
        q++
        if q > 1 {
          t.Errorf("Unexpected q value %v", q)
        }
        p.Set(17, nil)
      }
      if *v != 17 {
        t.Errorf("Unexpected v value %v", v)
      }
      if *err != nil {
        t.Errorf("Unexpected error %v", err)
      }
      wg.Done()
    }(&wg, i)
  }
  wg.Wait()
}
func TestIncomplete(t *testing.T) {
  pc := NewPromiseCache[string, int](12)
  var wg sync.WaitGroup
  q := 0
  for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(wg *sync.WaitGroup, i int) {
      defer wg.Done()
      ctx, cancel := context.WithCancel(context.Background())
      defer cancel()
      p, _ := pc.Lookup("x")
      v, err, ok := p.GetContext(ctx)
      if !ok {
        q++
        if q == 1 {
          return
        }
        if q > 2 {
          t.Errorf("Unexpected q value %v", q)
        }
        p.Set(17, nil)
      }
      if *v != 17 {
        t.Errorf("Unexpected v value %v", v)
      }
      if *err != nil {
        t.Errorf("Unexpected error %v", err)
      }
    }(&wg, i)
  }
  wg.Wait()
}
