package storage

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
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

func TestStorage_UpdateOrder(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := mocks.NewMockBaseOrdersManager(ctrl)
	gomock.InOrder(
		mockManager.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).Return(nil),
		mockManager.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).Return(fmt.Errorf("manager error")),
	)

	type fields struct {
		manager managers.BaseOrdersManager
		log     logger.BaseLogger
	}
	type args struct {
		ctx   context.Context
		order *data.Order
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
				ctx:   context.Background(),
				order: &data.Order{},
			},
			wantErr: false,
		},
		{
			name: "manager error",
			fields: fields{
				manager: mockManager,
				log:     log,
			},
			args: args{
				ctx:   context.Background(),
				order: &data.Order{},
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
			if err := s.UpdateOrder(tt.args.ctx, tt.args.order); (err != nil) != tt.wantErr {
				t.Errorf("UpdateOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorage_GetOrders(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orders := []data.Order{
		{
			Number: "1234",
			Status: "New",
		},
	}

	mockManager := mocks.NewMockBaseOrdersManager(ctrl)
	gomock.InOrder(
		mockManager.EXPECT().GetOrders(gomock.Any(), gomock.Any()).Return(orders, nil),
		mockManager.EXPECT().GetOrders(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("manager error")),
	)

	type fields struct {
		manager managers.BaseOrdersManager
		log     logger.BaseLogger
	}
	type args struct {
		ctx     context.Context
		filters map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []data.Order
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				manager: mockManager,
				log:     log,
			},
			args: args{
				ctx:     context.Background(),
				filters: nil,
			},
			want:    orders,
			wantErr: false,
		},
		{
			name: "manager error",
			fields: fields{
				manager: mockManager,
				log:     log,
			},
			args: args{
				ctx:     context.Background(),
				filters: nil,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				manager: tt.fields.manager,
				log:     tt.fields.log,
			}
			got, err := s.GetOrders(tt.args.ctx, tt.args.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOrders() got = %v, want %v", got, tt.want)
			}
		})
	}
}
