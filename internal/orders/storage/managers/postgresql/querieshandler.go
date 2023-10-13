package postgresql

//TODO: TO REMOVE THIS PACKAGE.
/*
// DeletePerson performs direct query request to database to delete person by id.
func (q *QueriesHandler) DeletePerson(ctx context.Context, tx *sql.Tx, id int64) (int64, error) {
	errMsg := fmt.Sprintf("delete person by id '%v' in '%s", id, PersonsTable) + ": %w"

	stmt, err := createDeletePersonStmt(ctx, tx)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, q.log)

	var result sql.Result
	query := func(context context.Context) error {
		result, err = stmt.ExecContext(context, id)
		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, q.log, []int{1, 1, 3}, databaseErrorsToRetry, query)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}

	return count, nil
}

// createDeletePersonStmt generates statement for delete query.
func createDeletePersonStmt(ctx context.Context, tx *sql.Tx) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlInsert, _, err := psql.Delete(getTableFullName(PersonsTable)).
		Where(sq.Eq{"id": "?"}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql delete statement for '"+getTableFullName(PersonsTable)+"': %w", err)

	}
	return tx.PrepareContext(ctx, psqlInsert)
}

// UpdateBonusesById generates statement for delete query.
func (q *QueriesHandler) UpdateBonusesById(ctx context.Context, tx *sql.Tx, id int64, values map[string]interface{}) (int64, error) {
	errMsg := fmt.Sprintf("update partially person by id '%d' with data '%v' in '%s'", id, values, PersonsTable) + ": %w"

	var columnsToUpdate []string
	var valuesToUpdate []interface{}
	for key, val := range values {
		columnsToUpdate = append(columnsToUpdate, key)
		valuesToUpdate = append(valuesToUpdate, val)
	}
	valuesToUpdate = append(valuesToUpdate, id)

	stmt, err := createUpdatePersonByIdStmt(ctx, tx, columnsToUpdate)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, q.log)

	var result sql.Result
	query := func(context context.Context) error {
		result, err = stmt.ExecContext(
			context,
			valuesToUpdate...,
		)
		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, q.log, []int{1, 1, 3}, databaseErrorsToRetry, query)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}

	return count, nil
}

// createUpdatePersonByIdStmt generates statement for update query.
func createUpdatePersonByIdStmt(ctx context.Context, tx *sql.Tx, values []string) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Update(getTableFullName(PersonsTable))
	for _, col := range values {
		builder = builder.Set(col, "?")
	}
	builder = builder.Where(sq.Eq{"id": "?"})
	psqlUpdate, _, err := builder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql update statement for '"+getTableFullName(PersonsTable)+"': %w", err)

	}
	return tx.PrepareContext(ctx, psqlUpdate)
}

// GetAdditionalId returns foreign key from linked table.
func (q *QueriesHandler) GetAdditionalId(ctx context.Context, tx *sql.Tx, name string, table string) (int64, error) {
	errMsg := fmt.Sprintf("get additional id for '%s' in '%s'", name, table) + ": %w"

	stmt, err := createSelectAdditionalIdStmt(ctx, tx, name, table)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, q.log)

	var id int64
	query := func(context context.Context) error {
		return stmt.QueryRowContext(ctx, name).Scan(&id)
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, q.log, []int{1, 1, 3}, databaseErrorsToRetry, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			id, err = q.InsertAdditionalId(ctx, tx, name, table)
			if err != nil {
				return 0, fmt.Errorf(errMsg, err)
			}
		} else {
			return 0, fmt.Errorf(errMsg, err)
		}
	}

	return id, nil
}

// createSelectAdditionalIdStmt generates statement for get foreign key id.
func createSelectAdditionalIdStmt(ctx context.Context, tx *sql.Tx, name string, table string) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlSelect, _, err := psql.Select("id").
		From(getTableFullName(table)).
		Where(sq.Eq{"name": name}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '"+getTableFullName(table)+"': %w", err)

	}
	return tx.PrepareContext(ctx, psqlSelect)
}

// InsertAdditionalId adds new value for linked table and returns foreign key from linked table.
func (q *QueriesHandler) InsertAdditionalId(ctx context.Context, tx *sql.Tx, name string, table string) (int64, error) {
	errMsg := fmt.Sprintf("insert additional value for '%s' in '%s'", name, table) + ": %w"

	stmt, err := createInsertAdditionalIdStmt(ctx, tx, name, table)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, q.log)

	var id int64
	query := func(context context.Context) error {
		return stmt.QueryRowContext(ctx, name).Scan(&id)
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, q.log, []int{1, 1, 3}, databaseErrorsToRetry, query)
	if err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}

	return id, nil
}

// createInsertAdditionalIdStmt generates statement for add and then get foreign key id.
func createInsertAdditionalIdStmt(ctx context.Context, tx *sql.Tx, name string, table string) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlInsert, _, err := psql.Insert(getTableFullName(table)).
		Columns("name").
		Values(name).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql insert statement for '"+getTableFullName(table)+"': %w", err)

	}
	return tx.PrepareContext(ctx, psqlInsert+"RETURNING id")
}
*/
