package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erupshis/bonusbridge/internal/auth/middleware"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	"github.com/erupshis/bonusbridge/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrders(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orders := []data.Order{
		{
			Number:  "12344",
			Status:  "NEW",
			Accrual: 500,
		},
	}

	mockStorage := mocks.NewMockBaseOrdersStorage(ctrl)
	gomock.InOrder(
		mockStorage.EXPECT().GetOrders(gomock.Any(), gomock.Any()).Return(orders, nil),
		mockStorage.EXPECT().GetOrders(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("storage error")),
		mockStorage.EXPECT().GetOrders(gomock.Any(), gomock.Any()).Return(nil, nil),
	)

	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxWithValue := context.WithValue(r.Context(), middleware.ContextString("userID"), fmt.Sprintf("%d", 1))
		GetOrders(mockStorage, log).ServeHTTP(w, r.WithContext(ctxWithValue))
	})

	type args struct {
		withUserIDinContext bool
	}
	type want struct {
		statusCode  int
		contentType string
		body        []byte
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
				statusCode:  http.StatusOK,
				contentType: "application/json",
				body:        []byte("[{\"number\":\"12344\",\"status\":\"NEW\",\"accrual\":500,\"uploaded_at\":\"0001-01-01T00:00:00Z\"}]"),
			},
		},
		{
			name: "without userID in context",
			args: args{
				withUserIDinContext: false,
			},
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "",
				body:        []byte(""),
			},
		},
		{
			name: "storage returns error",
			args: args{
				withUserIDinContext: true,
			},
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "",
				body:        []byte(""),
			},
		},
		{
			name: "no user's orders in db",
			args: args{
				withUserIDinContext: true,
			},
			want: want{
				statusCode:  http.StatusNoContent,
				contentType: "",
				body:        []byte(""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts *httptest.Server
			if tt.args.withUserIDinContext {
				ts = httptest.NewServer(handlerFunc)
			} else {
				ts = httptest.NewServer(GetOrders(mockStorage, log))
			}
			defer ts.Close()

			req, errReq := http.NewRequest(http.MethodPost, ts.URL, nil)
			require.NoError(t, errReq)

			resp, errResp := ts.Client().Do(req)
			require.NoError(t, errResp)
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, string(tt.want.body), string(respBody))

		})
	}
}
