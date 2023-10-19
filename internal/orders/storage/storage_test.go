package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/storage/managers"
	"github.com/erupshis/bonusbridge/mocks"
	"github.com/golang/mock/gomock"
)

func TestStorage_AddOrder(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := mocks.NewMockBaseOrdersManager(ctrl)
	gomock.InOrder(
		mockManager.EXPECT().AddOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil),
		mockManager.EXPECT().AddOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(-1), fmt.Errorf("manager error")),
	)

	type fields struct {
		manager managers.BaseOrdersManager
		log     logger.BaseLogger
	}
	type args struct {
		ctx    context.Context
		number string
		userID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				manager: mockManager,
				log:     log,
			},
			args: args{
				ctx:    context.Background(),
				number: "1234",
				userID: 1,
			},
			wantErr: false,
		},
		{
			name: "error from manager",
			fields: fields{
				manager: mockManager,
				log:     log,
			},
			args: args{
				ctx:    context.Background(),
				number: "1234",
				userID: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				manager: tt.fields.manager,
				log:     tt.fields.log,
			}
			if err := s.AddOrder(tt.args.ctx, tt.args.number, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("AddOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
