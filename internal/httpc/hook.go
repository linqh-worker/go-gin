package httpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"go-gin/internal/component/logx"
	"net"
	"net/url"
)

// destination仅返回目标协议与主机，丢弃用户信息、路径、查询和片段中的凭据。
func destination(base, target string) string {
	u, err := url.Parse(target)
	if err != nil {
		return "invalid_target"
	}
	if !u.IsAbs() {
		root, err := url.Parse(base)
		if err != nil {
			return "invalid_target"
		}
		u = root.ResolveReference(u)
	}
	if u.Scheme != "http" && u.Scheme != "https" || u.Host == "" {
		return "invalid_target"
	}
	return u.Scheme + "://" + u.Host
}
func LogBeforeRequest(c *resty.Client, r *resty.Request) error {
	if r == nil {
		return nil
	}
	base := ""
	if c != nil {
		base = c.BaseURL
	}
	logx.RestyLogger.Info().Ctx(r.Context()).Str("keywords", "request").Str("destination", destination(base, r.URL)).Str("method", r.Method).Send()
	return nil
}
func LogErrorHook(r *resty.Request, err error) {
	if r == nil {
		return
	}
	reason := "request_failed"
	switch {
	case errors.Is(err, context.Canceled):
		reason = "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		reason = "deadline_exceeded"
	default:
		var network net.Error
		if errors.As(err, &network) && network.Timeout() {
			reason = "timeout"
		}
	}
	// err.Error()可能包含带凭据的完整URL或响应正文，因此仅记录错误类别。
	logx.RestyLogger.Error().Ctx(r.Context()).Str("keywords", "request_error").Str("method", r.Method).Str("reason", reason).Str("error_type", fmt.Sprintf("%T", err)).Send()
	var responseErr *resty.ResponseError
	if errors.As(err, &responseErr) && responseErr.Response != nil {
		LogResponse(r.Context(), responseErr.Response)
	}
}
func LogSuccessHook(c *resty.Client, r *resty.Response) {
	if r == nil || r.Request == nil {
		return
	}
	LogResponse(r.Request.Context(), r)
}
func LogResponse(ctx context.Context, r *resty.Response) {
	if r == nil {
		return
	}
	logx.RestyLogger.Info().Ctx(ctx).Str("keywords", "response").Int("status", r.StatusCode()).Int64("size", r.Size()).Dur("duration_ms", r.Time()).Send()
}
