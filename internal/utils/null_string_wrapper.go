package utils

import (
	"database/sql"
	"encoding/json"
	"log"
)

type NullStringWrapper struct {
	sql.NullString
}

func (s NullStringWrapper) MarshalJSON() ([]byte, error) {
	if s.Valid {
		return json.Marshal(s.String)
	}
	return []byte(`null`), nil
}

func (s NullStringWrapper) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		log.Println(err)
		return err
	}
	if len(str) == 0 {
		s.Valid = false
		return nil
	}
	s.Valid = true
	s.String = str
	return nil
}
