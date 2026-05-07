package controller

import (
	"fayhub/internal/service"
	"fayhub/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SensitiveWordController struct{}

func (ctrl *SensitiveWordController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	category := c.Query("category")
	level, _ := strconv.Atoi(c.DefaultQuery("level", "0"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := service.ServiceGroupApp.SensitiveWordService.List(c.Request.Context(), page, pageSize, keyword, category, level)
	if err != nil {
		response.GinError(c, 50000, err.Error())
		return
	}

	response.GinSuccess(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

func (ctrl *SensitiveWordController) Create(c *gin.Context) {
	var req struct {
		Word     string `json:"word" binding:"required"`
		Category string `json:"category"`
		Level    int    `json:"level"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40000, "参数错误: "+err.Error())
		return
	}

	record, err := service.ServiceGroupApp.SensitiveWordService.Create(c.Request.Context(), req.Word, req.Category, req.Level)
	if err != nil {
		response.GinError(c, 50000, err.Error())
		return
	}

	response.GinSuccess(c, record)
}

func (ctrl *SensitiveWordController) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.GinError(c, 40000, "ID格式错误")
		return
	}

	var req struct {
		Word     string `json:"word"`
		Category string `json:"category"`
		Level    int    `json:"level"`
		Status   int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40000, "参数错误: "+err.Error())
		return
	}

	record, err := service.ServiceGroupApp.SensitiveWordService.Update(c.Request.Context(), id, req.Word, req.Category, req.Level, req.Status)
	if err != nil {
		response.GinError(c, 50000, err.Error())
		return
	}

	response.GinSuccess(c, record)
}

func (ctrl *SensitiveWordController) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.GinError(c, 40000, "ID格式错误")
		return
	}

	if err := service.ServiceGroupApp.SensitiveWordService.Delete(c.Request.Context(), id); err != nil {
		response.GinError(c, 50000, err.Error())
		return
	}

	response.GinSuccess(c, nil)
}

func (ctrl *SensitiveWordController) BatchCreate(c *gin.Context) {
	var req struct {
		Words    []string `json:"words" binding:"required"`
		Category string   `json:"category"`
		Level    int      `json:"level"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40000, "参数错误: "+err.Error())
		return
	}

	count, err := service.ServiceGroupApp.SensitiveWordService.BatchCreate(c.Request.Context(), req.Words, req.Category, req.Level)
	if err != nil {
		response.GinError(c, 50000, err.Error())
		return
	}

	response.GinSuccess(c, gin.H{"count": count})
}

func (ctrl *SensitiveWordController) Rebuild(c *gin.Context) {
	if err := service.ServiceGroupApp.SensitiveWordService.RebuildMatcher(c.Request.Context()); err != nil {
		response.GinError(c, 50000, err.Error())
		return
	}

	response.GinSuccess(c, nil)
}

func (ctrl *SensitiveWordController) Check(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, 40000, "参数错误: "+err.Error())
		return
	}

	hasSensitive, words, sanitized := service.ServiceGroupApp.SensitiveWordService.Check(c.Request.Context(), req.Text)

	response.GinSuccess(c, gin.H{
		"has_sensitive": hasSensitive,
		"words":         words,
		"sanitized":     sanitized,
	})
}
