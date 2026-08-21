package utils

import (
	"time"
)

func MonthYear(date string) (string, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", err
	}

	return t.Format("01-2006"), nil
}
