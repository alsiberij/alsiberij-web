package models

import (
	"database/sql/driver"
	"errors"
	"time"
)

type (
	NullTime struct {
		isValid bool
		time    time.Time
	}
)

func (t *NullTime) Scan(src interface{}) error {
	switch T := src.(type) {
	case nil:
		t.isValid = false
		t.time = time.Time{}
	case time.Time:
		t.isValid = true
		t.time = T
	default:
		return errors.New("can't scan into NullTime")
	}
	return nil
}

func (t *NullTime) Value() (driver.Value, error) {
	if t.isValid {
		return nil, nil
	}
	return t.time, nil
}

func (t *NullTime) MarshalJSON() ([]byte, error) {
	if !t.isValid {
		return nil, nil
	}

	return []byte(`"` + t.time.Format("02.01.2006 15:04:05") + `"`), nil
}

func (t *NullTime) UnmarshalJSON(bytes []byte) error {
	var err error
	var v time.Time
	if bytes == nil {
		t.isValid = false
		v = time.Time{}
	} else {
		t.isValid = true
		v, err = time.Parse("02.01.2006 15:04:05", string(bytes))
	}
	t.time = v
	return err
}

func (t *NullTime) IsNull() bool {
	return !t.isValid
}

func (t *NullTime) Time() time.Time {
	return t.time
}
