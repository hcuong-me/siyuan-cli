package tool

// CatalogData is returned by the zero-input tools command.
type CatalogData struct {
	Protocol string            `json:"protocol"`
	Tools    map[string]Schema `json:"tools"`
}

// ToolsResponse returns the discoverable catalog without reading stdin.
func ToolsResponse() Response {
	return success("tools", "available agent tools", CatalogData{
		Protocol: ProtocolVersion,
		Tools:    Catalog(),
	})
}
