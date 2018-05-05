/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package raft

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeersJson(t *testing.T) {
	path := "/tmp/test.peersjson"
	peers := []string{":0101", ":0202"}

	{
		err := writePeersJSON(path, peers)
		assert.Nil(t, err)
		os.Remove(path)
	}

	// read error
	{
		_, err := readPeersJSON(path)
		want := fmt.Sprintf("open %s: no such file or directory", path)
		got := err.Error()
		assert.Equal(t, want, got)
	}

	// write json
	{
		err := writePeersJSON(path, peers)
		assert.Nil(t, err)
	}

	// read json OK
	{
		ps, err := readPeersJSON(path)
		assert.Nil(t, err)
		want := peers
		got := ps
		assert.Equal(t, want, got)
	}

	// json broken
	{
		f, err := os.OpenFile(path, os.O_RDWR, 0644)
		assert.Nil(t, err)
		defer f.Close()

		_, err = f.WriteString("inject")
		assert.Nil(t, err)
	}

	// read error
	{
		_, err := readPeersJSON(path)
		want := "invalid character 'i' looking for beginning of value"
		got := err.Error()
		assert.Equal(t, want, got)
	}
}
