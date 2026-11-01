package config

import (
	auth "go-gin/app/system/logic/auth"
	logic "go-gin/app/system/logic/config"
	request "go-gin/app/system/request/config"
	"go-gin/internal/errorx"
	"go-gin/internal/httpx"
	"go-gin/internal/httpx/validation"
	"strconv"
)

func List(c *httpx.Context) (any, error) {
	c.Header("Cache-Control", "no-store")
	rows, err := logic.Groups(c.Request.Context())
	return rows, auth.HTTPError(err)
}
func Save(c *httpx.Context) (any, error) {
	var r request.SaveReq
	if err := c.ShouldBindJSON(&r); err != nil {
		return nil, validation.Error(err, &r)
	}
	return nil, auth.HTTPError(logic.Save(c.Request.Context(), c.Param("group"), c.GetInt("user_id"), r))
}

func CreateGroup(c *httpx.Context) (any, error) {
	var r request.GroupReq
	if err := c.ShouldBindJSON(&r); err != nil {
		return nil, validation.Error(err, &r)
	}
	return nil, auth.HTTPError(logic.CreateGroup(c.Request.Context(), c.GetInt("user_id"), r))
}
func UpdateGroup(c *httpx.Context) (any, error) {
	var r request.GroupReq
	if err := c.ShouldBindJSON(&r); err != nil {
		return nil, validation.Error(err, &r)
	}
	return nil, auth.HTTPError(logic.UpdateGroup(c.Request.Context(), c.Param("group"), c.GetInt("user_id"), r))
}
func DeleteGroup(c *httpx.Context) (any, error) {
	var r request.VersionReq
	if err := c.ShouldBindJSON(&r); err != nil {
		return nil, validation.Error(err, &r)
	}
	return nil, auth.HTTPError(logic.DeleteGroup(c.Request.Context(), c.Param("group"), r.Version))
}
func definitionID(c *httpx.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errorx.New(10001, "ID必须为正整数")
	}
	return id, nil
}
func SaveDefinition(c *httpx.Context) (any, error) {
	var id int64
	var err error
	if c.Param("id") != "" {
		id, err = definitionID(c)
		if err != nil {
			return nil, err
		}
	}
	var r request.DefinitionReq
	if err := c.ShouldBindJSON(&r); err != nil {
		return nil, validation.Error(err, &r)
	}
	return nil, auth.HTTPError(logic.SaveDefinition(c.Request.Context(), id, c.GetInt("user_id"), r))
}
func DeleteDefinition(c *httpx.Context) (any, error) {
	id, err := definitionID(c)
	if err != nil {
		return nil, err
	}
	var r request.VersionReq
	if err := c.ShouldBindJSON(&r); err != nil {
		return nil, validation.Error(err, &r)
	}
	return nil, auth.HTTPError(logic.DeleteDefinition(c.Request.Context(), id, r.Version, c.GetInt("user_id")))
}
