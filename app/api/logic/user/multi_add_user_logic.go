package user

import (
	"context"
	userRequest "go-gin/app/api/request/user"
	userResponse "go-gin/app/api/response/user"
	"go-gin/model/dao"
	"go-gin/model/do"
)

type MultiAddUserLogic struct{}

func NewMultiAddUserLogic() *MultiAddUserLogic {
	return &MultiAddUserLogic{}
}

func (l *MultiAddUserLogic) Handle(ctx context.Context, req userRequest.MultiUserAddReq) (resp *userResponse.MultiUserAddResp, err error) {
	users := make([]do.User, len(req.Users))
	for i, user := range req.Users {
		users[i] = do.User{Name: user.Name}
	}
	_, err = dao.User.Ctx(ctx).Data(users).Insert()
	if err != nil {
		return nil, err
	}
	return &userResponse.MultiUserAddResp{
		Message: "批量添加成功",
	}, nil
}
