package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ExportColumn struct {
	Header string
	Field  string
}

type ExportData struct {
	Columns []ExportColumn
	Rows    []map[string]interface{}
}

func ExportCSV(data *ExportData) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	headers := make([]string, len(data.Columns))
	for i, col := range data.Columns {
		headers[i] = col.Header
	}
	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("写入CSV表头失败: %w", err)
	}

	for _, row := range data.Rows {
		record := make([]string, len(data.Columns))
		for i, col := range data.Columns {
			val := row[col.Field]
			record[i] = formatValue(val)
		}
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("写入CSV数据行失败: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV写入错误: %w", err)
	}

	return buf.Bytes(), nil
}

func ExportExcel(data *ExportData, sheetName string) ([]byte, error) {
	if sheetName == "" {
		sheetName = "Sheet1"
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("创建工作表失败: %w", err)
	}
	f.SetActiveSheet(sheet)

	defaultSheet := f.GetSheetName(0)
	if defaultSheet != sheetName {
		f.DeleteSheet(defaultSheet)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	for i, col := range data.Columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, col.Header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	for rowIdx, row := range data.Rows {
		for colIdx, col := range data.Columns {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			val := row[col.Field]
			setCellValue(f, sheetName, cell, val)
		}
	}

	for i := range data.Columns {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, col, col, 20)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成Excel文件失败: %w", err)
	}

	return buf.Bytes(), nil
}

func ExportStructsCSV(structs interface{}, columns []ExportColumn) ([]byte, error) {
	data, err := structsToExportData(structs, columns)
	if err != nil {
		return nil, err
	}
	return ExportCSV(data)
}

func ExportStructsExcel(structs interface{}, columns []ExportColumn, sheetName string) ([]byte, error) {
	data, err := structsToExportData(structs, columns)
	if err != nil {
		return nil, err
	}
	return ExportExcel(data, sheetName)
}

func structsToExportData(structs interface{}, columns []ExportColumn) (*ExportData, error) {
	val := reflect.ValueOf(structs)
	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("参数必须是切片类型")
	}

	rows := make([]map[string]interface{}, val.Len())
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		row := make(map[string]interface{})
		for _, col := range columns {
			field := elem.FieldByName(col.Field)
			if field.IsValid() {
				row[col.Field] = field.Interface()
			} else {
				row[col.Field] = ""
			}
		}
		rows[i] = row
	}

	return &ExportData{Columns: columns, Rows: rows}, nil
}

func formatValue(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func setCellValue(f *excelize.File, sheet, cell string, val interface{}) {
	switch v := val.(type) {
	case nil:
		f.SetCellValue(sheet, cell, "")
	case string:
		if strings.HasPrefix(v, "=") {
			f.SetCellFormula(sheet, cell, v)
		} else {
			f.SetCellValue(sheet, cell, v)
		}
	case int:
		f.SetCellValue(sheet, cell, v)
	case int64:
		f.SetCellValue(sheet, cell, v)
	case float64:
		f.SetCellValue(sheet, cell, v)
	case bool:
		f.SetCellValue(sheet, cell, v)
	default:
		f.SetCellValue(sheet, cell, fmt.Sprintf("%v", v))
	}
}
