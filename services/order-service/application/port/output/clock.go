package output

import "time"

type Clock interface {
	Now() time.Time
}
