// Package retryer implements repeat requests logic.
package retryer

import (
	"context"
	"time"

	"github.com/erupshis/bonusbridge/internal/logger"
)

// defIntervals default intervals for repeats.
var defIntervals = []int{1, 3, 5}

// RetryCallWithTimeout generates repeats of function call if error occurs.
// Args:
//   - ctx(context.Context), log Logger.BaseLogger,
//   - intervals([]int) - count of repeats and pause between them (secs.);
//   - repeatableErrors([]error) - errors - reasons to make repeat call. If empty - any error is signal to repeat call;
//   - callback(func(context.Context) (int64, []byte, error)) - function to call.
func RetryCallWithTimeout(ctx context.Context, log logger.BaseLogger, intervals []int, repeatableErrors []error,
	callback func(context.Context) (int64, []byte, error)) (int64, []byte, error) {
	var status int64
	var body []byte
	var err error

	if intervals == nil {
		intervals = defIntervals
	}

	attempt := 0
	for _, interval := range intervals {
		ctxWithTime, cancel := context.WithTimeout(ctx, time.Duration(interval)*time.Second)
		go waitContextToCancel(ctxWithTime, cancel, interval)

		status, body, err = callback(ctxWithTime)
		if err == nil {
			return status, body, nil
		}

		attempt++
		if log != nil {
			log.Info("attempt '%d' to postJSON failed with error: %v", attempt, err)
		}

		if !canRetryCall(err, repeatableErrors) {
			break
		}
	}

	return status, body, err
}

// RetryCallWithTimeoutErrorOnly generates repeats of function call if error occurs.
// Args:
//   - ctx(context.Context), log Logger.BaseLogger,
//   - intervals([]int) - count of repeats and pause between them (secs.);
//   - repeatableErrors([]error) - errors - reasons to make repeat call. If empty - any error is signal to repeat call;
//   - callback (func(context.Context) error) - function to call.
func RetryCallWithTimeoutErrorOnly(ctx context.Context, log logger.BaseLogger, intervals []int, repeatableErrors []error,
	callback func(context.Context) error) error {
	var err error

	if intervals == nil {
		intervals = defIntervals
	}

	attemptNum := 0

	for _, interval := range intervals {
		ctxWithTime, cancel := context.WithTimeout(ctx, time.Duration(interval)*time.Second)
		go waitContextToCancel(ctxWithTime, cancel, interval)

		err = callback(ctxWithTime)
		if err == nil {
			return nil
		}

		attemptNum++
		if log != nil {
			log.Info("attemptNum '%d' to postJSON failed with error: %v", attemptNum, err)
		}

		if !canRetryCall(err, repeatableErrors) {
			log.Info("this kind of error is not retriable: %v", err)
			break
		}
	}

	return err
}

// waitContextToCancel goroutine to prevent timeout context leaking.
func waitContextToCancel(ctx context.Context, cancelFunc context.CancelFunc, interval int) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Duration(interval) * time.Second):
		cancelFunc()
	}
}

// canRetryCall checks if generated error is in list of repeatableErrors.
func canRetryCall(err error, repeatableErrors []error) bool {
	if repeatableErrors == nil {
		return true
	}

	canRetry := false
	for _, repeatableError := range repeatableErrors {
		if err.Error() == repeatableError.Error() {
			canRetry = true
		}
	}

	return canRetry
}
