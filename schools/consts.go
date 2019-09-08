package schools

// An ImportStatus describes the status of a user's imported data
type ImportStatus int

// The types of ImportStatus.
const (
	ImportStatusNone   ImportStatus = 0
	ImportStatusOK     ImportStatus = 1
	ImportStatusUpdate ImportStatus = 2
)
