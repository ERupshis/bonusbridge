package handlers

import (
	"bytes"
	"context"
	"fmt"
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

func TestWithdraw(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBaseBonusesStorage(ctrl)
	gomock.InOrder(
		mockStorage.EXPECT().WithdrawBonuses(gomock.Any(), gomock.Any()).Return(nil),
		mockStorage.EXPECT().WithdrawBonuses(gomock.Any(), gomock.Any()).Return(data.ErrNotEnoughBonuses),
		mockStorage.EXPECT().WithdrawBonuses(gomock.Any(), gomock.Any()).Return(fmt.Errorf("db error")),
	)

	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxWithValue := context.WithValue(r.Context(), middleware.ContextString("userID"), fmt.Sprintf("%d", 1))
		Withdraw(mockStorage, log).ServeHTTP(w, r.WithContext(ctxWithValue))
	})

	type args struct {
		withUserIDinContext bool
		body                []byte
	}
	type want struct {
		statusCode int
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
				body:                []byte("{\"order\":\"2377225624\",\"sum\":45}"),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "without userID in context",
			args: args{
				withUserIDinContext: false,
				body:                []byte(""),
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "damaged json body",
			args: args{
				withUserIDinContext: true,
				body:                []byte("{\"order\":\"2377225624\"\"sum\":45}"),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid order number",
			args: args{
				withUserIDinContext: true,
				body:                []byte("{\"order\":\"23772256241\",\"sum\":45}"),
			},
			want: want{
				statusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name: "not enough bonuses",
			args: args{
				withUserIDinContext: true,
				body:                []byte("{\"order\":\"2377225624\",\"sum\":45}"),
			},
			want: want{
				statusCode: http.StatusPaymentRequired,
			},
		},
		{
			name: "db error",
			args: args{
				withUserIDinContext: true,
				body:                []byte("{\"order\":\"2377225624\",\"sum\":45}"),
			},
			want: want{
				statusCode: http.StatusInternalServerError,
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

			body := bytes.NewBuffer(tt.args.body)
			req, errReq := http.NewRequest(http.MethodPost, ts.URL, body)
			require.NoError(t, errReq)

			resp, errResp := ts.Client().Do(req)
			require.NoError(t, errResp)
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

		})
	}
}
