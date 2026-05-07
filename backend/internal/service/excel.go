package service

import (
	"context"
	"fmt"
	"io"

	"fayhub/pkg/export"
	errs "fayhub/pkg/errors"
)

type ExcelService struct{}

type ImportResult struct {
	TotalRows   int                      `json:"total_rows"`
	SuccessRows int                      `json:"success_rows"`
	FailedRows  int                      `json:"failed_rows"`
	Errors      []ImportRowError         `json:"errors,omitempty"`
	Data        []map[string]interface{} `json:"data,omitempty"`
}

type ImportRowError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
}

type ImportHandler func(ctx context.Context, row map[string]interface{}, rowIndex int) error

func (s *ExcelService) Import(ctx context.Context, filename string, reader io.Reader, handler ImportHandler) (*ImportResult, error) {
	data, err := export.ParseFile(filename, reader)
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrParamValidation, err.Error())
	}

	result := &ImportResult{
		TotalRows: len(data.Rows),
		Data:      data.Rows,
	}

	if handler == nil {
		result.SuccessRows = len(data.Rows)
		return result, nil
	}

	for i, row := range data.Rows {
		if err := handler(ctx, row, i+2); err != nil {
			result.FailedRows++
			result.Errors = append(result.Errors, ImportRowError{
				Row:   i + 2,
				Error: err.Error(),
			})
		} else {
			result.SuccessRows++
		}
	}

	return result, nil
}

func (s *ExcelService) ExportExcel(data *export.ExportData, sheetName string) ([]byte, error) {
	return export.ExportExcel(data, sheetName)
}

func (s *ExcelService) ExportCSV(data *export.ExportData) ([]byte, error) {
	return export.ExportCSV(data)
}

func (s *ExcelService) ExportStructsExcel(structs interface{}, columns []export.ExportColumn, sheetName string) ([]byte, error) {
	return export.ExportStructsExcel(structs, columns, sheetName)
}

func (s *ExcelService) ExportStructsCSV(structs interface{}, columns []export.ExportColumn) ([]byte, error) {
	return export.ExportStructsCSV(structs, columns)
}

func (s *ExcelService) ParsePreview(ctx context.Context, filename string, reader io.Reader) (*ImportResult, error) {
	data, err := export.ParseFile(filename, reader)
	if err != nil {
		return nil, errs.NewServiceError(errs.ErrParamValidation, err.Error())
	}

	previewRows := data.Rows
	if len(previewRows) > 10 {
		previewRows = previewRows[:10]
	}

	return &ImportResult{
		TotalRows:   len(data.Rows),
		SuccessRows: 0,
		FailedRows:  0,
		Data:        previewRows,
	}, nil
}

func (s *ExcelService) GenerateTemplate(columns []export.ExportColumn, sheetName string) ([]byte, error) {
	if sheetName == "" {
		sheetName = "导入模板"
	}

	data := &export.ExportData{
		Columns: columns,
		Rows:    []map[string]interface{}{},
	}

	return export.ExportExcel(data, sheetName)
}

func (s *ExcelService) GetContentType(format string) string {
	switch format {
	case "csv":
		return "text/csv; charset=utf-8"
	default:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
}

func (s *ExcelService) GetDownloadFilename(prefix, format string) string {
	ext := ".xlsx"
	if format == "csv" {
		ext = ".csv"
	}
	return fmt.Sprintf("%s%s", prefix, ext)
}
