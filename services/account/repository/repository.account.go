package repository_account

import (
	"context"
	"database/sql"
	"fmt"
	"golang-backend-example/domain"
	"golang-backend-example/util"
)

type accountRepository struct {
	dbPool *sql.DB
	txPool *sql.Tx
}

func (c accountRepository) List(ctx context.Context, param domain.AccountListParam) ([]domain.Account, error) {
	var filters []string

	if param.Search != "" {
		filters = append(filters,
			fmt.Sprintf(`(LOWER(account_number) LIKE '%%%s%%' OR LOWER(customer_number) LIKE '%%%s%%')`,
				param.Search, param.Search))
	}

	filterQuery := util.BuildFilterQuery(filters)

	stmt, err := c.dbPool.Prepare(fmt.Sprintf(`
		SELECT
			account_number,
			customer_number,
			balance
		FROM account
			%s
		ORDER BY account_number %s
		LIMIT $1 OFFSET $2
	`, filterQuery, param.Order))
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, param.Limit, param.Offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var accounts []domain.Account
	for rows.Next() {
		var account domain.Account
		err := rows.Scan(
			&account.AccountNumber,
			&account.CustomerNumber,
			&account.Balance,
		)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (c accountRepository) GetByAccountNumber(ctx context.Context, accountNumber int) (domain.Account, error) {
	stmt, err := c.dbPool.Prepare(fmt.Sprintf(`
		SELECT
			account_number,
			customer_number,
			balance
		FROM account
		WHERE
			account_number = $1
	`))
	if err != nil {
		return domain.Account{}, err
	}

	var account domain.Account
	err = stmt.QueryRowContext(ctx, accountNumber).Scan(
		&account.AccountNumber,
		&account.CustomerNumber,
		&account.Balance,
	)
	if err != nil {
		return domain.Account{}, err
	}

	return account, nil
}

func (c accountRepository) Store(ctx context.Context, a *domain.Account) error {
	stmt, err := c.dbPool.Prepare(fmt.Sprintf(`
		INSERT INTO account (
			account_number,
			customer_number,
			balance
		) VALUES (
			$1, $2, $3
		)`))
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx,
		&a.AccountNumber,
		&a.CustomerNumber,
		&a.Balance)
	if err != nil {
		return err
	}

	return nil
}

func (c accountRepository) Update(ctx context.Context, a *domain.Account) error {
	stmt, err := c.dbPool.Prepare(fmt.Sprintf(`
		UPDATE account SET
			customer_number = $1,
			balance = $2
		WHERE
			account_number = $3
	`))

	_, err = stmt.ExecContext(ctx,
		a.CustomerNumber,
		a.Balance,
		a.AccountNumber)
	if err != nil {
		return err
	}

	return nil
}

func (c accountRepository) Delete(ctx context.Context, a *domain.Account) error {

	stmt, err := c.dbPool.Prepare(fmt.Sprintf(`
		DELETE FROM account
		WHERE
			account_number = $1
	`))
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx,
		a.AccountNumber)
	if err != nil {
		return err
	}

	return nil
}

func NewAccountRepository(db *sql.DB) domain.AccountRepository {
	return &accountRepository{
		dbPool: db,
	}
}
