package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicate  = errors.New("email or phone conflict")
	ErrUserNotFound   = gorm.ErrRecordNotFound
	ErrInvalidData    = gorm.ErrInvalidData
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(c context.Context, u User) error
	UpdateNonZeroFields(ctx context.Context, u User) error
	FindByEmail(c context.Context, email string) (User, error)
	FindByUserId(c context.Context, uid int64) (User, error)
	FindByUserPhone(c context.Context, phone string) (User, error)
}

type GormUserDAO struct {
	db *gorm.DB
}

type Address struct {
	Id     int64
	UserId int64
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GormUserDAO{
		db: db,
	}
}

func (ud *GormUserDAO) Insert(c context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now

	err := ud.db.WithContext(c).Create(&u).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) { // 类型断言
		const uniqueIndexErrNo uint16 = 1062
		if mysqlErr.Number == uniqueIndexErrNo {
			// 邮箱冲突 or 手机号码冲突
			return ErrUserDuplicate
		}
	}
	return nil
}

func (ud *GormUserDAO) UpdateNonZeroFields(c context.Context, u User) error {
	// 这种写法是很不清晰的，因为它依赖了 gorm 的两个默认语义
	// 会使用 ID 来作为 WHERE 条件
	// 会使用非零值来更新
	// 另外一种做法是显式指定只更新必要的字段，
	// 那么这意味着 DAO 和 service 中非敏感字段语义耦合了
	return ud.db.Updates(&u).Error
}

func (ud *GormUserDAO) FindByEmail(c context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(c).Where("email = ?", email).First(&u).Error
	return u, err
}

func (ud *GormUserDAO) FindByUserId(c context.Context, uid int64) (User, error) {
	var u User
	err := ud.db.WithContext(c).Where("id = ?", uid).First(&u).Error
	return u, err
}

func (ud *GormUserDAO) FindByUserPhone(c context.Context, phone string) (User, error) {
	var u User
	err := ud.db.WithContext(c).Where("phone = ?", phone).First(&u).Error
	return u, err
}

// User 直接对应数据库表结构
// 有人叫entity/model/PO(persistent object)
type User struct {
	Id       int64          `gorm:"primaryKey, autoIncrement" json:"id,omitempty"`
	Email    sql.NullString `gorm:"unique" json:"email,omitempty"`
	Phone    sql.NullString `gorm:"unique" json:"phone"`
	Password string         `json:"-"`
	Ctime    int64          `json:"ctime,omitempty"` // 创建时间 毫秒数
	Utime    int64          `json:"utime,omitempty"` // 更新时间 毫秒数
	NickName string         `json:"nickName,omitempty"`
	Birthday string         `json:"birthday,omitempty"`
	Bio      string         `json:"bio,omitempty"`
}
