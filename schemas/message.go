package schemas

type MCPClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type MCPServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type MCPClientCapabilities struct {
	Elicitation any `json:"elicitation"`
	Roots       any `json:"roots"`
}
type MCPServerCapabilities struct {
	Logging any `json:"logging"`
	Tools   any `json:"tools"`
}
type MCPToolInputProp struct {
	Type        string `json:"type"`
	Enum        []any  `json:"enum"`
	Description string `json:"description"`
	Default     any    `json:"default"`
}
type MCPToolInput struct {
	Type       string                      `json:"type"`
	Properties map[string]MCPToolInputProp `json:"properties"`
	Required   []string                    `json:"required"`
}
type MCPTool struct {
	Name        string       `json:"name"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	InputSchema MCPToolInput `json:"inputSchema"`
}
type MCPResource struct{}

type MCPRequestParams struct {
	// Запрос может содержать огромный перечень различных параметров,
	// что зависит от вызываемого метода (method: "...").
	// Но самые важные для работы программы объявлены ниже.
	ClientInfo      MCPClientInfo         `json:"clientInfo"`
	ProtocolVersion string                `json:"protocolVersion" validate:"len=0|datetime=2006-01-02"`
	Capabililties   MCPClientCapabilities `json:"capabilities"`
}
type MCPRequest struct {
	JsonRPCVersion string           `json:"jsonrpc" validate:"eq=2.0"`
	ID             uint32           `json:"id" validate:"number"`
	Method         string           `json:"method" validate:"ascii"`
	Params         MCPRequestParams `json:"params"`
}
type MCPResponseResult struct {
	// Тип Content стоит расписать подробнее, включая type, text и др.
	Content      []any                 `json:"content"`
	Capabilities MCPServerCapabilities `json:"capabilities"`
	ServerInfo   MCPServerInfo         `json:"serverInfo"`
	Tools        []MCPTool             `json:"tools"`
	Resources    []MCPResource         `json:"resources"`
}
type MCPResponseError struct {
	Code    int16  `json:"code"`
	Message string `json:"message"`
}
type MCPResponse struct {
	JsonRPCVersion  string            `json:"jsonrpc" validate:"eq=2.0"`
	ID              uint32            `json:"id" validate:"number"`
	Result          MCPResponseResult `json:"result"`
	ProtocolVersion string            `json:"protocolVersion" validate:"len=0|datetime=2006-01-02"`
	Error           MCPResponseError  `json:"error"`
}
