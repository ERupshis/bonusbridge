package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	jwtGen := jwtgenerator.Create("secret_key", 3, log)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user1 := data.User{
		Login:    "u1",
		Password: "p1",
		ID:       1,
		Role:     data.RoleUser,
	}

	mockStorage := mocks.NewMockBaseUsersManager(ctrl)
	gomock.InOrder(
		mockStorage.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(&user1, nil),
		mockStorage.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("db error")),
		mockStorage.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(nil, nil),
		mockStorage.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(&user1, nil),
	)

	ts := httptest.NewServer(Login(mockStorage, jwtGen, log))
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
						"login":"u1", 
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
						"login":"u1" 
						"password":"p1"
					}`),
			},
			want: want{
				statusCode:          http.StatusBadRequest,
				authorizationHeader: false,
			},
		},
		{
			name: "error from database",
			args: args{
				body: []byte(`{
						"login":"u1", 
						"password":"p1"
					}`),
			},
			want: want{
				statusCode:          http.StatusInternalServerError,
				authorizationHeader: false,
			},
		},
		{
			name: "missing user in DB",
			args: args{
				body: []byte(`{
						"login":"u1", 
						"password":"p1"
					}`),
			},
			want: want{
				statusCode:          http.StatusUnauthorized,
				authorizationHeader: false,
			},
		},
		{
			name: "incorrect password",
			args: args{
				body: []byte(`{
						"login":"u1", 
						"password":"p2"
					}`),
			},
			want: want{
				statusCode:          http.StatusUnauthorized,
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
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.authorizationHeader, resp.Header.Get("Authorization") != "")
		})
	}
}
