package retryer

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
)

var databaseErrorsToRetry = []error{
	errors.New(pgerrcode.UniqueViolation),
	errors.New(pgerrcode.ConnectionException),
	errors.New(pgerrcode.ConnectionDoesNotExist),
	errors.New(pgerrcode.ConnectionFailure),
	errors.New(pgerrcode.SQLClientUnableToEstablishSQLConnection),
	errors.New(pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection),
	errors.New(pgerrcode.TransactionResolutionUnknown),
	errors.New(pgerrcode.ProtocolViolation),
}

func Test_canRetryCall(t *testing.T) {
	type args struct {
		err              error
		repeatableErrors []error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid",
			args: args{
				err:              errors.New(`08000`),
				repeatableErrors: databaseErrorsToRetry,
			},
			want: true,
		},
		{
			name: "valid with missing slice",
			args: args{
				err:              errors.New(`any error`),
				repeatableErrors: nil,
			},
			want: true,
		},
		{
			name: "invalid error is not in slice",
			args: args{
				err:              errors.New(`any error`),
				repeatableErrors: databaseErrorsToRetry,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, canRetryCall(tt.args.err, tt.args.repeatableErrors))
		})
	}
}

func TestRetryCallWithTimeout(t *testing.T) {
	log, _ := logger.CreateZapLogger("Info")

	type args struct {
		ctx              context.Context
		log              logger.BaseLogger
		intervals        []int
		repeatableErrors []error
		callback         func(context.Context) (int64, []byte, error)
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "valid",
			args: args{
				ctx:              context.Background(),
				log:              log,
				intervals:        []int{1, 1, 1},
				repeatableErrors: nil,
				callback: func(ctx context.Context) (int64, []byte, error) {
					<-ctx.Done()
					return http.StatusRequestTimeout, []byte{}, errors.New(pgerrcode.ConnectionException)

				},
			},
			wantErr: errors.New(pgerrcode.ConnectionException),
		},
		{
			name: "valid with success",
			args: args{
				ctx:              context.Background(),
				log:              log,
				intervals:        nil,
				repeatableErrors: nil,
				callback: func(ctx context.Context) (int64, []byte, error) {
					return http.StatusOK, []byte{}, nil
				},
			},
			wantErr: nil,
		},
		{
			name: "valid should retry",
			args: args{
				ctx:              context.Background(),
				log:              log,
				intervals:        []int{1, 1, 1},
				repeatableErrors: databaseErrorsToRetry,
				callback: func(ctx context.Context) (int64, []byte, error) {
					<-ctx.Done()
					return http.StatusRequestTimeout, []byte{}, errors.New(pgerrcode.ConnectionException)
				},
			},
			wantErr: errors.New(pgerrcode.ConnectionException),
		},
		{
			name: "valid shouldn't retry",
			args: args{
				ctx:              context.Background(),
				log:              log,
				intervals:        []int{1, 1, 1},
				repeatableErrors: databaseErrorsToRetry,
				callback: func(ctx context.Context) (int64, []byte, error) {
					<-ctx.Done()
					return http.StatusRequestTimeout, []byte{}, errors.New("some error")
				},
			},
			wantErr: errors.New("some error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := RetryCallWithTimeout(tt.args.ctx, tt.args.log, tt.args.intervals, tt.args.repeatableErrors, tt.args.callback)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestRetryCallWithTimeoutErrorOnly(t *testing.T) {
	log, _ := logger.CreateZapLogger("Info")

	type args struct {
		ctx              context.Context
		log              logger.BaseLogger
		intervals        []int
		repeatableErrors []error
		callback         func(context.Context) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "valid",
			args: args{
				ctx:              context.Background(),
				log:              log,
				intervals:        []int{1, 1, 1},
				repeatableErrors: nil,
				callback: func(ctx context.Context) error {
					<-ctx.Done()
					return errors.New(pgerrcode.ConnectionException)

				},
			},
			wantErr: errors.New(pgerrcode.ConnectionException),
		},
		{
			name: "valid with success",
			args: args{
				ctx:              context.Background(),
				log:              log,
				intervals:        nil,
				repeatableErrors: nil,
				callback: func(ctx context.Context) error {
					return nil
				},
			},
			wantErr: nil,
		},
		{
			name: "valid should retry",
			args: args{
				ctx:              context.Background(),
				log:              log,
				intervals:        []int{1, 1, 1},
				repeatableErrors: databaseErrorsToRetry,
				callback: func(ctx context.Context) error {
					<-ctx.Done()
					return errors.New(pgerrcode.ConnectionException)
				},
			},
			wantErr: errors.New(pgerrcode.ConnectionException),
		},
		{
			name: "valid shouldn't retry",
			args: args{
				ctx:              context.Background(),
				log:              log,
				intervals:        []int{1, 1, 1},
				repeatableErrors: databaseErrorsToRetry,
				callback: func(ctx context.Context) error {
					<-ctx.Done()
					return errors.New("some error")
				},
			},
			wantErr: errors.New("some error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RetryCallWithTimeoutErrorOnly(tt.args.ctx, tt.args.log, tt.args.intervals, tt.args.repeatableErrors, tt.args.callback)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
