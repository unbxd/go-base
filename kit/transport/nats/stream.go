package nats

import (
	"time"

	natn "github.com/nats-io/nats.go"
)

type (
	StreamOption func(*natn.StreamConfig)
)

func WithRetentionPolicy(rp natn.RetentionPolicy) StreamOption {
	return func(s *natn.StreamConfig) {
		s.Retention = rp
	}
}

func WithMaxConsumers(maxConsumers int) StreamOption {
	return func(s *natn.StreamConfig) {
		s.MaxConsumers = maxConsumers
	}
}

func WithMaxMsgs(maxMsgs int64) StreamOption {
	return func(s *natn.StreamConfig) {
		s.MaxMsgs = maxMsgs
	}
}

func WithMaxBytes(maxBytes int64) StreamOption {
	return func(s *natn.StreamConfig) {
		s.MaxBytes = maxBytes
	}
}

func WithDiscardPolicy(dp natn.DiscardPolicy) StreamOption {
	return func(s *natn.StreamConfig) {
		s.Discard = dp
	}
}

func WithMaxAge(maxAge time.Duration) StreamOption {
	return func(s *natn.StreamConfig) {
		s.MaxAge = maxAge
	}
}

func WithMaxMsgSize(maxMsgSize int32) StreamOption {
	return func(s *natn.StreamConfig) {
		s.MaxMsgSize = maxMsgSize
	}
}

func WithStorage(storage natn.StorageType) StreamOption {
	return func(s *natn.StreamConfig) {
		s.Storage = storage
	}
}

func WithReplicas(replicas int) StreamOption {
	return func(s *natn.StreamConfig) {
		s.Replicas = replicas
	}
}

func WithNoAck(noAck bool) StreamOption {
	return func(s *natn.StreamConfig) {
		s.NoAck = noAck
	}
}

func WithDuplicates(duplicates time.Duration) StreamOption {
	return func(s *natn.StreamConfig) {
		s.Duplicates = duplicates
	}
}
