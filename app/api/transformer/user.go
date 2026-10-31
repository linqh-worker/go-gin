package transformer

import (
	userResponse "go-gin/app/api/response/user"
	"go-gin/const/enum"
	"go-gin/internal/etype"
	"go-gin/model/entity"
)

func ConvertUserToListData(u []*entity.User) []userResponse.ListData {
	var resp []userResponse.ListData
	for _, v := range u {
		var ageTips string
		var age int
		if v.Age != nil {
			age = *v.Age
			if *v.Age >= 18 {
				ageTips = "成年"
			} else {
				ageTips = "未成年"
			}
		}

		var statusText string
		if v.Status != nil {
			if status, err := etype.Parse[enum.UserStatus](*v.Status); err == nil {
				statusText = status.String()
			}
		}

		resp = append(resp, userResponse.ListData{
			Id:         int(v.Id),
			Name:       v.Name,
			AgeTips:    ageTips,
			Age:        age,
			Status:     v.Status,
			StatusText: statusText,
		})
	}
	return resp
}
