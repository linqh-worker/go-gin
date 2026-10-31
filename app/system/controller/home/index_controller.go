package home

import (
	logic "go-gin/app/system/logic/home"
	"go-gin/internal/httpx"
)

type indexController struct{}

var IndexController = &indexController{}

func (c *indexController) Ping(ctx *httpx.Context) (any, error) {
	return httpx.Handle(ctx, logic.NewPingLogic())
}
