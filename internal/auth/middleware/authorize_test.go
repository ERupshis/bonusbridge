package middleware

import (
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

type testHandler struct {
	Message string
}

func (th testHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintln(w, th.Message)
}

func TestAuthorizeUser(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	jwtGen := jwtgenerator.Create("secret_key", 3, log)
	validToken, _ := jwtGen.BuildJWTString(2)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBaseUsersManager(ctrl)
	gomock.InOrder(
		mockStorage.EXPECT().GetUserRole(gomock.Any(), gomock.Any()).Return(data.RoleUser, nil),
		mockStorage.EXPECT().GetUserRole(gomock.Any(), gomock.Any()).Return(-1, fmt.Errorf("db error")),
		mockStorage.EXPECT().GetUserRole(gomock.Any(), gomock.Any()).Return(-1, nil),
		mockStorage.EXPECT().GetUserRole(gomock.Any(), gomock.Any()).Return(data.RoleUser, nil),
	)

	type args struct {
		authorizationHeader string
		role                int
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
				authorizationHeader: string("Bearer ") + validToken,
				role:                data.RoleUser,
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "without Authorization header",
			args: args{
				authorizationHeader: "",
				role:                data.RoleUser,
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "invalid token",
			args: args{
				authorizationHeader: string("Basic ") + validToken,
				role:                data.RoleUser,
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "invalid token",
			args: args{
				authorizationHeader: validToken,
				role:                data.RoleUser,
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "db error",
			args: args{
				authorizationHeader: string("Bearer ") + validToken,
				role:                data.RoleUser,
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "user is not registered",
			args: args{
				authorizationHeader: string("Bearer ") + validToken,
				role:                data.RoleUser,
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "lack or permission",
			args: args{
				authorizationHeader: string("Bearer ") + validToken,
				role:                data.RoleAdmin,
			},
			want: want{
				statusCode: http.StatusForbidden,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(AuthorizeUser(testHandler{}, tt.args.role, mockStorage, jwtGen, log))
			defer ts.Close()

			req, errReq := http.NewRequest(http.MethodPost, ts.URL, nil)
			require.NoError(t, errReq)

			if tt.args.authorizationHeader != "" {
				req.Header.Add("Authorization", tt.args.authorizationHeader)
			}

			resp, errResp := ts.Client().Do(req)
			require.NoError(t, errResp)
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}
