// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package model

import (
	"fmt"
	"time"

	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/fs"
	"github.com/syncthing/syncthing/lib/versioner"
)

func init() {
	folderFactories[config.FolderTypeSink] = newSinkFolder
}

type sinkFolder struct {
	config.FolderConfiguration
}

func newSinkFolder(_ *Model, cfg config.FolderConfiguration, _ versioner.Versioner, _ *fs.MtimeFS) service {
	return &sinkFolder{
		FolderConfiguration: cfg,
	}
}

func (f *sinkFolder) Serve() {
	l.Debugln(f, "starting")
	defer l.Debugln(f, "exiting")
}

func (f *sinkFolder) String() string {
	return fmt.Sprintf("sinkFolder/%s@%p", f.ID, f)
}

func (f *sinkFolder) BringToFront(string)        {}
func (f *sinkFolder) DelayScan(d time.Duration)  {}
func (f *sinkFolder) IndexUpdated()              {}
func (f *sinkFolder) Jobs() ([]string, []string) { return nil, nil }
func (f *sinkFolder) Scan(subs []string) error   { return nil }
func (f *sinkFolder) Stop()                      {}

func (f *sinkFolder) getState() (folderState, time.Time, error) { return FolderIdle, time.Now(), nil }
func (f *sinkFolder) setState(state folderState)                {}
func (f *sinkFolder) clearError()                               {}
func (f *sinkFolder) setError(err error)                        {}
