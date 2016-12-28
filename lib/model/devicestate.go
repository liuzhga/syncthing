// Copyright (C) 2016 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package model

import (
	"fmt"
	"time"

	"net"

	"github.com/syncthing/syncthing/lib/events"
	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/sync"
)

// The deviceState represents the the current state and activity level of a
// given device.

type deviceState int

const (
	deviceDisconnected   deviceState = iota
	deviceSyncing                    // Have received a request within the last deviceIdleTimeout
	deviceIdle                       // Connected, but no recent request
	devicePreparingIndex             // Sorting index for transmission
	deviceSendingIndex               // Sending index data
)

func (s deviceState) String() string {
	switch s {
	case deviceDisconnected:
		return "disconnected"
	case deviceSyncing:
		return "syncing"
	case deviceIdle:
		return "idle"
	case devicePreparingIndex:
		return "preparingIndex"
	case deviceSendingIndex:
		return "sendingIndex"
	default:
		return "unknown"
	}
}

// deviceStateTracker keeps track of events and the current state for a
// given device

type deviceStateTracker struct {
	id protocol.DeviceID

	mut                  sync.Mutex
	connected            bool
	preparingIndex       int
	sendingIndex         int
	lastActivity         time.Time
	prevState            deviceState
	activityTimeoutTimer *time.Timer
}

func newDeviceStateTracker(id protocol.DeviceID) *deviceStateTracker {
	s := &deviceStateTracker{
		id:  id,
		mut: sync.NewMutex(),
	}
	s.activityTimeoutTimer = time.AfterFunc(time.Hour, s.activityTimeout)
	return s
}

// When we have not received a Request for this long the device is
// considered idle.
const deviceIdleTimeout = 30 * time.Second

func (s *deviceStateTracker) State() deviceState {
	s.mut.Lock()
	defer s.mut.Unlock()
	return s.stateLocked()
}

func (s *deviceStateTracker) stateLocked() deviceState {
	switch {
	case time.Since(s.lastActivity) < deviceIdleTimeout:
		return deviceSyncing
	case s.preparingIndex > 0:
		return devicePreparingIndex
	case s.sendingIndex > 0:
		return deviceSendingIndex
	case s.connected:
		return deviceIdle
	default:
		return deviceDisconnected
	}
}

func (s *deviceStateTracker) Connected(hello protocol.HelloResult, connType string, addr net.Addr) {
	s.mut.Lock()
	s.connected = true
	s.prevState = deviceIdle
	s.mut.Unlock()

	event := map[string]string{
		"id":            s.id.String(),
		"deviceName":    hello.DeviceName,
		"clientName":    hello.ClientName,
		"clientVersion": hello.ClientVersion,
		"type":          connType,
	}
	if addr != nil {
		event["addr"] = addr.String()
	}

	events.Default.Log(events.DeviceConnected, event)
}

func (s *deviceStateTracker) Disconnected(err error) {
	s.mut.Lock()
	s.connected = false
	s.prevState = deviceDisconnected
	s.mut.Unlock()

	events.Default.Log(events.DeviceDisconnected, map[string]string{
		"id":    s.id.String(),
		"error": err.Error(),
	})
}

func (s *deviceStateTracker) SyncActivity() {
	s.mut.Lock()
	fmt.Println("sync activity")
	s.lastActivity = time.Now()
	s.activityTimeoutTimer.Reset(deviceIdleTimeout)
	s.stateChangedLocked()
	s.mut.Unlock()
}

func (s *deviceStateTracker) PreparingIndex() {
	s.mut.Lock()
	s.preparingIndex++
	s.stateChangedLocked()
	s.mut.Unlock()
}

func (s *deviceStateTracker) SendingIndex() {
	s.mut.Lock()
	s.preparingIndex--
	if s.preparingIndex < 0 {
		panic("unmatched call to setSendingIndex")
	}
	s.sendingIndex++
	s.stateChangedLocked()
	s.mut.Unlock()
}

func (s *deviceStateTracker) DoneSendingIndex() {
	s.mut.Lock()
	s.sendingIndex--
	if s.sendingIndex < 0 {
		panic("unmatched call to doneSendingIndex")
	}
	s.stateChangedLocked()
	s.mut.Unlock()
}

func (s *deviceStateTracker) stateChangedLocked() {
	state := s.stateLocked()
	if state != s.prevState {
		events.Default.Log(events.DeviceStateChanged, map[string]string{
			"id":   s.id.String(),
			"from": s.prevState.String(),
			"to":   state.String(),
		})
		s.prevState = state
	}
}

func (s *deviceStateTracker) activityTimeout() {
	s.mut.Lock()
	fmt.Println("idle timer")
	s.stateChangedLocked()
	s.mut.Unlock()
}

// deviceStateMap is a concurrency safe container for deviceStateTrackers

type deviceStateMap struct {
	mut    sync.Mutex
	states map[protocol.DeviceID]*deviceStateTracker
}

func newDeviceStateMap() *deviceStateMap {
	return &deviceStateMap{
		mut:    sync.NewMutex(),
		states: make(map[protocol.DeviceID]*deviceStateTracker),
	}
}

func (s *deviceStateMap) Get(id protocol.DeviceID) *deviceStateTracker {
	s.mut.Lock()
	st, ok := s.states[id]
	if !ok {
		st = newDeviceStateTracker(id)
		s.states[id] = st
	}
	s.mut.Unlock()
	return st
}

func (s *deviceStateMap) Delete(id protocol.DeviceID) {
	s.mut.Lock()
	delete(s.states, id)
	s.mut.Unlock()
}
