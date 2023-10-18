package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erupshis/bonusbridge/internal/auth/middleware"
	"github.com/erupshis/bonusbridge/internal/bonuses/data"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalance(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	balance1 := data.Balance{
		Current:   345,
		Withdrawn: 100,
	}

	mockStorage := mocks.NewMockBaseBonusesStorage(ctrl)
	gomock.InOrder(
		mockStorage.EXPECT().GetBalance(gomock.Any(), gomock.Any()).Return(&balance1, nil),
		mockStorage.EXPECT().GetBalance(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("storage error")),
	)

	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxWithValue := context.WithValue(r.Context(), middleware.ContextString("userID"), fmt.Sprintf("%d", 1))
		Balance(mockStorage, log).ServeHTTP(w, r.WithContext(ctxWithValue))
	})

	type args struct {
		withUserIDinContext bool
	}
	type want struct {
		statusCode int
		body       []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid",
			args: args{
				withUserIDinContext: true,
			},
			want: want{
				statusCode: http.StatusOK,
				body:       []byte("{\"current\":345,\"withdrawn\":100}"),
			},
		},
		{
			name: "storage error",
			args: args{
				withUserIDinContext: true,
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				body:       []byte(""),
			},
		},
		{
			name: "without userID in context",
			args: args{
				withUserIDinContext: false,
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				body:       []byte(""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts *httptest.Server
			if tt.args.withUserIDinContext {
				ts = httptest.NewServer(handlerFunc)
			} else {
				ts = httptest.NewServer(Balance(mockStorage, log))
			}
			defer ts.Close()

			req, errReq := http.NewRequest(http.MethodGet, ts.URL, nil)
			require.NoError(t, errReq)

			resp, errResp := ts.Client().Do(req)
			require.NoError(t, errResp)
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, string(tt.want.body), string(respBody))
		})
	}
}
