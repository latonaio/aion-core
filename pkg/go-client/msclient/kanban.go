// Copyright (c) 2019-2020 Latona. All rights reserved.

package msclient

import (
	"bitbucket.org/latonaio/aion-core/proto/kanbanpb"
	"encoding/json"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

type WrapKanban struct {
	kanbanpb.StatusKanban
}

func (k *WrapKanban) GetMetadataByMap() (map[string]interface{}, error) {
	var ret map[string]interface{}

	m := k.GetMetadata()
	b, err := protojson.Marshal(m)
	if err != nil {
		return ret, errors.Wrap(err, "cant marshal to map[string]interface{}")
	}
	if err := json.Unmarshal(b, &ret); err != nil {
		return ret, errors.Wrap(err, "cant marshal to map[string]interface{}")
	}
	return ret, nil
}
