package validator

import (
	"testing"
)

func TestIsLuhnValid(t *testing.T) {
	type args struct {
		number string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid",
			args: args{
				number: "2377225624",
			},
			want: true,
		},
		{
			name: "failed validation",
			args: args{
				number: "7225624",
			},
			want: false,
		},
		{
			name: "failed validation (invalid symbol)",
			args: args{
				number: "as7225624",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLuhnValid(tt.args.number); got != tt.want {
				t.Errorf("IsLuhnValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
