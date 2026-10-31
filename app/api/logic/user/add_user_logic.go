package user

import (
	"context"
	"fmt"
	userRequest "go-gin/app/api/request/user"
	userResponse "go-gin/app/api/response/user"
	"go-gin/model/dao"
	"go-gin/model/do"
)

type AddUserLogic struct{}

func NewAddUserLogic() *AddUserLogic {
	return &AddUserLogic{}
}

func (l *AddUserLogic) Handle(ctx context.Context, req userRequest.AddUserReq) (resp *userResponse.AddUserResp, err error) {
	id, err := dao.User.Ctx(ctx).Data(do.User{Name: req.Name}).InsertAndGetId()
	if err != nil {
		return nil, err
	}
	resp = &userResponse.AddUserResp{
		Message: fmt.Sprintf("message:%d", id),
	}
	return
}
