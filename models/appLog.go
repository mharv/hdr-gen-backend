package models

type Applog struct {
    Id              int32  `json:"Id,omitempty" gorm:"column:Id"`
    ProjectId       int32  `json:"ProjectId,omitempty" gorm:"column:ProjectId"`
    ImageId         int32  `json:"ImageId,omitempty" gorm:"column:ImageId"`
    Time            string `json:"Time,omitempty" gorm:"column:Time"`
    Message         string `json:"Message,omitempty" gorm:"column:Message"`
}
