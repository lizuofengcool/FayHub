package export

import (
	"testing"
)

func TestExportCSV(t *testing.T) {
	data := &ExportData{
		Columns: []ExportColumn{
			{Header: "名称", Field: "name"},
			{Header: "年龄", Field: "age"},
		},
		Rows: []map[string]interface{}{
			{"name": "张三", "age": 25},
			{"name": "李四", "age": 30},
		},
	}

	result, err := ExportCSV(data)
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("ExportCSV returned empty result")
	}

	csvStr := string(result)
	if !contains(csvStr, "名称") || !contains(csvStr, "张三") {
		t.Fatalf("CSV output missing expected content: %s", csvStr)
	}
}

func TestExportExcel(t *testing.T) {
	data := &ExportData{
		Columns: []ExportColumn{
			{Header: "名称", Field: "name"},
			{Header: "年龄", Field: "age"},
		},
		Rows: []map[string]interface{}{
			{"name": "张三", "age": 25},
			{"name": "李四", "age": 30},
		},
	}

	result, err := ExportExcel(data, "测试")
	if err != nil {
		t.Fatalf("ExportExcel failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("ExportExcel returned empty result")
	}
}

func TestExportCSVEmpty(t *testing.T) {
	data := &ExportData{
		Columns: []ExportColumn{
			{Header: "名称", Field: "name"},
		},
		Rows: []map[string]interface{}{},
	}

	result, err := ExportCSV(data)
	if err != nil {
		t.Fatalf("ExportCSV empty failed: %v", err)
	}

	csvStr := string(result)
	if !contains(csvStr, "名称") {
		t.Fatalf("CSV output missing header: %s", csvStr)
	}
}

func TestExportStructsCSV(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	people := []Person{
		{Name: "Alice", Age: 25},
		{Name: "Bob", Age: 30},
	}

	columns := []ExportColumn{
		{Header: "姓名", Field: "Name"},
		{Header: "年龄", Field: "Age"},
	}

	result, err := ExportStructsCSV(people, columns)
	if err != nil {
		t.Fatalf("ExportStructsCSV failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("ExportStructsCSV returned empty result")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
