// Package postgresql postgresql handling PostgreSQL database.
package postgresql

import (
	"context"
	"fmt"

	"github.com/erupshis/bonusbridge/internal/bonuses/data"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/db/queries/bonuses"
	"github.com/erupshis/bonusbridge/internal/db/queries/withdrawals"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// manager storageManager implementation for PostgreSQL
// All requests are done on united transaction. Multi insert/update/delete is not supported at the moment.
type manager struct {
	*db.Conn

	log logger.BaseLogger
}

// Create creates manager implementation. Supports migrations and check connection to database.
func Create(dbConn *db.Conn, log logger.BaseLogger) managers.BaseBonusesManager {
	return &manager{
		Conn: dbConn,
		log:  log,
	}
}

func (p *manager) GetBalanceDif(ctx context.Context, userID int64) (float32, error) {
	p.log.Info("[bonuses:manager:GetBalanceDif] start transaction for userID '%d'", userID)
	errMsg := "get bonuses balance in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return -1.0, fmt.Errorf(errMsg, err)
	}

	bonusesDif, err := bonuses.SelectSumByUserID(ctx, tx, bonuses.SumTotal, userID, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return -1.0, fmt.Errorf(errMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return -1.0, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[bonuses:manager:GetBalanceDif] transaction successful")
	return bonusesDif, nil
}

func (p *manager) GetBalance(ctx context.Context, income bool, userID int64) (float32, error) {
	p.log.Info("[bonuses:manager:GetBalance] start transaction for userID '%d' for income? '%t'", userID, income)
	errMsg := "get bonuses income sum in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return -1.0, fmt.Errorf(errMsg, err)
	}

	var filter int
	if income {
		filter = bonuses.SumIn
	} else {
		filter = bonuses.SumOut
	}

	bonusesIncome, err := bonuses.SelectSumByUserID(ctx, tx, filter, userID, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return -1.0, fmt.Errorf(errMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return -1.0, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[bonuses:manager:GetBalance] transaction successful")
	return bonusesIncome, nil
}

func (p *manager) WithdrawBonuses(ctx context.Context, withdrawal *data.Withdrawal) error {
	p.log.Info("[bonuses:manager:WithdrawBonuses] start transaction for withdrawal '%v'", *withdrawal)
	errMsg := "withdraw bonuses in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	bonusesDif, err := bonuses.SelectSumByUserID(ctx, tx, bonuses.SumTotal, withdrawal.UserID, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, err)
	}

	if bonusesDif < withdrawal.Sum {
		return fmt.Errorf("userID '%d' balance '%f' is not enough for withdrawn: %w", withdrawal.UserID, bonusesDif, data.ErrNotEnoughBonuses)
	}

	withdrawal.BonusID, err = bonuses.Insert(ctx, tx, withdrawal.UserID, -withdrawal.Sum, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, err)
	}

	if err = withdrawals.Insert(ctx, tx, withdrawal, p.log); err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	p.log.Info("[bonuses:manager:WithdrawBonuses] transaction successful")
	return nil
}

func (p *manager) GetWithdrawals(ctx context.Context, userID int64) ([]data.Withdrawal, error) {
	p.log.Info("[bonuses:manager:GetWithdrawals] start transaction for userID '%d'", userID)
	errMsg := "get withdrawals from db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	withdrawalsArr, err := withdrawals.Select(ctx, tx, map[string]interface{}{"user_id": userID}, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return nil, fmt.Errorf(errMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[bonuses:manager:GetWithdrawals] transaction successful")
	return withdrawalsArr, nil
}
