package models

import "time"

type Rating struct {
	ID          uint64    `form:"id" json:"id,omitempty"`
	ProductID        uint64    `form:"product_id" json:"product_id" binding:"required"`
	Score        uint32    `form:"score" json:"score" binding:"required"`
	UpdatedTime time.Time `form:"updated_time" json:"updated_time,omitempty"`
}
