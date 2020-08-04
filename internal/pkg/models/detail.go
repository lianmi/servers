package models

import "time"

type Detail struct {
	ID          uint64    `form:"id" json:"id,omitempty"`
	Name        string    `form:"name" json:"name" binding:"required"`
	Price       float32   `form:"price" json:"price" binding:"required"`
	CreatedTime time.Time `form:"created_time" json:"created_time,omitempty"`
}
