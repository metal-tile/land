package fcfs

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	multierror "github.com/hashicorp/go-multierror"
)

type result struct {
	value interface{}
	err   error
}

// Group runs some goroutines and returns a result from first finished goroutine.
// If all goroutines return errors, Group returns also an error.
// The error includes each goroutines' error as "github.com/hashicorp/go-multierror".Error.
// Group can use zero value.
type Group struct {
	initOnce sync.Once
	ch       chan result
	ctx      context.Context
	cancel   func()
	count    int64

	mu     sync.RWMutex
	result *result
}

func (g *Group) init(ctx context.Context) {
	g.ch = make(chan result)
	g.ctx, g.cancel = context.WithCancel(ctx)
}

// WithContext creates Group with given context.
func WithContext(ctx context.Context) (*Group, context.Context) {
	var g Group
	g.initOnce.Do(func() {
		g.init(ctx)
	})
	return &g, g.ctx
}

// Go runs given funcion on a goroutine.
// The result value and error of the funcion can get Wait and Result method.
func (g *Group) Go(f func() (interface{}, error)) {
	g.initOnce.Do(func() {
		g.init(context.Background())
	})

	atomic.AddInt64(&g.count, 1)
	go func() {
		v, err := f()
		select {
		case g.ch <- result{value: v, err: err}:
		case <-g.ctx.Done():
		}
		atomic.AddInt64(&g.count, -1)
	}()
}

// Delay runs funcion on a goroutine with given delay.
func (g *Group) Delay(d time.Duration, f func() (interface{}, error)) {
	g.Go(func() (interface{}, error) {
		select {
		case <-time.After(d):
			return f()
		case <-g.ctx.Done():
		}
		return nil, nil
	})
}

// Wait waits that one goroutine returns value or all goroutines return error.
func (g *Group) Wait() (v interface{}, err error) {
	<-g.Done()
	err = g.Result(&v)
	return
}

// Done returns a channel which can wait that one goroutine returns value or all goroutines return error.
// The result and error of goroutines can get from Result method.
func (g *Group) Done() <-chan struct{} {
	g.initOnce.Do(func() {
		g.init(context.Background())
	})

	go func() {
		defer g.cancel()
		for g.waitOneGoroutine() {
			// do nothing
		}
	}()

	return g.ctx.Done()
}

// Result sets the result of first finished goroutine to given argument.
// If all goroutines return errors, Group returns also an error.
// The error includes each goroutines' error as "github.com/hashicorp/go-multierror".Error.
// v must be pointer type.
// If the result cannot be set to v, Result returns error.
func (g *Group) Result(v interface{}) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.result == nil {
		return errors.New("any goroutines have not finished or run")
	}

	if g.result.value == nil {
		return g.result.err
	}

	vp := reflect.ValueOf(v)
	if vp.Kind() != reflect.Ptr {
		return errors.New("the argument must be pointer")
	}

	elm := vp.Elem()
	if !elm.CanSet() {
		return errors.New("the argument cannot set value")
	}

	gv := reflect.ValueOf(g.result.value)
	if !gv.Type().AssignableTo(elm.Type()) {
		return errors.New("the value cannot assign to the argument")
	}

	elm.Set(gv)

	return g.result.err
}

func (g *Group) waitOneGoroutine() bool {
	if atomic.LoadInt64(&g.count) <= 0 {
		return false
	}

	r := <-g.ch

	g.mu.Lock()
	defer g.mu.Unlock()

	if r.err == nil {
		g.result = &result{value: r.value}
		return false
	}

	if g.result == nil {
		g.result = &result{}
	}
	g.result.err = multierror.Append(g.result.err, r.err)

	return true
}
