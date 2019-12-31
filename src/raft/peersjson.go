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
	"io/ioutil"

	"github.com/pkg/errors"
)

func writePeersJSON(path string, peers []string, idlePeers []string) error {
	allPeers := make(map[string][]string)

	allPeers["peers"] = peers
	allPeers["idlepeers"] = idlePeers

	jsonStr, err := json.Marshal(allPeers)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := ioutil.WriteFile(path, []byte(jsonStr), 0755); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func readPeersJSON(path string) ([]string, []string, error) {
	//var peers []string
	//var idlePeers []string
	allPeers := make(map[string][]string)

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return []string{}, []string{}, errors.WithStack(err)
	}

	err = json.Unmarshal(buf, &allPeers)
	if err != nil {
		return []string{}, []string{}, errors.WithStack(err)
	}
	return allPeers["peers"], allPeers["idlepeers"], nil
}
