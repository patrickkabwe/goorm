package goorm

const (
	DB_OPT        = "goorm:"
	DB_COL        = "db_col:"
	DB_OPT_LOOKUP = "goorm"
	DB_COL_LOOKUP = "db_col"
)

var MODEL_CONSTRAINTS_TO_OMIT = []string{"type", "column", "options", "index", "goorm", "db_col", "auto_increment"}
