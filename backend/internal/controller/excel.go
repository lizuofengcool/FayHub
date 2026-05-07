package controller

import (
	"bytes"
	"fmt"
	"net/http"

	"fayhub/internal/service"
	"fayhub/pkg/export"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/response"

	"github.com/gin-gonic/gin"
)

type ExcelController struct{}

func (ec *ExcelController) Import(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "请选择要导入的文件")
		return
	}
	defer file.Close()

	ctx := c.Request.Context()
	result, err := service.ServiceGroupApp.ExcelService.Import(ctx, header.Filename, file, nil)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, result)
}

func (ec *ExcelController) Preview(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.GinError(c, errs.ErrParamValidation, "请选择要预览的文件")
		return
	}
	defer file.Close()

	ctx := c.Request.Context()
	result, err := service.ServiceGroupApp.ExcelService.ParsePreview(ctx, header.Filename, file)
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	response.GinSuccess(c, result)
}

func (ec *ExcelController) DownloadTemplate(c *gin.Context) {
	columns := []export.ExportColumn{
		{Header: "名称", Field: "name"},
		{Header: "描述", Field: "description"},
		{Header: "备注", Field: "remark"},
	}

	data, err := service.ServiceGroupApp.ExcelService.GenerateTemplate(columns, "导入模板")
	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	filename := "import_template.xlsx"
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (ec *ExcelController) ExportGeneric(c *gin.Context) {
	format := c.DefaultQuery("format", "xlsx")
	prefix := c.DefaultQuery("prefix", "export")

	columns := []export.ExportColumn{
		{Header: "ID", Field: "id"},
		{Header: "名称", Field: "name"},
		{Header: "创建时间", Field: "created_at"},
	}

	data := &export.ExportData{
		Columns: columns,
		Rows:    []map[string]interface{}{},
	}

	var fileData []byte
	var err error

	switch format {
	case "csv":
		fileData, err = service.ServiceGroupApp.ExcelService.ExportCSV(data)
	default:
		fileData, err = service.ServiceGroupApp.ExcelService.ExportExcel(data, "Sheet1")
	}

	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	filename := service.ServiceGroupApp.ExcelService.GetDownloadFilename(prefix, format)
	contentType := service.ServiceGroupApp.ExcelService.GetContentType(format)

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, fileData)
}

func (ec *ExcelController) ExportData(c *gin.Context) {
	var req struct {
		Format   string                `json:"format"`
		Prefix   string                `json:"prefix"`
		Columns  []export.ExportColumn `json:"columns"`
		Rows     []map[string]interface{} `json:"rows"`
		SheetName string               `json:"sheet_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinError(c, errs.ErrParamValidation, "参数格式错误")
		return
	}

	if req.Format == "" {
		req.Format = "xlsx"
	}
	if req.SheetName == "" {
		req.SheetName = "Sheet1"
	}

	data := &export.ExportData{
		Columns: req.Columns,
		Rows:    req.Rows,
	}

	var fileData []byte
	var err error

	switch req.Format {
	case "csv":
		fileData, err = service.ServiceGroupApp.ExcelService.ExportCSV(data)
	default:
		fileData, err = service.ServiceGroupApp.ExcelService.ExportExcel(data, req.SheetName)
	}

	if err != nil {
		response.GinError(c, errs.ErrInternalServer, err.Error())
		return
	}

	filename := service.ServiceGroupApp.ExcelService.GetDownloadFilename(req.Prefix, req.Format)
	contentType := service.ServiceGroupApp.ExcelService.GetContentType(req.Format)

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, fileData)
}

func (ec *ExcelController) ExportBytes(c *gin.Context, data []byte, filename, contentType string) {
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, data)
}

func (ec *ExcelController) ExportFromReader(c *gin.Context, reader *bytes.Reader, filename, contentType string) {
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", contentType)
	c.DataFromReader(http.StatusOK, reader.Size(), contentType, reader, nil)
}
