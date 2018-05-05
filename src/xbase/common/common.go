/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package common

import (
	"math/rand"
	"time"
)

func RandomTimeout(min int) *time.Timer {
	max := min * 2
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	d, delta := min, (max - min)
	if delta > 0 {
		d += rand.Intn(int(delta))
	}
	return time.NewTimer(time.Duration(d) * time.Millisecond)
}

func RandomPort(min int, max int) int {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	d, delta := min, (max - min)
	if delta > 0 {
		d += rand.Intn(int(delta))
	}
	return d
}

func NormalTimeout(d int) *time.Timer {
	return time.NewTimer(time.Duration(d) * time.Millisecond)
}

func NormalTimerRelaese(t *time.Timer) {
	if t == nil {
		return
	}

	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func NormalTicker(d int) *time.Ticker {
	return time.NewTicker(time.Duration(d) * time.Millisecond)
}
