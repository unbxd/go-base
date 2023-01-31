package notifier

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/Showmax/go-fqdn"
	"github.com/gofrs/uuid"
)

type (
	conx struct {
		Time   int64    `json:"timestamp"`
		Caller string   `json:"caller"`
		Calls  []string `json:"call_stack"`
		Host   string   `json:"host"`
		Impl   string   `json:"impl"`
	}

	event struct {
		Data    interface{} `json:"data"`
		Context *conx       `json:"context"`
		ID      string      `json:"id"`
		Subject string      `json:"subject"`
	}

	decorator func(*conx)

	Notifier interface {
		// Notify publishes data to the notifier
		// after decorating it with additional details
		Notify(
			cx context.Context, data interface{},
		) error
	}
)

func decorateHost() decorator {
	return func(cx *conx) {
		host, err := fqdn.FqdnHostname()
		if err != nil {
			host, err = os.Hostname()
			if err != nil {
				host = "unknown-host-name"
			}
		}

		cx.Host = host
	}
}

func decorateCallstack() decorator {
	return func(cx *conx) {
		// caller
		_, ff, ln, ok := runtime.Caller(3)
		if ok {
			cx.Caller = fmt.Sprintf("%s:[%d]", ff, ln)
		}

		// stack trace
		cx.Calls = []string{cx.Caller}

		for i := 4; i < 6; i++ {
			_, ff, ln, ok := runtime.Caller(i)
			if ok {
				cx.Calls = append(cx.Calls, fmt.Sprintf("%s:[%d]", ff, ln))
			}
		}
	}
}

func decorateTimestamp() decorator {
	return func(cx *conx) { cx.Time = time.Now().Unix() }
}

func uuidfn() string {
	var id string

	uid, err := uuid.NewV4()
	if err != nil {
		id = strconv.FormatInt(
			time.Now().UnixNano(), 10,
		)
	} else {
		id = uid.String()
	}

	return id
}

func newEvent(impl string, data interface{}) *event {
	cx := &conx{}

	for _, fn := range []decorator{
		decorateHost(),
		decorateTimestamp(),
		decorateCallstack(),
	} {
		fn(cx)
	}

	cx.Impl = impl

	return &event{
		ID:      uuidfn(),
		Context: cx,
		Data:    data,
		Subject: "subject-not-set",
	}
}
