package output

import "time"

type TokenIssuer interface {
	Issue(userID, role string, now time.Time) (token string, expiresAt time.Time, err error)
}
