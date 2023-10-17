package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	jwtGen := jwtgenerator.Create("secret_key", 3, log)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBaseUsersManager(ctrl)
	gomock.InOrder(
		mockStorage.EXPECT().GetUserID(gomock.Any(), gomock.Any()).Return(int64(-1), nil),
		mockStorage.EXPECT().AddUser(gomock.Any(), gomock.Any()).Return(int64(1), nil),
		mockStorage.EXPECT().GetUserID(gomock.Any(), gomock.Any()).Return(int64(-1), fmt.Errorf("failed to find user(db error)")),
		mockStorage.EXPECT().GetUserID(gomock.Any(), gomock.Any()).Return(int64(1), nil),
		mockStorage.EXPECT().GetUserID(gomock.Any(), gomock.Any()).Return(int64(-1), nil),
		mockStorage.EXPECT().AddUser(gomock.Any(), gomock.Any()).Return(int64(1), fmt.Errorf("failed to add user(db error)")),
	)

	ts := httptest.NewServer(Register(mockStorage, jwtGen, log))
	defer ts.Close()

	type args struct {
		body []byte
	}
	type want struct {
		statusCode          int
		authorizationHeader bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid",
			args: args{
				body: []byte(`{
						"login":"u2", 
						"password":"p1"
					}`),
			},
			want: want{
				statusCode:          http.StatusOK,
				authorizationHeader: true,
			},
		},
		{
			name: "fail unmarshalling request body",
			args: args{
				body: []byte(`{
						"login":"u2" 
						"password":"p1"
					}`),
			},
			want: want{
				statusCode:          http.StatusBadRequest,
				authorizationHeader: false,
			},
		},
		{
			name: "db returns error",
			args: args{
				body: []byte(`{
						"login":"u2", 
						"password":"p1"
					}`),
			},
			want: want{
				statusCode:          http.StatusInternalServerError,
				authorizationHeader: false,
			},
		},
		{
			name: "user login already exists in db",
			args: args{
				body: []byte(`{
						"login":"u2", 
						"password":"p1"
					}`),
			},
			want: want{
				statusCode:          http.StatusConflict,
				authorizationHeader: false,
			},
		},
		{
			name: "error on user add in db",
			args: args{
				body: []byte(`{
						"login":"u2", 
						"password":"p1"
					}`),
			},
			want: want{
				statusCode:          http.StatusInternalServerError,
				authorizationHeader: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewBuffer(tt.args.body)
			req, errReq := http.NewRequest(http.MethodPost, ts.URL, body)
			require.NoError(t, errReq)

			req.Header.Add("Content-Type", "application/json")

			resp, errResp := ts.Client().Do(req)
			require.NoError(t, errResp)
			defer helpers.ExecuteWithLogError(resp.Body.Close, log)

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.authorizationHeader, resp.Header.Get("Authorization") != "")
		})
	}
}
