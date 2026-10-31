package user

import (
	"context"
	userRequest "go-gin/app/api/request/user"
	userResponse "go-gin/app/api/response/user"
	"go-gin/app/api/transformer"
	"go-gin/internal/errorx"
	"go-gin/model/dao"
	"go-gin/model/entity"
)

type GetUsersLogic struct{}

func NewGetUsersLogic() *GetUsersLogic {
	return &GetUsersLogic{}
}

func (l *GetUsersLogic) Handle(ctx context.Context, req userRequest.ListReq) (resp *userResponse.ListResp, err error) {
	var users []*entity.User
	err = dao.User.Ctx(ctx).All(&users)
	if errorx.IsError(err) {
		return nil, err
	}

	return &userResponse.ListResp{
		Data: transformer.ConvertUserToListData(users),
	}, nil
}
