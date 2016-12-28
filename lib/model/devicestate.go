// Copyright (C) 2016 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package model

import (
	"time"

	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/sync"
)

const deviceIdleTimeout = time.Minute

type deviceState int

const (
	DeviceOffline deviceState = iota
	DeviceSyncing
	DeviceIdle
	DeviceSendingIndexes
)

func (s deviceState) String() string {
	switch s {
	case DeviceOffline:
		return "disconnected"
	case DeviceSyncing:
		return "syncing"
	case DeviceIdle:
		return "idle"
	case DeviceSendingIndexes:
		return "sendingIndexes"
	default:
		return "unknown"
	}
}

type deviceStateTracker struct {
	deviceID protocol.DeviceID

	mut            sync.Mutex
	connected      bool
	sendingIndexes bool
	lastActivity   time.Time
}

func newDeviceStateTracker(id protocol.DeviceID) *deviceStateTracker {
	return &deviceStateTracker{
		deviceID: id,
		mut:      sync.NewMutex(),
	}
}

func (s *deviceStateTracker) state() deviceState {
	if s == nil {
		return DeviceOffline
	}

	s.mut.Lock()
	defer s.mut.Unlock()

	if s.sendingIndexes {
		return DeviceSendingIndexes
	}
	if time.Since(s.lastActivity) < deviceIdleTimeout {
		return DeviceSyncing
	}
	if s.connected {
		return DeviceIdle
	}
	return DeviceOffline
}

func (s *deviceStateTracker) syncActivity() {
	s.mut.Lock()
	s.lastActivity = time.Now()
	s.mut.Unlock()
}

func (s *deviceStateTracker) setConnected() {
	s.mut.Lock()
	s.connected = true
	s.mut.Unlock()
}

func (s *deviceStateTracker) clearConnected() {
	s.mut.Lock()
	s.connected = false
	s.mut.Unlock()
}

func (s *deviceStateTracker) setSendingIndexes() {
	s.mut.Lock()
	s.sendingIndexes = true
	s.mut.Unlock()
}

func (s *deviceStateTracker) clearSendingIndexes() {
	s.mut.Lock()
	s.sendingIndexes = false
	s.mut.Unlock()
}
