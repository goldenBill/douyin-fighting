package global

import (
	"github.com/jmoiron/sqlx"
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
)

type JWT struct {
	SigningKey string
}

var (
	GVAR_SQLX_DB      *sqlx.DB
	GVAR_DB           *gorm.DB
	GVAR_JWT          JWT
	GVAR_ID_GENERATOR *sonyflake.Sonyflake
)
