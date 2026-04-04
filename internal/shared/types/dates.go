package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/thalalhassan/edu_management/internal/constants"
)

type Date struct {
	time.Time
}

func (d *Date) Convert() string {
	return d.Format(constants.DateLayout)
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	t, err := time.Parse(constants.DateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}

	d.Time = t
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format(constants.DateLayout) + `"`), nil
}
