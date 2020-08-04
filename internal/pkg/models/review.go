package models

import "time"

type Review struct {
	ID          uint64    `form:"id" json:"id,omitempty"`
	ProductID        uint64    `form:"product_id" json:"product_id" binding:"required"`
	Message        string    `form:"message" json:"message" binding:"required"`
	CreatedTime time.Time `form:"created_time" json:"created_time,omitempty"`
}

