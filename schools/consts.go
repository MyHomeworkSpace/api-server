package schools

import "errors"

// An ImportStatus describes the status of a user's imported data
type ImportStatus int

// The types of ImportStatus.
const (
	ImportStatusNone   ImportStatus = 0
	ImportStatusOK     ImportStatus = 1
	ImportStatusUpdate ImportStatus = 2
)

// ErrUnsupportedOperation is returned when an operation is performed on a school that isn't supported.
var ErrUnsupportedOperation = errors.New("schools: unsupported operation")
