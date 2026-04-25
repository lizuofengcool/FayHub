package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DemoController 演示控制器
type DemoController struct {
	DB *gorm.DB
}

// NewDemoController 创建演示控制器
func NewDemoController(db *gorm.DB) *DemoController {
	return &DemoController{DB: db}
}

// CreateUser 创建用户
func (c *DemoController) CreateUser(ctx *gin.Context) {
	// 这里可以编写创建用户的逻辑
	ctx.JSON(http.StatusOK, gin.H{
		"message": "user created successfully",
	})
}

// GetUsers 获取用户列表
func (c *DemoController) GetUsers(ctx *gin.Context) {
	// 这里可以编写获取用户列表的逻辑
	ctx.JSON(http.StatusOK, gin.H{
		"message": "users retrieved successfully",
	})
}