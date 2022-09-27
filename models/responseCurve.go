package models

type Responsecurve struct {
	Id              int32  `json:"Id,omitempty" gorm:"column:Id"`
	FileName        string `json:"FileName,omitempty" gorm:"column:FileName"`
	DisplayName     string `json:"DisplayName,omitempty" gorm:"column:DisplayName"`
}
