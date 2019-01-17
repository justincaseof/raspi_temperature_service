package model

import "time"

type Measurement struct {
	Id         int64 `xorm:"pk not null autoincr"`
	Value      float32
	Unit       string
	InstanceId string    `xorm:"varchar(200)"`
	Created    time.Time `xorm:"created"`
}
