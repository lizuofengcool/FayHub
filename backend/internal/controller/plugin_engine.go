// Copyright (c) 2026 FayHub Team
// SPDX-License-Identifier: MIT

package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fayhub/internal/model"
	"fayhub/internal/service"
	"fayhub/pkg/errors"
	"fayhub/pkg/market"
	"fayhub/pkg/response"
	"fayhub/pkg/utils"

	"github.com/gin-gonic/gin"
)

type PluginEngineController struct{}

func validatePluginID(c *gin.Context) string {
	pluginID := c.Param("id")
	if !utils.ValidateCUID(pluginID) {
		response.GinError(c, errors.ErrParamValidation, "无效的插件ID格式")
		return ""
	}
	return pluginID
}

// ListPlugins godoc
// @Summary 列出已安装插件
// @Description 获取当前租户已安装的所有插件列表
// @Tags 插件引擎
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "获取成功"
// @Router /api/plugin-engine/plugins [get]
func (pec *PluginEngineController) ListPlugins(c *gin.Context) {
	ctx := c.Request.Context()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	plugins, total, err := service.ServiceGroupApp.PluginEngineService.ListPlugins(ctx, page, pageSize)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, gin.H{"list": plugins, "total": total, "page": page, "page_size": pageSize})
}

// GetPlugin godoc
// @Summary 获取插件详情
// @Description 获取指定已安装插件的详细信息
// @Tags 插件引擎
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 404 {object} response.Response "插件不存在"
// @Router /api/plugin-engine/plugins/{id} [get]
func (pec *PluginEngineController) GetPlugin(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}
	ctx := c.Request.Context()

	plugin, err := service.ServiceGroupApp.PluginEngineService.GetPlugin(ctx, pluginID)
	if err != nil {
		response.GinError(c, errors.ErrResourceNotFound, err.Error())
		return
	}

	response.GinSuccess(c, plugin)
}

// InstallCallback godoc
// @Summary 插件安装回调
// @Description 从市场跳转回底座的安装回调接口
// @Tags 插件引擎
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.InstallPluginRequest true "安装请求"
// @Success 200 {object} response.Response "安装成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /api/plugin-engine/install-callback [post]
func (pec *PluginEngineController) InstallCallback(c *gin.Context) {
	var req service.InstallPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errors.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.ServiceGroupApp.PluginEngineService.InstallPlugin(ctx, &req); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "插件安装成功", nil)
}

// UninstallPlugin godoc
// @Summary 卸载插件
// @Description 卸载指定已安装插件
// @Tags 插件引擎
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Success 200 {object} response.Response "卸载成功"
// @Failure 404 {object} response.Response "插件不存在"
// @Router /api/plugin-engine/plugins/{id} [delete]
func (pec *PluginEngineController) UninstallPlugin(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}
	ctx := c.Request.Context()

	if err := service.ServiceGroupApp.PluginEngineService.UninstallPlugin(ctx, pluginID); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "插件卸载成功", nil)
}

// EnablePlugin godoc
// @Summary 启用插件
// @Description 启用指定已安装插件
// @Tags 插件引擎
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Success 200 {object} response.Response "启用成功"
// @Failure 404 {object} response.Response "插件不存在"
// @Router /api/plugin-engine/plugins/{id}/enable [put]
func (pec *PluginEngineController) EnablePlugin(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}
	ctx := c.Request.Context()

	if err := service.ServiceGroupApp.PluginEngineService.EnablePlugin(ctx, pluginID); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "插件启用成功", nil)
}

// DisablePlugin godoc
// @Summary 禁用插件
// @Description 禁用指定已安装插件
// @Tags 插件引擎
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Success 200 {object} response.Response "禁用成功"
// @Failure 404 {object} response.Response "插件不存在"
// @Router /api/plugin-engine/plugins/{id}/disable [put]
func (pec *PluginEngineController) DisablePlugin(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}
	ctx := c.Request.Context()

	if err := service.ServiceGroupApp.PluginEngineService.DisablePlugin(ctx, pluginID); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "插件禁用成功", nil)
}

// UpgradePlugin godoc
// @Summary 升级插件
// @Description 升级指定已安装插件到新版本
// @Tags 插件引擎
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Param request body object true "升级请求"
// @Success 200 {object} response.Response "升级成功"
// @Failure 404 {object} response.Response "插件不存在"
// @Router /api/plugin-engine/plugins/{id}/upgrade [put]
func (pec *PluginEngineController) UpgradePlugin(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}

	var req struct {
		NewVersion    string `json:"new_version"`
		NewLicenseKey string `json:"new_license_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errors.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.ServiceGroupApp.PluginEngineService.UpgradePlugin(ctx, pluginID, req.NewVersion, req.NewLicenseKey); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "插件升级成功", nil)
}

// GetPluginConfig godoc
// @Summary 获取插件配置
// @Description 获取指定插件的配置信息
// @Tags 插件引擎
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 404 {object} response.Response "插件不存在"
// @Router /api/plugin-engine/plugins/{id}/config [get]
func (pec *PluginEngineController) GetPluginConfig(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}
	ctx := c.Request.Context()

	config, err := service.ServiceGroupApp.PluginEngineService.GetPluginConfig(ctx, pluginID)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, config)
}

// UpdatePluginConfig godoc
// @Summary 更新插件配置
// @Description 更新指定插件的配置信息
// @Tags 插件引擎
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Param request body map[string]interface{} true "配置信息"
// @Success 200 {object} response.Response "更新成功"
// @Failure 404 {object} response.Response "插件不存在"
// @Router /api/plugin-engine/plugins/{id}/config [put]
func (pec *PluginEngineController) UpdatePluginConfig(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}

	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		response.GinError(c, errors.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.ServiceGroupApp.PluginEngineService.UpdatePluginConfig(ctx, pluginID, config); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "插件配置更新成功", nil)
}

func (pec *PluginEngineController) GetPluginPage(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}
	ctx := c.Request.Context()

	pageData, err := service.ServiceGroupApp.PluginEngineService.GetPluginPage(ctx, pluginID)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, pageData)
}

func (pec *PluginEngineController) InstallDemoPlugin(c *gin.Context) {
	ctx := c.Request.Context()

	if err := service.ServiceGroupApp.PluginEngineService.InstallDemoPlugin(ctx); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "示例插件安装成功", nil)
}

func (pec *PluginEngineController) SearchMarketPlugins(c *gin.Context) {
	keyword := c.Query("keyword")
	page := 1
	pageSize := 9
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	categoryID := c.Query("category_id")

	ctx := c.Request.Context()
	result, err := service.ServiceGroupApp.PluginEngineService.SearchMarketPlugins(ctx, keyword, page, pageSize, categoryID)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, result)
}

func (pec *PluginEngineController) GetMarketPluginDetail(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}
	ctx := c.Request.Context()

	result, err := service.ServiceGroupApp.PluginEngineService.GetMarketPluginDetail(ctx, pluginID)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, result)
}

func (pec *PluginEngineController) GetMarketCategories(c *gin.Context) {
	ctx := c.Request.Context()

	result, err := service.ServiceGroupApp.PluginEngineService.GetMarketCategories(ctx)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, result)
}

func (pec *PluginEngineController) InstallFromMarket(c *gin.Context) {
	var req struct {
		MarketPluginID string `json:"market_plugin_id"`
		TargetVersion  string `json:"target_version"`
		LicenseKey     string `json:"license_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errors.ErrParamValidation, "参数错误")
		return
	}

	if !utils.ValidateCUID(req.MarketPluginID) {
		response.GinError(c, errors.ErrParamValidation, "无效的市场插件ID格式")
		return
	}

	userID, _ := c.Get("user_id")
	ctx := c.Request.Context()
	if err := service.ServiceGroupApp.PluginEngineService.InstallFromMarket(ctx, req.MarketPluginID, req.TargetVersion, req.LicenseKey, fmt.Sprintf("%v", userID)); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "从市场安装插件成功", nil)
}

func (pec *PluginEngineController) GetVersionHistory(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	ctx := c.Request.Context()
	histories, total, err := service.ServiceGroupApp.PluginEngineService.GetVersionHistory(ctx, pluginID, page, pageSize)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, gin.H{
		"list":  histories,
		"total": total,
		"page":  page,
	})
}

func (pec *PluginEngineController) RollbackPlugin(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}
	var req struct {
		TargetVersion string `json:"target_version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errors.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.ServiceGroupApp.PluginEngineService.RollbackPlugin(ctx, pluginID, req.TargetVersion); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "回滚成功", nil)
}

func (pec *PluginEngineController) CheckForUpdates(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}

	ctx := c.Request.Context()
	update, err := service.ServiceGroupApp.PluginEngineService.CheckForUpdates(ctx, pluginID)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	if update == nil {
		response.GinSuccess(c, gin.H{"has_update": false})
		return
	}

	response.GinSuccess(c, gin.H{
		"has_update":     true,
		"latest_version": update.Version,
		"changelog":      update.Changelog,
		"wasm_url":       update.WasmURL,
		"manifest_url":   update.ManifestURL,
	})
}

func (pec *PluginEngineController) CheckAllUpdates(c *gin.Context) {
	ctx := c.Request.Context()
	updates, err := service.ServiceGroupApp.PluginEngineService.CheckAllUpdates(ctx)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	if updates == nil {
		updates = []map[string]interface{}{}
	}

	response.GinSuccess(c, gin.H{
		"updates": updates,
		"count":   len(updates),
	})
}

func (pec *PluginEngineController) GetDependencies(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}

	ctx := c.Request.Context()
	deps, err := service.ServiceGroupApp.PluginEngineService.GetDependencies(ctx, pluginID)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	if deps == nil {
		deps = []*model.PluginDependency{}
	}

	response.GinSuccess(c, deps)
}

func (pec *PluginEngineController) SaveDependencies(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}
	var req struct {
		Dependencies []map[string]interface{} `json:"dependencies" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errors.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.ServiceGroupApp.PluginEngineService.SaveDependencies(ctx, pluginID, req.Dependencies); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "保存依赖成功", nil)
}

func (pec *PluginEngineController) ValidateDependencies(c *gin.Context) {
	pluginID := validatePluginID(c)
	if pluginID == "" {
		return
	}

	ctx := c.Request.Context()
	issues, err := service.ServiceGroupApp.PluginEngineService.ValidateDependencies(ctx, pluginID)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	if issues == nil {
		issues = []map[string]interface{}{}
	}

	response.GinSuccess(c, gin.H{
		"valid":  len(issues) == 0,
		"issues": issues,
	})
}

func (pec *PluginEngineController) GetMarketPublicKey(c *gin.Context) {
	ctx := c.Request.Context()

	client := market.GetClient()
	if client != nil {
		publicKey, err := client.GetPublicKey(ctx)
		if err == nil && publicKey != "" {
			response.GinSuccess(c, gin.H{
				"public_key": publicKey,
			})
			return
		}
	}

	publicKeyPem := os.Getenv("FAYHUB_MARKET_PUBLIC_KEY")
	if publicKeyPem == "" {
		response.GinError(c, errors.ErrInternalServer, "市场公钥不可用")
		return
	}

	response.GinSuccess(c, gin.H{
		"public_key": strings.TrimSpace(publicKeyPem),
	})
}

func (pec *PluginEngineController) ServePluginAsset(c *gin.Context) {
	pluginID := c.Param("pluginId")
	if !utils.ValidateCUID(pluginID) {
		c.Status(http.StatusBadRequest)
		return
	}
	filePath := c.Param("filepath")

	if filePath == "" || filePath == "/" {
		c.Status(http.StatusNotFound)
		return
	}

	filePath = strings.TrimPrefix(filePath, "/")

	if strings.Contains(filePath, "..") {
		c.Status(http.StatusForbidden)
		return
	}

	searchPaths := []string{
		filepath.Join("plugins", "assets", pluginID, filePath),
		filepath.Join("data", "plugins", pluginID, filePath),
		filepath.Join("plugins", pluginID, "frontend", filePath),
	}

	execDir, _ := os.Getwd()
	projectRoot := os.Getenv("FAYHUB_PROJECT_ROOT")
	if projectRoot == "" {
		projectRoot = execDir
	}

	absSearchPaths := make([]string, 0, len(searchPaths)+2)
	for _, p := range searchPaths {
		absSearchPaths = append(absSearchPaths, p)
		if !filepath.IsAbs(p) {
			absSearchPaths = append(absSearchPaths, filepath.Join(projectRoot, p))
		}
	}
	absSearchPaths = append(absSearchPaths,
		filepath.Join(projectRoot, "plugins", "assets", pluginID, filePath),
	)

	for _, p := range absSearchPaths {
		data, err := os.ReadFile(p)
		if err == nil {
			ext := strings.ToLower(filepath.Ext(filePath))
			contentType := "application/octet-stream"
			switch ext {
			case ".js":
				contentType = "application/javascript"
			case ".css":
				contentType = "text/css"
			case ".html":
				contentType = "text/html"
			case ".json":
				contentType = "application/json"
			case ".wasm":
				contentType = "application/wasm"
			case ".png":
				contentType = "image/png"
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".svg":
				contentType = "image/svg+xml"
			}

			c.Header("Content-Type", contentType)
			c.Header("Cache-Control", "public, max-age=3600")
			c.Header("X-Plugin-Id", pluginID)
			c.Data(http.StatusOK, contentType, data)
			return
		}
	}

	c.Status(http.StatusNotFound)
}

func (pec *PluginEngineController) GetPluginData(c *gin.Context) {
	table := c.Param("table")
	if !utils.ValidateTableName(table) {
		response.GinError(c, errors.ErrParamValidation, "无效的表名")
		return
	}
	apiPath := fmt.Sprintf("/api/plugin-data/%s", table)
	ctx := c.Request.Context()

	items, err := service.ServiceGroupApp.PluginEngineService.GetPluginData(ctx, apiPath)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, items)
}

func (pec *PluginEngineController) CreatePluginData(c *gin.Context) {
	table := c.Param("table")
	if !utils.ValidateTableName(table) {
		response.GinError(c, errors.ErrParamValidation, "无效的表名")
		return
	}
	apiPath := fmt.Sprintf("/api/plugin-data/%s", table)
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.GinError(c, errors.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	result, err := service.ServiceGroupApp.PluginEngineService.CreatePluginData(ctx, apiPath, data)
	if err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "创建成功", result)
}

func (pec *PluginEngineController) UpdatePluginData(c *gin.Context) {
	table := c.Param("table")
	if !utils.ValidateTableName(table) {
		response.GinError(c, errors.ErrParamValidation, "无效的表名")
		return
	}
	apiPath := fmt.Sprintf("/api/plugin-data/%s", table)
	recordID := c.Param("id")
	if recordID == "" || len(recordID) > 50 {
		response.GinError(c, errors.ErrParamValidation, "无效的记录ID")
		return
	}
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.GinError(c, errors.ErrParamValidation, "参数错误")
		return
	}

	ctx := c.Request.Context()
	if err := service.ServiceGroupApp.PluginEngineService.UpdatePluginData(ctx, apiPath, recordID, data); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "更新成功", nil)
}

func (pec *PluginEngineController) DeletePluginData(c *gin.Context) {
	table := c.Param("table")
	if !utils.ValidateTableName(table) {
		response.GinError(c, errors.ErrParamValidation, "无效的表名")
		return
	}
	apiPath := fmt.Sprintf("/api/plugin-data/%s", table)
	recordID := c.Param("id")
	if recordID == "" || len(recordID) > 50 {
		response.GinError(c, errors.ErrParamValidation, "无效的记录ID")
		return
	}

	ctx := c.Request.Context()
	if err := service.ServiceGroupApp.PluginEngineService.DeletePluginData(ctx, apiPath, recordID); err != nil {
		response.GinError(c, errors.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccessWithMessage(c, "删除成功", nil)
}
