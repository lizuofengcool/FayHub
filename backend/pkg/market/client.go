package market

import (
	"bytes"
	"context"
	"encoding/json"
	"fayhub/pkg/config"
	"fayhub/pkg/logger"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Client struct {
	baseURL      string
	httpClient   *http.Client
	ssoToken     string
	serviceToken string
}

type PluginListItem struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Slug           string                 `json:"slug"`
	Description    string                 `json:"description"`
	CoverImage     string                 `json:"coverImage"`
	Price          float64                `json:"price"`
	OriginalPrice  *float64               `json:"originalPrice"`
	Tags           []string               `json:"tags"`
	TotalDownloads int                    `json:"totalDownloads"`
	TotalSales     int                    `json:"totalSales"`
	AverageRating  float64                `json:"averageRating"`
	Category       string                 `json:"category"`
	Developer      map[string]interface{} `json:"developer"`
	Status         string                 `json:"status"`
	CreatedAt      string                 `json:"createdAt"`
	UpdatedAt      string                 `json:"updatedAt"`
}

// 获取开发者名称的辅助方法
func (p *PluginListItem) GetDeveloperName() string {
	if p.Developer == nil {
		return "未知开发者"
	}
	if name, ok := p.Developer["name"].(string); ok {
		return name
	}
	if teamName, ok := p.Developer["teamName"].(string); ok {
		return teamName
	}
	return "未知开发者"
}

type PluginListResponse struct {
	List     []PluginListItem `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type PluginDetail struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Slug           string                 `json:"slug"`
	Description    string                 `json:"description"`
	CoverImage     string                 `json:"coverImage"`
	Price          float64                `json:"price"`
	OriginalPrice  *float64               `json:"originalPrice"`
	Tags           []string               `json:"tags"`
	TotalDownloads int                    `json:"totalDownloads"`
	TotalSales     int                    `json:"totalSales"`
	AverageRating  float64                `json:"averageRating"`
	Category       string                 `json:"category"`
	Developer      map[string]interface{} `json:"developer"`
	Status         string                 `json:"status"`
	Versions       []PluginVersion        `json:"versions"`
	CreatedAt      string                 `json:"createdAt"`
	UpdatedAt      string                 `json:"updatedAt"`
}

// 获取开发者名称的辅助方法
func (p *PluginDetail) GetDeveloperName() string {
	if p.Developer == nil {
		return "未知开发者"
	}
	if name, ok := p.Developer["name"].(string); ok {
		return name
	}
	if teamName, ok := p.Developer["teamName"].(string); ok {
		return teamName
	}
	return "未知开发者"
}

type PluginVersion struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Changelog   string `json:"changelog"`
	DownloadURL string `json:"downloadUrl"`
	WasmURL     string `json:"wasmUrl"`
	ManifestURL string `json:"manifestUrl"`
	Signature   string `json:"signature"`
	CreatedAt   string `json:"createdAt"`
	Status      string `json:"status"`
}

type InstallTokenResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	InstallToken  string `json:"installToken"`
	PluginID      string `json:"pluginId"`
	TargetVersion string `json:"targetVersion"`
}

type CategoryItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	SortOrder int    `json:"sortOrder"`
}

type VerifyLicenseRequest struct {
	LicenseKey string `json:"licenseKey"`
	FayhubURL  string `json:"fayhubUrl"`
}

type VerifyLicenseResponse struct {
	Valid     bool        `json:"valid"`
	Message   string      `json:"message"`
	Plugin    interface{} `json:"plugin"`
	ExpiresAt string      `json:"expiresAt"`
}

type VerifyInstallTokenResponse struct {
	Valid             bool                `json:"valid"`
	Message           string              `json:"message"`
	PluginID          string              `json:"pluginId"`
	PluginName        string              `json:"pluginName"`
	PluginSlug        string              `json:"pluginSlug"`
	PluginDescription string              `json:"pluginDescription"`
	PluginIcon        string              `json:"pluginIcon"`
	LicenseKey        string              `json:"licenseKey"`
	FayhubURL         string              `json:"fayhubUrl"`
	LatestVersion     *InstallVersionInfo `json:"latestVersion"`
}

type InstallVersionInfo struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	DownloadURL string `json:"downloadUrl"`
	PackageHash string `json:"packageHash"`
	Changelog   string `json:"changelog"`
}

type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Error      string `json:"error"`
}

var defaultClient *Client

func InitClient() {
	cfg := config.GlobalConfig
	if cfg == nil {
		return
	}

	baseURL := cfg.Domains.MarketAPIURL
	if baseURL == "" {
		baseURL = cfg.Domains.MarketURL
	}
	if baseURL == "" {
		baseURL = "https://www.fayhub.com"
	}

	serviceToken := cfg.System.ServiceToken
	if serviceToken == "" {
		log.Println("⚠️  FAYHUB_SERVICE_TOKEN 未配置，市场 API 调用将失败")
	}

	defaultClient = &Client{
		baseURL:      baseURL,
		serviceToken: serviceToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func GetClient() *Client {
	if defaultClient == nil {
		InitClient()
	}
	return defaultClient
}

func (c *Client) SetSSOToken(token string) {
	c.ssoToken = token
}

func (c *Client) EnsureAuthenticated(ctx context.Context) error {
	if c.ssoToken != "" {
		return nil
	}
	return nil
}

func (c *Client) SearchPlugins(ctx context.Context, keyword string, page, pageSize int, categoryID string) (*PluginListResponse, error) {
	url := fmt.Sprintf("%s/api/plugins?page=%d&pageSize=%d", c.baseURL, page, pageSize)
	if keyword != "" {
		url += fmt.Sprintf("&keyword=%s", keyword)
	}
	if categoryID != "" {
		url += fmt.Sprintf("&categoryId=%s", categoryID)
	}

	body, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("搜索插件失败: %w", err)
	}

	var wrapper struct {
		Code    int                `json:"code"`
		Message string             `json:"message"`
		Data    PluginListResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("解析插件列表失败: %w", err)
	}

	if wrapper.Code != 0 {
		return nil, fmt.Errorf("Market API错误: %s", wrapper.Message)
	}

	return &wrapper.Data, nil
}

func (c *Client) GetPluginDetail(ctx context.Context, pluginID string) (*PluginDetail, error) {
	url := fmt.Sprintf("%s/api/plugins/%s", c.baseURL, pluginID)

	body, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("获取插件详情失败: %w", err)
	}

	var wrapper struct {
		Code    int          `json:"code"`
		Message string       `json:"message"`
		Data    PluginDetail `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("解析插件详情失败: %w", err)
	}

	if wrapper.Code != 0 {
		return nil, fmt.Errorf("Market API错误: %s", wrapper.Message)
	}

	return &wrapper.Data, nil
}

func (c *Client) GetPluginVersions(ctx context.Context, pluginID string) ([]PluginVersion, error) {
	url := fmt.Sprintf("%s/api/plugins/%s/versions", c.baseURL, pluginID)

	body, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("获取插件版本列表失败: %w", err)
	}

	var wrapper struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    []PluginVersion `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("解析插件版本列表失败: %w", err)
	}

	if wrapper.Code != 0 {
		return nil, fmt.Errorf("Market API错误: %s", wrapper.Message)
	}

	return wrapper.Data, nil
}

func (c *Client) InstallPlugin(ctx context.Context, pluginID string, targetVersion string, fayhubUserId string) (*InstallTokenResponse, error) {
	url := fmt.Sprintf("%s/api/plugins/%s/install", c.baseURL, pluginID)

	reqBody := map[string]string{
		"targetVersion": targetVersion,
	}
	if fayhubUserId != "" {
		reqBody["fayhubUserId"] = fayhubUserId
	}

	body, err := c.doRequest(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("安装插件失败: %w", err)
	}

	var wrapper struct {
		Code    int                  `json:"code"`
		Message string               `json:"message"`
		Data    InstallTokenResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("解析安装响应失败: %w", err)
	}

	if wrapper.Code != 0 {
		return nil, fmt.Errorf("Market API错误: %s", wrapper.Message)
	}

	return &wrapper.Data, nil
}

func (c *Client) VerifyLicense(ctx context.Context, licenseKey string, fayhubURL string) (*VerifyLicenseResponse, error) {
	url := fmt.Sprintf("%s/api/licenses/verify", c.baseURL)

	reqBody := VerifyLicenseRequest{
		LicenseKey: licenseKey,
		FayhubURL:  fayhubURL,
	}

	body, err := c.doRequest(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("验证License失败: %w", err)
	}

	var wrapper struct {
		Code    int                   `json:"code"`
		Message string                `json:"message"`
		Data    VerifyLicenseResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("解析License验证响应失败: %w", err)
	}

	if wrapper.Code != 0 {
		return nil, fmt.Errorf("Market API错误: %s", wrapper.Message)
	}

	return &wrapper.Data, nil
}

func (c *Client) VerifyInstallToken(ctx context.Context, installToken string) (*VerifyInstallTokenResponse, error) {
	url := fmt.Sprintf("%s/api/licenses/verify-install", c.baseURL)

	reqBody := map[string]string{
		"installToken": installToken,
	}

	body, err := c.doRequest(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("验证安装令牌失败: %w", err)
	}

	var wrapper struct {
		Code    int                        `json:"code"`
		Message string                     `json:"message"`
		Data    VerifyInstallTokenResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("解析安装令牌验证响应失败: %w", err)
	}

	if wrapper.Code != 0 {
		return nil, fmt.Errorf("Market API错误: %s", wrapper.Message)
	}

	return &wrapper.Data, nil
}

func (c *Client) GetCategories(ctx context.Context) ([]CategoryItem, error) {
	url := fmt.Sprintf("%s/api/plugins/categories", c.baseURL)

	body, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("获取分类列表失败: %w", err)
	}

	var wrapper struct {
		Code    int            `json:"code"`
		Message string         `json:"message"`
		Data    []CategoryItem `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("解析分类列表失败: %w", err)
	}

	if wrapper.Code != 0 {
		return nil, fmt.Errorf("Market API错误: %s", wrapper.Message)
	}

	return wrapper.Data, nil
}

func (c *Client) DownloadWASM(ctx context.Context, downloadURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建下载请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载WASM文件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载WASM文件失败: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取WASM文件失败: %w", err)
	}

	return data, nil
}

func (c *Client) DownloadManifest(ctx context.Context, manifestURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", manifestURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建下载请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载清单文件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载清单文件失败: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取清单文件失败: %w", err)
	}

	return string(data), nil
}

func (c *Client) GetPublicKey(ctx context.Context) (string, error) {
	url := c.baseURL + "/api/v1/public-key"

	data, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("获取市场公钥失败: %w", err)
	}

	var result struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("解析公钥响应失败: %w", err)
	}

	if result.PublicKey == "" {
		return "", fmt.Errorf("市场未返回公钥")
	}

	return result.PublicKey, nil
}

func (c *Client) doRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
	}

	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
		}

		var reader io.Reader
		if reqBody != nil {
			reader = bytes.NewReader(reqBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reader)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		if c.ssoToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.ssoToken)
		}
		if c.serviceToken != "" {
			req.Header.Set("X-Service-Token", c.serviceToken)
		}

		logger.Info(ctx, fmt.Sprintf("[MarketAPI] %s %s (attempt %d/%d)", method, url, attempt+1, maxRetries))

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("请求失败: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("读取响应失败: %w", err)
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("Market API服务端错误: HTTP %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			var errResp ErrorResponse
			if json.Unmarshal(respBody, &errResp) == nil {
				return nil, fmt.Errorf("Market API错误(%d): %s", resp.StatusCode, errResp.Message)
			}
			return nil, fmt.Errorf("Market API错误: HTTP %d", resp.StatusCode)
		}

		return respBody, nil
	}

	return nil, fmt.Errorf("重试%d次后仍失败: %w", maxRetries, lastErr)
}
