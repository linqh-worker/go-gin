package account

import (
	"context"
	"go-gin/app/system/dao"
	request "go-gin/app/system/request/account"
	response "go-gin/app/system/response/account"
	"go-gin/internal/errorx"
	"go-gin/util"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

func profile(u *dao.AuthUser) response.Profile {
	return response.Profile{ID: u.ID, Username: u.Name, Nickname: u.Nickname, AvatarURL: u.AvatarURL, Phone: u.Phone, Email: u.Email, Age: u.Age, Gender: u.Gender, Status: u.Status, LoginCount: u.LoginCount, LastLoginAt: util.FormatTime(u.LastLoginAt), CreatedAt: util.FormatTime(u.CreatedAt), UpdatedAt: util.FormatTime(u.UpdatedAt)}
}
func List(ctx context.Context, admin bool, r request.ListReq) (response.Page, error) {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PageSize == 0 {
		r.PageSize = 20
	}
	out := response.Page{Items: []response.Profile{}, Page: r.Page, PageSize: r.PageSize}
	rows, total, err := dao.ListAccounts(ctx, admin, r.Page, r.PageSize, strings.TrimSpace(r.Keyword), r.Status)
	if err != nil {
		return out, err
	}
	out.Total = total
	for i := range rows {
		out.Items = append(out.Items, profile(&rows[i]))
	}
	return out, nil
}
func Detail(ctx context.Context, admin bool, id int) (response.Profile, error) {
	u, err := dao.Account(ctx, admin, id)
	if err != nil {
		return response.Profile{}, err
	}
	return profile(u), nil
}
func Create(ctx context.Context, admin bool, r request.CreateReq) (response.Profile, error) {
	r.Username = strings.ToLower(strings.TrimSpace(r.Username))
	r.Nickname = strings.TrimSpace(r.Nickname)
	if !regexp.MustCompile(`^[a-z0-9_]{3,32}$`).MatchString(r.Username) {
		return response.Profile{}, errorx.New(10001, "用户名须为3至32位字母数字下划线")
	}
	if utf8.RuneCountInString(r.Password) < 8 || len(r.Password) > 72 {
		return response.Profile{}, errorx.New(10001, "密码至少8个字符且不超过72字节")
	}
	if r.Nickname == "" {
		r.Nickname = r.Username
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.Profile{}, err
	}
	now := time.Now()
	empty := ""
	age := 0
	u := dao.AuthUser{Name: r.Username, Nickname: r.Nickname, PasswordHash: string(hash), Status: 1, Phone: &empty, Email: &empty, Age: &age, PasswordChangedAt: &now, CreatedAt: &now, UpdatedAt: &now}
	if err := dao.CreateAccount(ctx, admin, &u); err != nil {
		return response.Profile{}, err
	}
	return profile(&u), nil
}
func Update(ctx context.Context, admin bool, id int, r request.UpdateReq) error {
	fields := map[string]any{}
	if r.Nickname != nil {
		v := strings.TrimSpace(*r.Nickname)
		if v == "" {
			return errorx.New(10001, "昵称不能为空")
		}
		fields["nickname"] = v
	}
	if r.AvatarURL != nil {
		fields["avatar_url"] = *r.AvatarURL
	}
	if r.Phone != nil {
		fields["phone"] = *r.Phone
	}
	if r.Email != nil {
		fields["email"] = *r.Email
	}
	if r.Age != nil {
		fields["age"] = *r.Age
	}
	if r.Gender != nil {
		fields["gender"] = *r.Gender
	}
	return dao.UpdateAccount(ctx, admin, id, fields, false)
}
func Status(ctx context.Context, admin bool, id, actor int, status int) error {
	if admin && id == actor {
		return errorx.New(10001, "不能通过此接口修改自己的状态")
	}
	return dao.UpdateAccount(ctx, admin, id, map[string]any{"status": status}, true)
}
