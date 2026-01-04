package schemas

type Message struct {
	JsonRPCVersion string `json:"jsonrpc"`
	ID             uint32 `json:"id"`
	Method         string `json:"method"`
	Params         any    `json:"params"`
}
