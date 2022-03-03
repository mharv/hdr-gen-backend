package models

type Project struct {
	Id       int32  `json:"Id,omitempty" gorm:"column:Id"`
	Number   string `json:"Number,omitempty" gorm:"column:Number"`
	Name     string `json:"Name,omitempty" gorm:"column:Name"`
	Location string `json:"Location,omitempty" gorm:"column:Location"`
	Camera   string `json:"Camera,omitempty" gorm:"column:Camera"`
}
