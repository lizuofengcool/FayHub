package model

type ErrorCode struct {
	SnowflakeModel
	Code   int    `json:"code" gorm:"uniqueIndex;not null"`
	Name   string `json:"name" gorm:"size:100;not null"`
	Msg    string `json:"msg" gorm:"size:500;not null"`
	Status int    `json:"status" gorm:"default:1"`
}

func (ErrorCode) TableName() string {
	return "error_codes"
}
