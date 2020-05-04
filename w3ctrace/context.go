package w3ctrace

import (
	"errors"
	"net/http"
	"strings"
)

const (
	// The max number of items in `tracestate` as defined by https://www.w3.org/TR/trace-context/#tracestate-header-field-values
	MaxStateEntries = 32

	// W3C trace context header names as defined by https://www.w3.org/TR/trace-context/
	TraceParentHeader = "traceparent"
	TraceStateHeader  = "tracestate"
)

var (
	ErrContextNotFound    = errors.New("no w3c context")
	ErrContextCorrupted   = errors.New("corrupted w3c context")
	ErrUnsupportedVersion = errors.New("unsupported w3c context version")
)

// Context represents the W3C trace context
type Context struct {
	RawParent string
	RawState  string
}

// Extract extracts the W3C trace context from HTTP headers. Returns ErrContextNotFound if
// provided value doesn't contain traceparent header.
func Extract(headers http.Header) (Context, error) {
	var tr Context

	for k, v := range headers {
		if len(v) == 0 {
			continue
		}

		switch {
		case strings.EqualFold(k, TraceParentHeader):
			tr.RawParent = v[0]
		case strings.EqualFold(k, TraceStateHeader):
			tr.RawState = v[0]
		}
	}

	if tr.RawParent == "" {
		return tr, ErrContextNotFound
	}

	return tr, nil
}

// Inject adds the w3c trace context headers, overriding any previously set values
func Inject(trCtx Context, headers http.Header) {
	// delete existing headers ignoring the header name case
	for k := range headers {
		if strings.EqualFold(k, TraceParentHeader) || strings.EqualFold(k, TraceStateHeader) {
			delete(headers, k)
		}
	}

	headers.Set(TraceParentHeader, trCtx.RawParent)
	headers.Set(TraceStateHeader, trCtx.RawState)
}

// State parses RawState and returns the corresponding list.
// It silently discards malformed state. To check errors use ParseState().
func (trCtx Context) State() State {
	st, err := ParseState(trCtx.RawState)
	if err != nil {
		return State{}
	}

	return st
}

// Parent parses RawParent and returns the corresponding list.
// It silently discards malformed value. To check errors use ParseParent().
func (trCtx Context) Parent() Parent {
	st, err := ParseParent(trCtx.RawParent)
	if err != nil {
		return Parent{}
	}

	return st
}
