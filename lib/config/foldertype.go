// Copyright (C) 2016 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package config

type FolderType int

const (
	FolderTypeSendReceive FolderType = iota // default is sendreceive
	FolderTypeSendOnly
	FolderTypeSink
)

func (t FolderType) String() string {
	switch t {
	case FolderTypeSendReceive:
		return "readwrite"
	case FolderTypeSendOnly:
		return "readonly"
	case FolderTypeSink:
		return "sink"
	default:
		return "unknown"
	}
}

func (t FolderType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *FolderType) UnmarshalText(bs []byte) error {
	switch string(bs) {
	case "readwrite", "sendreceive":
		*t = FolderTypeSendReceive
	case "readonly", "sendonly":
		*t = FolderTypeSendOnly
	case "sink":
		*t = FolderTypeSink
	default:
		*t = FolderTypeSendReceive
	}
	return nil
}
