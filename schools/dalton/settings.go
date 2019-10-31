package dalton

import (
	"database/sql"

	"github.com/MyHomeworkSpace/api-server/data"
)

func (s *school) GetSettings(db *sql.DB, user *data.User) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (s *school) SetSettings(db *sql.DB, user *data.User, settings map[string]interface{}) error {
	return nil
}
