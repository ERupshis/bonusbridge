package auth

import (
	"context"
	"testing"

	"github.com/erupshis/bonusbridge/internal/auth/middleware"
	"github.com/erupshis/bonusbridge/internal/auth/users/data"
)

func TestGetUserIDFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				ctx: context.WithValue(context.Background(), middleware.ContextString(data.UserID), "5"),
			},
			want:    5,
			wantErr: false,
		},
		{
			name: "without field in context",
			args: args{
				ctx: context.Background(),
			},
			want:    -1,
			wantErr: true,
		},
		{
			name: "wrong value type",
			args: args{
				ctx: context.WithValue(context.Background(), middleware.ContextString(data.UserID), "as"),
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserIDFromContext(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserIDFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetUserIDFromContext() got = %v, want %v", got, tt.want)
			}
		})
	}
}
