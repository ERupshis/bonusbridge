// Package postgresql postgresql handling PostgreSQL database.
package postgresql

import (
	"context"
	"fmt"

	"github.com/erupshis/bonusbridge/internal/bonuses/data"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql/queries/bonuses"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql/queries/withdrawals"
	"github.com/erupshis/bonusbridge/internal/db"
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
		DBConn: dbConn,
		log:    log,
	}
}

func (p *manager) AddBonuses(ctx context.Context, userID int64, count float32) error {
	p.log.Info("[bonuses:manager:AddBonuses] start transaction for userID '%d', count '%f'", userID, count)
	errMsg := "add bonuses in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	bonusesArr, err := bonuses.Select(ctx, tx, map[string]interface{}{"user_id": userID}, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, err)
	}

	if len(bonusesArr) == 0 {
		if err = bonuses.Insert(ctx, tx, userID, count, p.log); err != nil {
			helpers.ExecuteWithLogError(tx.Rollback, p.log)
			return fmt.Errorf(errMsg, err)
		}
	} else {
		if err = bonuses.UpdateByID(ctx, tx, bonusesArr[0].ID, map[string]interface{}{"balance": bonusesArr[0].Current + count}, p.log); err != nil {
			helpers.ExecuteWithLogError(tx.Rollback, p.log)
			return fmt.Errorf(errMsg, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	p.log.Info("[bonuses:manager:AddBonuses] transaction successful")
	return nil
}

func (p *manager) GetBalance(ctx context.Context, userID int64) (*data.Balance, error) {
	p.log.Info("[bonuses:manager:GetBalance] start transaction for userID '%d'", userID)
	errMsg := "get bonuses balance in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	bonusesArr, err := bonuses.Select(ctx, tx, map[string]interface{}{"user_id": userID}, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return nil, fmt.Errorf(errMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[bonuses:manager:GetBalance] transaction successful")
	if len(bonusesArr) == 0 {
		return nil, nil
	}
	return &bonusesArr[0], nil
}

func (p *manager) WithdrawBonuses(ctx context.Context, withdrawal *data.Withdrawal) error {
	p.log.Info("[bonuses:manager:WithdrawBonuses] start transaction for withdrawal '%v'", *withdrawal)
	errMsg := "withdraw bonuses in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	bonusesArr, err := bonuses.Select(ctx, tx, map[string]interface{}{"user_id": withdrawal.UserID}, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, err)
	}

	if len(bonusesArr) == 0 {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, fmt.Errorf("bonuses record wasn't found for userID '%d'", withdrawal.UserID))
	}

	bonusesArr[0].Current -= withdrawal.Sum
	bonusesArr[0].Withdrawn += withdrawal.Sum

	valuesToUpdate := map[string]interface{}{
		"balance":   bonusesArr[0].Current,
		"withdrawn": bonusesArr[0].Withdrawn,
	}

	if err = bonuses.UpdateByID(ctx, tx, bonusesArr[0].ID, valuesToUpdate, p.log); err != nil {
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
