package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicateEmail = errors.New("email conflict")
	ErrUserNotFound       = gorm.ErrRecordNotFound
	ErrInvalidData        = gorm.ErrInvalidData
	ErrRecordNotFound     = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

type Address struct {
	Id     int64
	UserId int64
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (ud *UserDAO) Insert(c context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now

	err := ud.db.WithContext(c).Create(&u).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) { // 类型断言
		const uniqueIndexErrNo uint16 = 1062
		if mysqlErr.Number == uniqueIndexErrNo {
			// 邮箱冲突（邮箱在此是唯一索引）
			return ErrUserDuplicateEmail
		}
	}
	return nil
}

func (ud *UserDAO) Update(c context.Context, uid interface{}, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now

	err := ud.db.WithContext(c).Where("id = ?", uid).Updates(u).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) { // 类型断言
		const uniqueIndexErrNo uint16 = 1062
		if mysqlErr.Number == uniqueIndexErrNo {
			// 邮箱冲突（邮箱在此是唯一索引）
			return ErrUserDuplicateEmail
		}
	}
	return nil
}

func (ud *UserDAO) FindByEmail(c context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(c).Where("email = ?", email).First(&u).Error
	return u, err
}

func (ud *UserDAO) FindByUserId(c context.Context, uid int64) (User, error) {
	var u User
	err := ud.db.WithContext(c).Where("id = ?", uid).First(&u).Error
	return u, err
}

// User 直接对应数据库表结构
// 有人叫entity/model/PO(persistent object)
type User struct {
	Id       int64  `gorm:"primaryKey, autoIncrement" json:"id,omitempty"`
	Email    string `gorm:"unique" json:"email,omitempty"`
	Password string `json:"-"`
	Ctime    int64  `json:"ctime,omitempty"` // 创建时间 毫秒数
	Utime    int64  `json:"utime,omitempty"` // 更新时间 毫秒数
	NickName string `json:"nickName,omitempty"`
	Birthday string `json:"birthday,omitempty"`
	Bio      string `json:"bio,omitempty"`
}
