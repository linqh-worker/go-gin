package user

type ListData struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	AgeTips    string `json:"age_tips"`
	Age        int    `json:"age"`
	Status     *int   `json:"status"`
	StatusText string `json:"status_text"`
}

type ListResp struct {
	Data []ListData `json:"data"`
}

type AddUserResp struct {
	Message string `json:"message"`
}

type MultiUserAddResp struct {
	Message string `json:"message"`
}
