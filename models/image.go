package models

type Image struct {
	Id        int32  `json:"Id,omitempty" gorm:"column:Id"`
	ProjectId int32  `json:"ProjectId,omitempty" gorm:"column:ProjectId"`
	Name      string `json:"Name,omitempty" gorm:"column:Name"`
	Type      string `json:"Type,omitempty" gorm:"column:Type"`
}
