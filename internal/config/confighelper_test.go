package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtoi64(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "valid base case",
			args: args{
				value: "123",
			},
			want:    123,
			wantErr: false,
		},
		{
			name: "invalid float on input",
			args: args{
				value: "123.5",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "invalid text on input",
			args: args{
				value: "add",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Atoi64(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQueryValuesIntoMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "Atoi64(%v)", tt.args.value)
		})
	}
}

func TestSetEnvToParamIfNeedInt64(t *testing.T) {
	type args struct {
		param int64
		val   string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "valid base case",
			args: args{
				param: 0,
				val:   "123",
			},
			want:    123,
			wantErr: false,
		},
		{
			name: "valid empty val",
			args: args{
				param: 0,
				val:   "",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "invalid text as val",
			args: args{
				param: 0,
				val:   "dd",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetEnvToParamIfNeed(&tt.args.param, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetEnvToParamIfNeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, tt.args.param)
		})
	}
}

func TestSetEnvToParamIfNeedInt(t *testing.T) {
	type args struct {
		param int
		val   string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "valid base case",
			args: args{
				param: 0,
				val:   "123",
			},
			want:    123,
			wantErr: false,
		},
		{
			name: "valid empty val",
			args: args{
				param: 0,
				val:   "",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "invalid text as val",
			args: args{
				param: 0,
				val:   "dd",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetEnvToParamIfNeed(&tt.args.param, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetEnvToParamIfNeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, tt.args.param)
		})
	}
}

func TestSetEnvToParamIfNeedString(t *testing.T) {
	type args struct {
		param string
		val   string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid base case",
			args: args{
				param: "",
				val:   "dd",
			},
			want:    "dd",
			wantErr: false,
		},
		{
			name: "valid empty val",
			args: args{
				param: "",
				val:   "",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "valid int as val",
			args: args{
				param: "",
				val:   "123",
			},
			want:    "123",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetEnvToParamIfNeed(&tt.args.param, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetEnvToParamIfNeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, tt.args.param)
		})
	}
}

func TestSetEnvToParamIfNeedStringSlice(t *testing.T) {
	type args struct {
		param []string
		val   string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "valid base case",
			args: args{
				param: []string{},
				val:   "dd",
			},
			want:    []string{"dd"},
			wantErr: false,
		},
		{
			name: "valid base case",
			args: args{
				param: []string{},
				val:   "dd,asd",
			},
			want:    []string{"dd", "asd"},
			wantErr: false,
		},
		{
			name: "valid empty val",
			args: args{
				param: []string{},
				val:   "",
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "valid int as val",
			args: args{
				param: []string{},
				val:   "123",
			},
			want:    []string{"123"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetEnvToParamIfNeed(&tt.args.param, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetEnvToParamIfNeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, tt.args.param)
		})
	}
}

func TestSetEnvToParamIfNeedInterface(t *testing.T) {
	type args struct {
		param interface{}
		val   string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "valid base case",
			args: args{
				param: 0,
				val:   "123",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "valid empty val",
			args: args{
				param: 0,
				val:   "",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "invalid text as val",
			args: args{
				param: 0,
				val:   "dd",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetEnvToParamIfNeed(&tt.args.param, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetEnvToParamIfNeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, tt.args.param)
		})
	}
}
