package models

import (
	"context"

	"github.com/wufe/polo/pkg/utils"
)

type contextStore struct {
	utils.RWLocker
	contexts map[string]*contextAndCancel
}

type contextAndCancel struct {
	Context context.Context
	Cancel  context.CancelFunc
}

func NewContextStore(mutexBuilder utils.MutexBuilder) *contextStore {
	return &contextStore{
		RWLocker: mutexBuilder(),
		contexts: make(map[string]*contextAndCancel),
	}
}

func (s *contextStore) Named(key string) struct {
	With func(context.Context, context.CancelFunc) struct{ Delete func() }
} {
	return struct {
		With func(context.Context, context.CancelFunc) struct{ Delete func() }
	}{
		With: func(c context.Context, cf context.CancelFunc) struct{ Delete func() } {
			del := s.Add(key, c, cf)
			return struct{ Delete func() }{
				Delete: del,
			}
		},
	}
}

func (s *contextStore) Add(key string, ctx context.Context, cancel context.CancelFunc) func() {
	s.Lock()
	defer s.Unlock()
	s.contexts[key] = &contextAndCancel{
		Context: ctx,
		Cancel:  cancel,
	}

	return func() {
		s.Del(key)
	}
}

func (s *contextStore) Del(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.contexts, key)
}

func (s *contextStore) TryGet(key string) (context.Context, context.CancelFunc, bool) {
	s.RLock()
	defer s.RUnlock()
	ctx, exists := s.contexts[key]
	if exists {
		return ctx.Context, ctx.Cancel, exists
	}
	return nil, nil, false
}
