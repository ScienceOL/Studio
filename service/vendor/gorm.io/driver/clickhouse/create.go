package clickhouse

import (
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
)

func (dialector *Dialector) Create(db *gorm.DB) {
	if db.Error == nil {
		if db.Statement.Schema != nil && !db.Statement.Unscoped {
			for _, c := range db.Statement.Schema.CreateClauses {
				db.Statement.AddClause(c)
			}
		}

		if db.Statement.SQL.String() == "" {
			db.Statement.SQL.Grow(180)
			db.Statement.AddClauseIfNotExists(clause.Insert{})

			if values := callbacks.ConvertToCreateValues(db.Statement); len(values.Values) >= 1 {
				prepareValues := clause.Values{
					Columns: values.Columns,
					Values:  [][]interface{}{values.Values[0]},
				}
				db.Statement.AddClause(prepareValues)
				db.Statement.Build("INSERT", "VALUES", "ON CONFLICT")

				stmt, err := db.Statement.ConnPool.PrepareContext(db.Statement.Context, db.Statement.SQL.String())
				if db.AddError(err) != nil {
					return
				}
				defer stmt.Close()

				for _, value := range values.Values {
					if _, err := stmt.Exec(value...); db.AddError(err) != nil {
						return
					}
				}
				return
			}
		}

		if !db.DryRun && db.Error == nil {
			result, err := db.Statement.ConnPool.ExecContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)

			if db.Statement.Result != nil {
				db.Statement.Result.Result = result
			}
			db.AddError(err)
		}
	}
}
