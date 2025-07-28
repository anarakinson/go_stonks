package prototime

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoTime(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func FromProtoTime(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	res := t.AsTime()
	return &res
}
