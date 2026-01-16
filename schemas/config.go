package schemas

type ConfigLog struct {
	Mode     string `validate:"oneof=both stderr file"`
	Filename string `validate:"required_unless=Mode stderr"`
	Level    string `validate:"oneof=fatal error warn info debug"`
}
type Config struct {
	Host           string   `validate:"hostname_port"`
	InCheck        string   `validate:"oneof=on off"`
	InTransport    string   `validate:"oneof=http stdio"`
	OutCheck       string   `validate:"oneof=on off"`
	OutCommand     []string `validate:"required_if=OutCheck on OutTransport stdio,dive"`
	OutEndpoint    string   `validate:"required_if=OutCheck on OutTransport http,url"`
	OutTransport   string   `validate:"oneof=http stdio"`
	Audit          string   `validate:"oneof=on off"`
	AuditEndpoint  string   `validate:"required_if=Audit on,url"`
	AuditToken     string   `validate:"required_if=Audit on"`
	AuditThreshold int      `validate:"required_if=Audit on,number,min=0,max=100"`
	AuditTimeout   int      `validate:"number,min=0"`
}
