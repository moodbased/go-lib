package libasync

import (
	"errors"
	"sync"
	"sync/atomic"
)

type GoGroup struct {
	errs  []error
	anyCh chan error

	onceInit sync.Once
	lock     sync.Mutex
	wg       sync.WaitGroup
	nonEmpty atomic.Bool
}

func (gg *GoGroup) Go(f func() error) {
	gg.wg.Add(1)
	gg.nonEmpty.Store(true)
	go func() {
		defer gg.wg.Done()
		err := f()
		if err != nil {
			gg.lock.Lock()
			gg.errs = append(gg.errs, err)
			gg.lock.Unlock()
		}
		gg.notifyAny(err)
	}()
}

// All waits for all goroutines to finish and returns a combined error of all errors encountered.
func (gg *GoGroup) All() error {
	gg.wg.Wait()
	gg.lock.Lock()
	defer gg.lock.Unlock()
	return errors.Join(gg.errs...)
}

func (gg *GoGroup) initAnyCh() {
	gg.onceInit.Do(func() {
		// the size is set to a non-zero value to avoid blocking (blocked by Any())
		// when goroutines finish before Any() is called
		gg.anyCh = make(chan error, 1)
	})
}

// notifyAny send return value of goroutines to anyCh in non-blocking way.
func (gg *GoGroup) notifyAny(err error) {
	gg.initAnyCh()
	select {
	case gg.anyCh <- err:
	default:
	}
}

// Any blocks until any goroutine returns.
// if no goroutine has been added, it returns an error immediately.
func (gg *GoGroup) Any() error {
	if !gg.nonEmpty.Load() {
		return errors.New("empty GoGroup")
	}
	gg.initAnyCh()
	return <-gg.anyCh
}
