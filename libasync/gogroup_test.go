package libasync

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestGoGroup_All(t *testing.T) {
	var gg GoGroup

	for i := range 5 {
		gg.Go(func() error {
			dur := time.Duration(i) * time.Second
			time.Sleep(dur)
			t.Logf("task %d, process duration %s", i, dur)
			if rand.Int()%2 == 0 {
				return nil
			} else {
				return fmt.Errorf("task %d failed", i)
			}
		})
	}

	err := gg.All()
	if err != nil {
		t.Log(err)
	}
}

func TestGoGroup_Any(t *testing.T) {
	var gg GoGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := range 5 {
		gg.Go(func() (err error) {
			defer func() {
				if err != nil {
					err = fmt.Errorf("task %d failed: %w", i, err)
				}
			}()
			dur := time.Duration(rand.Intn(10)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(dur):
				t.Logf("task %d, process duration %s", i, dur)
			}
			return nil
		})
	}

	err := gg.Any()
	if err != nil {
		t.Log(err)
	}
	// when any task done, cancel other tasks
	cancel()
	err = gg.All()
	if err != nil {
		t.Log(err)
	}
}
