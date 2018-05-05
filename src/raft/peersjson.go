/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package raft

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
)

func writePeersJSON(path string, peers []string) error {
	peersJSON, err := json.Marshal(peers)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := ioutil.WriteFile(path, []byte(peersJSON), 0755); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func readPeersJSON(path string) ([]string, error) {
	var peers []string

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return []string{}, errors.WithStack(err)
	}

	err = json.Unmarshal(buf, &peers)
	if err != nil {
		return []string{}, errors.WithStack(err)
	}
	return peers, nil
}
