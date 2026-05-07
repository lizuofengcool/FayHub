package model

type DictType struct {
	SnowflakeModel
	DictName string `json:"dict_name" gorm:"size:100;not null"`
	DictType string `json:"dict_type" gorm:"size:100;not null;uniqueIndex"`
	Status   int    `json:"status" gorm:"default:1"`
	Remark   string `json:"remark" gorm:"size:255"`
}

func (DictType) TableName() string {
	return "dict_types"
}

type DictData struct {
	SnowflakeModel
	DictType  string `json:"dict_type" gorm:"size:100;not null;index"`
	DictLabel string `json:"dict_label" gorm:"size:100;not null"`
	DictValue string `json:"dict_value" gorm:"size:100;not null"`
	CssClass  string `json:"css_class" gorm:"size:100"`
	ListClass string `json:"list_class" gorm:"size:100"`
	IsDefault int    `json:"is_default" gorm:"default:0"`
	Sort      int    `json:"sort" gorm:"default:0"`
	Status    int    `json:"status" gorm:"default:1"`
	Remark    string `json:"remark" gorm:"size:255"`
}

func (DictData) TableName() string {
	return "dict_data"
}
