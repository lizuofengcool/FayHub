package model

type SensitiveWord struct {
	TenantModel
	Word     string `json:"word" gorm:"size:200;not null;index"`
	Category string `json:"category" gorm:"size:50;default:''"`
	Level    int    `json:"level" gorm:"default:1"`
	Status   int    `json:"status" gorm:"default:1"`
}

func (SensitiveWord) TableName() string {
	return "sensitive_words"
}
