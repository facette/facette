package orm

import "errors"

var (
	// ErrConstraintForeignKey represents a foreign key constraint violation error.
	ErrConstraintForeignKey = errors.New("foreign key constraint violation")
	// ErrConstraintNotNull represents a not null constraint violation error.
	ErrConstraintNotNull = errors.New("not null constraint violation")
	// ErrConstraintUnique represents a primary key or unique constraint violation error.
	ErrConstraintUnique = errors.New("primary key or unique constraint violation")
	// ErrEmptyScan represents an empty row scan error.
	ErrEmptyScan = errors.New("scan on empty row")
	// ErrInvalidScanValue represents an invalid scan value error.
	ErrInvalidScanValue = errors.New("invalid scan value")
	// ErrInvalidStruct represents an invalid struct error.
	ErrInvalidStruct = errors.New("invalid struct")
	// ErrMissingPrimaryKey represents a missing primary key error.
	ErrMissingPrimaryKey = errors.New("missing primary key")
	// ErrMissingTransaction represents a missing transaction error.
	ErrMissingTransaction = errors.New("missing transaction")
	// ErrNotConvertible represents an non-convertible column value error.
	ErrNotConvertible = errors.New("unconvertible value")
	// ErrNotScanable represents an unscannable column value error.
	ErrNotScanable = errors.New("unscannable value")
	// ErrUnsupportedDriver represents an unsupported database driver error.
	ErrUnsupportedDriver = errors.New("unsupported driver")
	// ErrUnsupportedType represents an unsupported column type error.
	ErrUnsupportedType = errors.New("unsupported type")
)
