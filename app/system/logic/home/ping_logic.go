package home

import (
	"context"
	homeResponse "go-gin/app/system/response/home"
)

type PingLogic struct{}

func NewPingLogic() *PingLogic { return &PingLogic{} }

func (l *PingLogic) Handle(ctx context.Context, req struct{}) (*homeResponse.PingResp, error) {
	return &homeResponse.PingResp{Module: "system", Message: "pong"}, nil
}
