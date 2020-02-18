/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package common

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandomProt(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Printf("%v\n", RandomPort(8000, 9000))
	}
}

func TestRandomTimeout(t *testing.T) {
	var fired int32
	var want int32

	tick := RandomTimeout(100)
	go func() {
		for range tick.C {
			atomic.AddInt32(&fired, 1)
		}
	}()

	time.Sleep(210 * time.Millisecond)
	want = 1
	got := atomic.LoadInt32(&fired)
	assert.Equal(t, want, got)
}

func TestNormalTimeout(t *testing.T) {
	var fired int32
	var want int32

	tick := NormalTimeout(100)
	go func() {
		for range tick.C {
			atomic.AddInt32(&fired, 1)
		}
	}()

	{
		time.Sleep(150 * time.Millisecond)
		want = 1
		got := atomic.LoadInt32(&fired)
		assert.Equal(t, want, got)
	}

	{
		NormalTimerRelaese(tick)
		tick = NormalTimeout(100)
		time.Sleep(150 * time.Millisecond)
		want = 1
		got := atomic.LoadInt32(&fired)
		assert.Equal(t, want, got)
	}
}

func TestNormalTicker(t *testing.T) {
	var fired int32
	var want int32

	tick := NormalTicker(100)
	go func() {
		for range tick.C {
			atomic.AddInt32(&fired, 1)
		}
	}()

	time.Sleep(101 * time.Millisecond)
	want = 1
	got := atomic.LoadInt32(&fired)
	assert.Equal(t, want, got)
}
