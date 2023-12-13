package nats

import (
	"errors"

	"github.com/nats-io/nats.go"
)

var ErrNotFound = errors.New("not found")

type (
	StreamConfig nats.StreamConfig
	StreamInfo   nats.StreamInfo

	ConsumerConfig nats.ConsumerConfig
	ConsumerInfo   nats.ConsumerInfo
)

var (
	storageType = map[string]nats.StorageType{
		"file":   nats.FileStorage,
		"memory": nats.MemoryStorage,
	}
	invertedStorageType = invertMap(storageType)
	ToStorageType       = mapFn(storageType)
	FromStorageType     = mapFn(invertedStorageType)
)

var (
	discardPolicy = map[string]nats.DiscardPolicy{
		"old": nats.DiscardOld,
		"new": nats.DiscardNew,
	}
	invertedDiscardPolicy = invertMap(discardPolicy)
	ToDiscardPolicy       = mapFn(discardPolicy)
	FromDiscardPolicy     = mapFn(invertedDiscardPolicy)
)

var (
	retentionPolicy = map[string]nats.RetentionPolicy{
		"limits":   nats.LimitsPolicy,
		"interest": nats.InterestPolicy,
		"work":     nats.WorkQueuePolicy,
	}
	invertedRetentionPolicy = invertMap(retentionPolicy)
	ToRetentionPolicy       = mapFn(retentionPolicy)
	FromRetentionPolicy     = mapFn(invertedRetentionPolicy)
)

var (
	ackPolicy = map[string]nats.AckPolicy{
		"none":     nats.AckNonePolicy,
		"all":      nats.AckAllPolicy,
		"explicit": nats.AckExplicitPolicy,
	}
	invertedAckPolicy = invertMap(ackPolicy)
	ToAckPolicy       = mapFn(ackPolicy)
	FromAckPolicy     = mapFn(invertedAckPolicy)
)

var (
	deliverPolicy = map[string]nats.DeliverPolicy{
		"all":  nats.DeliverAllPolicy,
		"new":  nats.DeliverNewPolicy,
		"last": nats.DeliverLastPolicy,
	}
	invertedDeliverPolicy = invertMap(deliverPolicy)
	ToDeliverPolicy       = mapFn(deliverPolicy)
	FromDeliverPolicy     = mapFn(invertedDeliverPolicy)
)
