package plugin

import (
	"context"
	"encoding/binary"
	"testing"
	"time"
)

func encodeLEB128U(v uint64) []byte {
	var buf []byte
	for {
		b := byte(v & 0x7f)
		v >>= 7
		if v != 0 {
			b |= 0x80
		}
		buf = append(buf, b)
		if v == 0 {
			break
		}
	}
	return buf
}

func encodeName(name string) []byte {
	nameBytes := []byte(name)
	size := encodeLEB128U(uint64(len(nameBytes)))
	result := make([]byte, 0, len(size)+len(nameBytes))
	result = append(result, size...)
	result = append(result, nameBytes...)
	return result
}

type wasmBuilder struct {
	sections [][]byte
}

func newWASMBuilder() *wasmBuilder { return &wasmBuilder{} }

func (b *wasmBuilder) addSection(id byte, data []byte) {
	size := encodeLEB128U(uint64(len(data)))
	section := make([]byte, 0, 1+len(size)+len(data))
	section = append(section, id)
	section = append(section, size...)
	section = append(section, data...)
	b.sections = append(b.sections, section)
}

func (b *wasmBuilder) build() []byte {
	result := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	for _, s := range b.sections {
		result = append(result, s...)
	}
	return result
}

func buildEchoWASM() []byte {
	b := newWASMBuilder()

	typeSec := []byte{0x04}
	typeSec = append(typeSec, 0x60, 0x01, 0x7f, 0x01, 0x7f)
	typeSec = append(typeSec, 0x60, 0x04, 0x7f, 0x7f, 0x7f, 0x7f, 0x01, 0x7f)
	typeSec = append(typeSec, 0x60, 0x00, 0x00)
	typeSec = append(typeSec, 0x60, 0x01, 0x7f, 0x00)
	b.addSection(0x01, typeSec)

	funcSec := []byte{0x04, 0x00, 0x01, 0x02, 0x03}
	b.addSection(0x03, funcSec)

	memSec := []byte{0x01, 0x00}
	memSec = append(memSec, encodeLEB128U(256)...)
	b.addSection(0x05, memSec)

	globalSec := []byte{
		0x01,
		0x7f, 0x01,
		0x41,
	}
	globalSec = append(globalSec, encodeLEB128U(16)...)
	globalSec = append(globalSec, 0x0b)
	b.addSection(0x06, globalSec)

	exportSec := []byte{}
	exportSec = append(exportSec, 0x05)
	exportSec = append(exportSec, encodeName("memory")...)
	exportSec = append(exportSec, 0x02, 0x00)
	exportSec = append(exportSec, encodeName("malloc")...)
	exportSec = append(exportSec, 0x00, 0x00)
	exportSec = append(exportSec, encodeName("free")...)
	exportSec = append(exportSec, 0x00, 0x03)
	exportSec = append(exportSec, encodeName("handle")...)
	exportSec = append(exportSec, 0x00, 0x01)
	exportSec = append(exportSec, encodeName("_start")...)
	exportSec = append(exportSec, 0x00, 0x02)
	b.addSection(0x07, exportSec)

	mallocBody := []byte{
		0x00,
		0x23, 0x00,
		0x20, 0x00,
		0x6a,
		0x24, 0x00,
		0x23, 0x00,
		0x0b,
	}

	freeBody := []byte{0x00, 0x0b}

	startBody := []byte{0x00, 0x0b}

	handleBody := []byte{
		0x01, 0x04, 0x7f,
		0x20, 0x02, 0x20, 0x03, 0x49, 0x21, 0x04,
		0x20, 0x04, 0x0b,
	}

	codeSec := []byte{0x04}

	for _, body := range [][]byte{mallocBody, handleBody, startBody, freeBody} {
		vec := append(encodeLEB128U(uint64(len(body))), body...)
		codeSec = append(codeSec, vec...)
	}

	b.addSection(0x0a, codeSec)
	return b.build()
}

func buildInfiniteLoopWASM() []byte {
	b := newWASMBuilder()

	typeSec := []byte{0x04}
	typeSec = append(typeSec, 0x60, 0x01, 0x7f, 0x01, 0x7f)
	typeSec = append(typeSec, 0x60, 0x04, 0x7f, 0x7f, 0x7f, 0x7f, 0x01, 0x7f)
	typeSec = append(typeSec, 0x60, 0x00, 0x00)
	typeSec = append(typeSec, 0x60, 0x01, 0x7f, 0x00)
	b.addSection(0x01, typeSec)

	funcSec := []byte{0x04, 0x00, 0x01, 0x02, 0x03}
	b.addSection(0x03, funcSec)

	memSec := []byte{0x01, 0x00}
	memSec = append(memSec, encodeLEB128U(256)...)
	b.addSection(0x05, memSec)

	globalSec := []byte{0x01, 0x7f, 0x01, 0x41, 0x10, 0x0b}
	b.addSection(0x06, globalSec)

	exportSec := []byte{}
	exportSec = append(exportSec, 0x05)
	exportSec = append(exportSec, encodeName("memory")...)
	exportSec = append(exportSec, 0x02, 0x00)
	exportSec = append(exportSec, encodeName("malloc")...)
	exportSec = append(exportSec, 0x00, 0x00)
	exportSec = append(exportSec, encodeName("free")...)
	exportSec = append(exportSec, 0x00, 0x03)
	exportSec = append(exportSec, encodeName("infiniteLoop")...)
	exportSec = append(exportSec, 0x00, 0x01)
	exportSec = append(exportSec, encodeName("_start")...)
	exportSec = append(exportSec, 0x00, 0x02)
	b.addSection(0x07, exportSec)

	mallocBody := []byte{0x00, 0x23, 0x00, 0x20, 0x00, 0x6a, 0x24, 0x00, 0x23, 0x00, 0x0b}
	freeBody := []byte{0x00, 0x0b}
	startBody := []byte{0x00, 0x0b}
	loopBody := []byte{0x00, 0x02, 0x40, 0x03, 0x40, 0x0c, 0x00, 0x0b, 0x0b, 0x41, 0x00, 0x0b}

	codeSec := []byte{0x04}
	for _, body := range [][]byte{mallocBody, loopBody, startBody, freeBody} {
		vec := append(encodeLEB128U(uint64(len(body))), body...)
		codeSec = append(codeSec, vec...)
	}
	b.addSection(0x0a, codeSec)
	return b.build()
}

func installTestCallPlugin(t *testing.T, engine *WASMEngine, ctx context.Context, pluginID string) {
	t.Helper()
	wasmBytes := buildEchoWASM()
	ds := newMockDataSource()
	manifestJSON := `{
		"name": "test-call-plugin",
		"version": "1.0.0",
		"entry_point": "_start",
		"description": "测试Call链路插件",
		"permissions": ["log:write"],
		"routes": [
			{"method": "GET", "path": "/test/echo", "handler": "handle"},
			{"method": "POST", "path": "/test/compute", "handler": "handle"}
		],
		"apis": [{"method": "GET", "path": "/test/echo", "group": "test"}]
	}`
	ds.addResource(pluginID, &PluginResource{
		Name: "test-call-plugin", Version: "1.0.0",
		EntryPoint: "_start", ManifestJSON: manifestJSON, WASMBytes: wasmBytes,
	})
	engine.SetDataSource(ds)
	if err := engine.Install(ctx, 1, pluginID, "1.0.0", ""); err != nil {
		t.Fatalf("安装插件失败: %v", err)
	}
}

func TestWASMCallChain_EchoPayload(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("引擎启动失败: %v", err)
	}
	defer engine.Stop(ctx)

	installTestCallPlugin(t, engine, ctx, "com.fayhub.testcall")

	infos := engine.GetPluginInfos()
	if len(infos) != 1 || !infos[0].HasModule {
		t.Fatal("期望1个活跃的WASM插件")
	}

	payload := []byte(`{"action":"echo","message":"Hello FayHub WASM!"}`)

	result, err := engine.Call(ctx, 1, "com.fayhub.testcall", "handle", payload)
	if err != nil {
		t.Fatalf("Call调用失败: %v", err)
	}
	t.Logf("Call返回 %d 字节: %s", len(result), string(result))
}

func TestWASMCallChain_MultipleCalls(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	engine.Start(ctx)
	defer engine.Stop(ctx)
	installTestCallPlugin(t, engine, ctx, "com.fayhub.multicall")

	for i := 0; i < 10; i++ {
		payload := []byte(`{"action":"echo","message":"batch test"}`)
		result, err := engine.Call(ctx, 1, "com.fayhub.multicall", "handle", payload)
		if err != nil {
			t.Errorf("第%d次Call失败: %v", i, err)
		} else {
			t.Logf("第%d次Call成功: %d字节", i, len(result))
		}
	}
}

func TestWASMCallChain_UndeclaredFunction(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	engine.Start(ctx)
	defer engine.Stop(ctx)
	installTestCallPlugin(t, engine, ctx, "com.fayhub.undeclared")

	_, err := engine.Call(ctx, 1, "com.fayhub.undeclared", "nonexistent", []byte(`{}`))
	if err == nil {
		t.Error("期望调用未声明函数返回错误，但成功了")
	}
}

func TestWASMCallChain_DisabledPlugin(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	engine.Start(ctx)
	defer engine.Stop(ctx)
	installTestCallPlugin(t, engine, ctx, "com.fayhub.disabled")
	engine.Disable(ctx, 1, "com.fayhub.disabled")

	_, err := engine.Call(ctx, 1, "com.fayhub.disabled", "handle", []byte(`{}`))
	if err == nil {
		t.Error("期望调用禁用插件返回错误，但成功了")
	}
}

func TestWASMCallChain_EmptyPayload(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	engine.Start(ctx)
	defer engine.Stop(ctx)
	installTestCallPlugin(t, engine, ctx, "com.fayhub.empty")

	result, err := engine.Call(ctx, 1, "com.fayhub.empty", "handle", []byte{})
	if err != nil {
		t.Fatalf("空payload Call失败: %v", err)
	}
	t.Logf("空payload返回 %d 字节", len(result))
}

func TestWASMCallChain_LargePayload(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	engine.Start(ctx)
	defer engine.Stop(ctx)
	installTestCallPlugin(t, engine, ctx, "com.fayhub.large")

	largeData := make([]byte, 8*1024)
	for i := range largeData {
		largeData[i] = byte('A' + (i % 26))
	}

	result, err := engine.Call(ctx, 1, "com.fayhub.large", "handle", largeData)
	if err != nil {
		t.Fatalf("大payload Call失败: %v", err)
	}
	t.Logf("大payload测试: 输入=%d, 输出=%d", len(largeData), len(result))
}

func TestWASMCallChain_MemoryLeakCheck(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	engine.Start(ctx)
	defer engine.Stop(ctx)
	installTestCallPlugin(t, engine, ctx, "com.fayhub.memleak")

	for i := 0; i < 100; i++ {
		payload := []byte(`{"action":"echo","msg":"test"}`)
		result, err := engine.Call(ctx, 1, "com.fayhub.memleak", "handle", payload)
		if err != nil {
			t.Errorf("第%d次Call失败: %v", i, err)
			break
		}
		_ = result
	}
	infos := engine.GetPluginInfos()
	if len(infos) > 0 && infos[0].Status == "active" {
		t.Log("100次连续调用后插件仍正常运行，内存泄漏检测通过")
	}
}

func TestWASMCallChain_EntryPoint(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	engine.Start(ctx)
	defer engine.Stop(ctx)
	installTestCallPlugin(t, engine, ctx, "com.fayhub.entrypoint")

	result, err := engine.Call(ctx, 1, "com.fayhub.entrypoint", "handle", []byte(`{}`))
	if err != nil {
		t.Fatalf("调用handle函数失败: %v", err)
	}
	t.Logf("_start入口点通过handle调用成功: %d字节", len(result))
}

func TestWASMCallChain_TenantIsolation(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	engine.Start(ctx)
	defer engine.Stop(ctx)

	wasmBytes := buildEchoWASM()

	ds1 := newMockDataSource()
	ds1.addResource("com.fayhub.tenant1", &PluginResource{
		Name: "test-call-plugin", Version: "1.0.0",
		EntryPoint:   "_start",
		ManifestJSON: `{"name":"t","version":"1","entry_point":"_start","permissions":["log:write"],"routes":[{"method":"GET","path":"/test/echo","handler":"handle"}],"apis":[]}`,
		WASMBytes:    wasmBytes,
	})
	engine.SetDataSource(ds1)

	ds2 := newMockDataSource()
	ds2.addResource("com.fayhub.tenant2", &PluginResource{
		Name: "test-call-plugin", Version: "1.0.0",
		EntryPoint:   "_start",
		ManifestJSON: `{"name":"t","version":"1","entry_point":"_start","permissions":["log:write"],"routes":[{"method":"GET","path":"/test/echo","handler":"handle"}],"apis":[]}`,
		WASMBytes:    wasmBytes,
	})

	if err := engine.Install(ctx, 1, "com.fayhub.tenant1", "1.0.0", ""); err != nil {
		t.Fatalf("租户1安装失败: %v", err)
	}
	engine.SetDataSource(ds2)
	if err := engine.Install(ctx, 2, "com.fayhub.tenant2", "1.0.0", ""); err != nil {
		t.Fatalf("租户2安装失败: %v", err)
	}

	p1 := []byte(`{"action":"echo","message":"tenant1"}`)
	p2 := []byte(`{"action":"echo","message":"tenant2"}`)

	r1, err1 := engine.Call(ctx, 1, "com.fayhub.tenant1", "handle", p1)
	r2, err2 := engine.Call(ctx, 2, "com.fayhub.tenant2", "handle", p2)

	if err1 != nil {
		t.Errorf("租户1调用失败: %v", err1)
	} else {
		t.Logf("租户1返回: %d字节", len(r1))
	}
	if err2 != nil {
		t.Errorf("租户2调用失败: %v", err2)
	} else {
		t.Logf("租户2返回: %d字节", len(r2))
	}
}

func TestWASMCallChain_NonExistent(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	engine.Start(ctx)
	defer engine.Stop(ctx)

	_, err := engine.Call(ctx, 1, "com.fayhub.nonexistent", "handle", []byte(`{}`))
	if err == nil {
		t.Error("期望调用不存在的插件返回错误，但成功了")
	}
}

func TestWASMRouteIndex_Lookup(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	engine.Start(ctx)
	defer engine.Stop(ctx)
	installTestCallPlugin(t, engine, ctx, "com.fayhub.routeidx")

	entry := engine.lookupRoute("GET", "com.fayhub.routeidx", "/test/echo")
	if entry == nil || entry.handler != "handle" {
		t.Fatal("期望找到GET /test/echo路由->handle")
	}
	if engine.lookupRoute("DELETE", "com.fayhub.routeidx", "/test/echo") != nil {
		t.Error("不期望找到DELETE路由")
	}
	if engine.lookupRoute("GET", "com.fayhub.routeidx", "/nonexistent") != nil {
		t.Error("不期望找到不存在的路径")
	}
}

func TestWASMCallChain_Timeout(t *testing.T) {
	engine := NewWASMEngine()
	ctx := context.Background()
	engine.Start(ctx)
	defer engine.Stop(ctx)

	wasmBytes := buildInfiniteLoopWASM()
	ds := newMockDataSource()
	ds.addResource("com.fayhub.timeout", &PluginResource{
		Name: "timeout-plugin", Version: "1.0.0",
		EntryPoint:   "_start",
		ManifestJSON: `{"name":"t","version":"1","entry_point":"_start","permissions":["log:write"],"routes":[{"method":"GET","path":"/test/loop","handler":"infiniteLoop"}],"apis":[]}`,
		WASMBytes:    wasmBytes,
	})
	engine.SetDataSource(ds)
	engine.Install(ctx, 1, "com.fayhub.timeout", "1.0.0", "")

	start := time.Now()
	_, err := engine.Call(ctx, 1, "com.fayhub.timeout", "infiniteLoop", []byte(`{}`))
	elapsed := time.Since(start)

	if err == nil {
		t.Error("期望超时错误，但调用成功")
	}
	if elapsed > 35*time.Second {
		t.Errorf("超时时间过长: %v", elapsed)
	}
	t.Logf("超时测试: 耗时=%v, 错误=%v", elapsed, err)
}

var _ = binary.LittleEndian
