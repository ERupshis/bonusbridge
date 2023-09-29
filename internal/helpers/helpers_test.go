package helpers

import (
	"testing"

	"github.com/erupshis/bonusbridge/internal/logger"
)

func TestExecuteWithLogError(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")

	type args struct {
		callback func() error
		log      logger.BaseLogger
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "valid base case",
			args: args{
				callback: func() error {
					return nil
				},
				log: log,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ExecuteWithLogError(tt.args.callback, tt.args.log)
		})
	}
}

func TestInterfaceToString(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid base case",
			args: args{
				i: "string type",
			},
			want: "string type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InterfaceToString(tt.args.i); got != tt.want {
				t.Errorf("InterfaceToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
