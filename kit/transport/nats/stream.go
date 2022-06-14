package nats

import (
	"context"
	"time"

	natn "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

type (
	StreamOption func(*Stream)

	// StreamErrorHandler is a function that is called when an error occurs
	StreamErrorHandler func(context.Context, error) error

	Stream struct {
		conn *natn.Conn
		js   natn.JetStreamContext
		opts *natn.Options
		cfg  *natn.StreamConfig
		info *natn.StreamInfo

		errorHandler StreamErrorHandler
	}
)

func WithSubjects(subjects []string) StreamOption {
	return func(s *Stream) {
		s.cfg.Subjects = subjects
	}
}

func WithRetentionPolicy(rp natn.RetentionPolicy) StreamOption {
	return func(s *Stream) {
		s.cfg.Retention = rp
	}
}

func WithMaxConsumers(maxConsumers int) StreamOption {
	return func(s *Stream) {
		s.cfg.MaxConsumers = maxConsumers
	}
}

func WithMaxMsgs(maxMsgs int64) StreamOption {
	return func(s *Stream) {
		s.cfg.MaxMsgs = maxMsgs
	}
}

func WithMaxBytes(maxBytes int64) StreamOption {
	return func(s *Stream) {
		s.cfg.MaxBytes = maxBytes
	}
}

func WithDiscardPolicy(dp natn.DiscardPolicy) StreamOption {
	return func(s *Stream) {
		s.cfg.Discard = dp
	}
}

func WithMaxAge(maxAge time.Duration) StreamOption {
	return func(s *Stream) {
		s.cfg.MaxAge = maxAge
	}
}

func WithMaxMsgSize(maxMsgSize int32) StreamOption {
	return func(s *Stream) {
		s.cfg.MaxMsgSize = maxMsgSize
	}
}

func WithStorage(storage natn.StorageType) StreamOption {
	return func(s *Stream) {
		s.cfg.Storage = storage
	}
}

func WithReplicas(replicas int) StreamOption {
	return func(s *Stream) {
		s.cfg.Replicas = replicas
	}
}

func WithNoAck(noAck bool) StreamOption {
	return func(s *Stream) {
		s.cfg.NoAck = noAck
	}
}

func WithDuplicates(duplicates time.Duration) StreamOption {
	return func(s *Stream) {
		s.cfg.Duplicates = duplicates
	}
}

func defaultStreamErrorHandler(cx context.Context, err error) error {
	return err
}

func NewStream(connstr string, name string, options ...StreamOption) (*Stream, error) {
	var (
		err  error
		cc   *natn.Conn
		opts = natn.GetDefaultOptions()
		st   = &Stream{
			conn:         nil,
			opts:         &opts,
			cfg:          &natn.StreamConfig{},
			info:         &natn.StreamInfo{},
			errorHandler: defaultStreamErrorHandler,
		}
	)

	st.cfg.Name = name

	for _, fn := range options {
		fn(st)
	}

	st.opts.Url = connstr

	// Connect to NATS
	cc, err = st.opts.Connect()

	if err != nil {
		return nil, errors.Wrap(
			err, "unable to connect to nats server",
		)
	}

	// Create JetStream Context
	js, err := cc.JetStream()

	if err != nil {
		return nil, errors.Wrap(
			err, "unable to create JetStream context",
		)
	}

	// Create a Stream
	info, err := js.AddStream(st.cfg)

	if err != nil {
		return nil, errors.Wrap(err, "unable to create jetstream")
	}

	st.js = js
	st.conn = cc
	st.info = info

	return st, nil
}

func (st *Stream) UpdateStream(options ...StreamOption) (*Stream, error) {
	for _, fn := range options {
		fn(st)
	}

	// Update the Stream
	info, err := st.js.UpdateStream(st.cfg)

	if err != nil {
		return nil, errors.Wrap(err, "unable to update jetstream")
	}

	st.info = info

	return st, nil
}

func (st *Stream) DeleteStream() error {

	// Delete the stream
	err := st.js.DeleteStream(st.cfg.Name)

	if err != nil {
		return errors.Wrap(err, "unable to delete jetstream")
	}

	return nil
}
