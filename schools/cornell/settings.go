package cornell

import (
	"database/sql"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

func (s *school) GetSettings(db *sql.DB, user *data.User) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (s *school) CallSettingsMethod(db *sql.DB, user *data.User, methodName string, methodParams map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"status": "error",
		"error":  "invalid_params",
	}, nil
}

func (s *school) SetSettings(db *sql.DB, user *data.User, settings map[string]interface{}) (*sql.Tx, map[string]interface{}, error) {
	return nil, nil, schools.ErrUnsupportedOperation
}
