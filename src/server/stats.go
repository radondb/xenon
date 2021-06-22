/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package server

import (
	"time"

	"model"
)

func (s *Server) getStats() *model.ServerStats {
	return &model.ServerStats{
		Uptimes: uint64(time.Since(s.begin).Seconds()),
	}
}
