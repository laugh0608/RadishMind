package bridge

import (
	"context"
	"errors"
	"strings"
)

const (
	ErrorCodeWorkerQueueFull     = "BRIDGE_WORKER_QUEUE_FULL"
	ErrorCodeWorkerTimeout       = "BRIDGE_WORKER_TIMEOUT"
	ErrorCodeWorkerCanceled      = "BRIDGE_WORKER_CANCELED"
	ErrorCodeWorkerExited        = "BRIDGE_WORKER_EXITED"
	ErrorCodeWorkerProtocol      = "BRIDGE_WORKER_PROTOCOL_ERROR"
	ErrorCodeWorkerUnavailable   = "BRIDGE_WORKER_UNAVAILABLE"
	ErrorCodeWorkerRequestFailed = "BRIDGE_WORKER_REQUEST_FAILED"
	ErrorCodeClientClosed        = "BRIDGE_CLIENT_CLOSED"
	ErrorCodeProcessFailed       = "BRIDGE_PROCESS_FAILED"
)

type BridgeError struct {
	code    string
	message string
}

func (e *BridgeError) Error() string {
	if e == nil {
		return "bridge request failed"
	}
	return e.message
}

func (e *BridgeError) Code() string {
	if e == nil {
		return ""
	}
	return e.code
}

func newBridgeError(code string, message string) error {
	return &BridgeError{
		code:    strings.TrimSpace(code),
		message: strings.TrimSpace(message),
	}
}

func ErrorCode(err error) string {
	if err == nil {
		return ""
	}
	var bridgeError *BridgeError
	if errors.As(err, &bridgeError) {
		return bridgeError.Code()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorCodeWorkerTimeout
	}
	if errors.Is(err, context.Canceled) {
		return ErrorCodeWorkerCanceled
	}
	return ""
}

func bridgeContextError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return newBridgeError(ErrorCodeWorkerTimeout, "bridge worker request timed out")
	}
	if errors.Is(err, context.Canceled) {
		return newBridgeError(ErrorCodeWorkerCanceled, "bridge worker request was canceled")
	}
	return err
}
