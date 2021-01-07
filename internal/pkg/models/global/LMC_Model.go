package global

import (
	"gorm.io/gorm"
	// "time"
)

type LMC_Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt int64
	UpdatedAt int64
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
