package handlers

import (
	"bytes"
	"context"
	"fmt"
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

func TestAddOrderHandler(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockBaseOrdersStorage(ctrl)
	gomock.InOrder(
		mockStorage.EXPECT().AddOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil),
		mockStorage.EXPECT().AddOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(data.ErrOrderWasAddedBefore),
		mockStorage.EXPECT().AddOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(data.ErrOrderWasAddedByAnotherUser),
		mockStorage.EXPECT().AddOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("unexpected storage error")),
	)

	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxWithValue := context.WithValue(r.Context(), middleware.ContextString("userID"), fmt.Sprintf("%d", 1))
		AddOrder(mockStorage, log).ServeHTTP(w, r.WithContext(ctxWithValue))
	})

	type args struct {
		withUserIDinContext bool
		contentType         string
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
				contentType:         "text/plain",
				body:                []byte("371449635398431"),
			},
			want: want{
				statusCode: http.StatusAccepted,
			},
		},
		{
			name: "wrong content type",
			args: args{
				withUserIDinContext: true,
				contentType:         "application/json",
				body:                []byte("371449635398431"),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed luhn's check",
			args: args{
				withUserIDinContext: true,
				contentType:         "text/plain",
				body:                []byte("371449635398"),
			},
			want: want{
				statusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name: "without userID in context",
			args: args{
				withUserIDinContext: false,
				contentType:         "text/plain",
				body:                []byte("371449635398431"),
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "order was added before",
			args: args{
				withUserIDinContext: true,
				contentType:         "text/plain",
				body:                []byte("371449635398431"),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "order was added by another person",
			args: args{
				withUserIDinContext: true,
				contentType:         "text/plain",
				body:                []byte("371449635398431"),
			},
			want: want{
				statusCode: http.StatusConflict,
			},
		},
		{
			name: "unexpected error from storage",
			args: args{
				withUserIDinContext: true,
				contentType:         "text/plain",
				body:                []byte("371449635398431"),
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
				ts = httptest.NewServer(AddOrder(mockStorage, log))
			}
			defer ts.Close()

			body := bytes.NewBuffer(tt.args.body)
			req, errReq := http.NewRequest(http.MethodPost, ts.URL, body)
			require.NoError(t, errReq)

			req.Header.Set("Content-Type", tt.args.contentType)

			resp, errResp := ts.Client().Do(req)
			require.NoError(t, errResp)
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

		})
	}
}
