package domain

import (
	"time"
)

// User 领域对象， 是DDD中的聚合根/entity
type User struct {
	Id       int64
	Email    string
	Phone    string
	Password string
	Ctime    time.Time
	NickName string
	Birthday time.Time
	Bio      string
}
