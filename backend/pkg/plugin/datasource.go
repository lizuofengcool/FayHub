package plugin

import "context"

type PluginResource struct {
	Name          string
	Slug          string
	Version       string
	Description   string
	EntryPoint    string
	MinAppVersion string
	DownloadURL   string
	PackageHash   string
	PackageSize   int64
	ManifestJSON  string
	SandboxConfig string
	WASMBytes     []byte
}

type PluginDataSource interface {
	GetPluginResource(ctx context.Context, pluginID string, version string) (*PluginResource, error)
}
