package dao

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGormUserDAO_Insert(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(t *testing.T) *sql.DB
		c       context.Context
		user    User
		wantErr error
		wantId  int64
	}{
		{
			name: "insert success",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				// 正则的写法 只要 insert 到 users 的语句就行
				mock.ExpectExec("INSERT INTO `user` .*").WillReturnError(errors.New("sql error"))
				require.NoError(t, err)
				return mockDB
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      tt.mock(t),
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
			})
			d := NewUserDAO(db)
			err = d.Insert(tt.c, tt.user)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
