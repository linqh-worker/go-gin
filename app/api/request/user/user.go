package user

import "time"

type ListReq struct {
}

type AddUserReq struct {
	Name   string    `form:"name" binding:"required" label:"姓名"`
	Age    int       `form:"age" binding:"required" label:"年龄"`
	Status bool      `form:"status"`
	Ctime  time.Time `form:"ctime"`
}

type UserData struct {
	Id   int    `json:"id" form:"id"`
	Name string `json:"name" form:"name"`
	Age  int    `json:"age" form:"age"`
}

type MultiUserAddReq struct {
	Users []UserData `json:"users" form:"users"`
}
