package storage

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/erupshis/bonusbridge/internal/bonuses/data"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/mocks"
	"github.com/golang/mock/gomock"
)

func TestStorage_WithdrawBonuses(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := mocks.NewMockBaseBonusesManager(ctrl)
	gomock.InOrder(
		mockManager.EXPECT().WithdrawBonuses(gomock.Any(), gomock.Any()).Return(nil),
		mockManager.EXPECT().WithdrawBonuses(gomock.Any(), gomock.Any()).Return(fmt.Errorf("manager error")),
	)

	type fields struct {
		manager managers.BaseBonusesManager
		log     logger.BaseLogger
	}
	type args struct {
		ctx        context.Context
		withdrawal *data.Withdrawal
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
				ctx:        context.Background(),
				withdrawal: &data.Withdrawal{},
			},
			wantErr: false,
		},
		{
			name: "manager returns error",
			fields: fields{
				manager: mockManager,
				log:     log,
			},
			args: args{
				ctx:        context.Background(),
				withdrawal: &data.Withdrawal{},
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
			if err := s.WithdrawBonuses(tt.args.ctx, tt.args.withdrawal); (err != nil) != tt.wantErr {
				t.Errorf("WithdrawBonuses() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorage_GetBalance(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := mocks.NewMockBaseBonusesManager(ctrl)
	gomock.InOrder(
		mockManager.EXPECT().GetBalanceDif(gomock.Any(), gomock.Any()).Return(float32(100.0), nil),
		mockManager.EXPECT().GetBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(float32(-30.0), nil),
		mockManager.EXPECT().GetBalanceDif(gomock.Any(), gomock.Any()).Return(float32(100.0), fmt.Errorf("dif error")),
		mockManager.EXPECT().GetBalanceDif(gomock.Any(), gomock.Any()).Return(float32(100.0), nil),
		mockManager.EXPECT().GetBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(float32(-30.0), fmt.Errorf("common error")),
	)

	type fields struct {
		manager managers.BaseBonusesManager
		log     logger.BaseLogger
	}
	type args struct {
		ctx    context.Context
		userID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *data.Balance
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
				userID: 1,
			},
			want: &data.Balance{
				Current:   100,
				Withdrawn: 30,
			},
			wantErr: false,
		},
		{
			name: "GetBalanceDif generates error",
			fields: fields{
				manager: mockManager,
				log:     log,
			},
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBalance generates error",
			fields: fields{
				manager: mockManager,
				log:     log,
			},
			args: args{
				ctx:    context.Background(),
				userID: 1,
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
			got, err := s.GetBalance(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBalance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_GetWithdrawals(t *testing.T) {
	log, _ := logger.CreateZapLogger("info")
	defer log.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withdrawals := []data.Withdrawal{
		{
			Order: "2377225624",
			Sum:   100.0,
		},
	}

	mockManager := mocks.NewMockBaseBonusesManager(ctrl)
	gomock.InOrder(
		mockManager.EXPECT().GetWithdrawals(gomock.Any(), gomock.Any()).Return(withdrawals, nil),
		mockManager.EXPECT().GetWithdrawals(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("manager error")),
		mockManager.EXPECT().GetWithdrawals(gomock.Any(), gomock.Any()).Return(nil, nil),
	)

	type fields struct {
		manager managers.BaseBonusesManager
		log     logger.BaseLogger
	}
	type args struct {
		ctx    context.Context
		userID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []data.Withdrawal
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
				userID: 1,
			},
			want:    withdrawals,
			wantErr: false,
		},
		{
			name: "manager returns error",
			fields: fields{
				manager: mockManager,
				log:     log,
			},
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "manager returns empty slice",
			fields: fields{
				manager: mockManager,
				log:     log,
			},
			args: args{
				ctx:    context.Background(),
				userID: 1,
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
			got, err := s.GetWithdrawals(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWithdrawals() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetWithdrawals() got = %v, want %v", got, tt.want)
			}
		})
	}
}
