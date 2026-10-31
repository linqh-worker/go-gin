package middleware

import (
	logic "go-gin/app/api/logic/auth"
	"go-gin/internal/httpx"
	"strings"
)

const CurrentUserKey = "api.auth.user"
const CurrentTokenKey = "api.auth.token"

func TokenCheck() httpx.HandlerFunc {
	return func(ctx *httpx.Context) (any, error) {
		parts := strings.Fields(ctx.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return nil, logic.HTTPError(logic.ErrUnauthorized)
		}
		user, err := logic.NewAuthService().Current(ctx.Request.Context(), parts[1])
		if err != nil {
			return nil, logic.HTTPError(err)
		}
		ctx.Set(CurrentUserKey, user)
		ctx.Set(CurrentTokenKey, parts[1])
		ctx.Set("user_id", user.ID)
		return nil, nil
	}
}
