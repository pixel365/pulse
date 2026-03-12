package config

import "time"

type CheckType string
type GRPCHealthStatus string

const (
	HTTP CheckType = "http"
	TCP  CheckType = "tcp"
	GRPC CheckType = "grpc"
	TLS  CheckType = "tls"
	DNS  CheckType = "dns"

	GRPCHealthUnknown        GRPCHealthStatus = "UNKNOWN"
	GRPCHealthServing        GRPCHealthStatus = "SERVING"
	GRPCHealthNotServing     GRPCHealthStatus = "NOT_SERVING"
	GRPCHealthServiceUnknown GRPCHealthStatus = "SERVICE_UNKNOWN"
)

type StringExpect struct {
	Contains string `yaml:"contains" json:"contains"`
	Equals   string `yaml:"equals"   json:"equals"`
}

type DNSExpect struct {
	Contains []string `yaml:"contains" json:"contains"`
	Equals   []string `yaml:"equals"   json:"equals"`
}

type GRPCHealthRequest struct {
	Service string `yaml:"service" json:"service"`
}

type ServiceSet struct {
	Services []Service `yaml:"services" json:"services" validate:"required,min=1,dive"`
}

type Service struct {
	ID   string `yaml:"id"   json:"id"   validate:"required,min=1"`
	Name string `yaml:"name" json:"name" validate:"required,min=1"`
}

type CheckFields struct {
	ID      string `yaml:"id"      json:"id"      validate:"required,min=1"`
	Name    string `yaml:"name"    json:"name"    validate:"required,min=1"`
	Service string `yaml:"service" json:"service" validate:"required,min=1"`

	Type CheckType `yaml:"type" json:"type" validate:"required,oneof=http tcp grpc tls dns"`

	//nolint:lll,nolintlint
	Tags []string `yaml:"tags" json:"tags" validate:"omitempty,min=1,dive,min=1"`

	Timeout          time.Duration `yaml:"timeout"           json:"timeout"           validate:"required,gt=0ms"`
	Jitter           time.Duration `yaml:"jitter"            json:"jitter"            validate:"gte=0"`
	Retries          int           `yaml:"retries"           json:"retries"           validate:"required,gte=0"`
	FailureThreshold int           `yaml:"failure_threshold" json:"failure_threshold" validate:"required,gte=1"`
	SuccessThreshold int           `yaml:"success_threshold" json:"success_threshold" validate:"required,gte=1"`
	Interval         time.Duration `yaml:"interval"          json:"interval"          validate:"required,gt=0ms"`
	Enabled          bool          `yaml:"enabled"           json:"enabled"`
}

type SpecField[T any] struct {
	Spec T `yaml:"spec" json:"spec" validate:"required"`
}

type Check struct {
	SpecField[any] `yaml:",inline" json:",inline"`
	CheckFields    `yaml:",inline" json:",inline"`
}

type TypedCheck[T any] struct {
	SpecField[T] `yaml:",inline" json:",inline"`
	CheckFields  `yaml:",inline" json:",inline"`
}

type CheckSet struct {
	Checks []Check `yaml:"checks" json:"checks" validate:"required,min=1,dive"`
}

type HttpSpec struct {
	Headers      map[string]string `yaml:"headers"       json:"headers"       validate:"omitempty,min=1"`
	Payload      map[string]any    `yaml:"payload"       json:"payload"       validate:"omitempty,min=1"`
	ExpectedBody *StringExpect     `yaml:"expected_body" json:"expected_body" validate:"omitempty"`
	URL          string            `yaml:"url"           json:"url"           validate:"required,url"`

	//nolint:lll,nolintlint
	Method string `yaml:"method" json:"method" validate:"required,oneof=GET POST"`

	SuccessCodes    []int `yaml:"success_codes"    json:"success_codes"    validate:"required,min=1"`
	FollowRedirects bool  `yaml:"follow_redirects" json:"follow_redirects" validate:"omitempty"`
}

type TCPSpec struct {
	Expect *StringExpect `yaml:"expect" json:"expect" validate:"omitempty"`
	Host   string        `yaml:"host"   json:"host"   validate:"required,hostname_rfc1123|ip"`
	Send   string        `yaml:"send"   json:"send"   validate:"omitempty,min=1"`
	Port   int           `yaml:"port"   json:"port"   validate:"required,gte=1,lte=65535"`
}

type GRPCSpec struct {
	Metadata map[string]string  `yaml:"metadata" json:"metadata" validate:"omitempty,min=1"`
	Host     string             `yaml:"host"     json:"host"     validate:"required,hostname_rfc1123|ip"`
	Service  string             `yaml:"service"  json:"service"  validate:"required"`
	Method   string             `yaml:"method"   json:"method"   validate:"required"`
	Request  *GRPCHealthRequest `yaml:"request"  json:"request"  validate:"omitempty"`

	//nolint:lll,nolintlint
	ExpectedHealthStatus GRPCHealthStatus `yaml:"expected_status" json:"expected_status" validate:"required,oneof=UNKNOWN SERVING NOT_SERVING SERVICE_UNKNOWN"`

	Port int `yaml:"port" json:"port" validate:"required,gte=1,lte=65535"`
}

type DNSSpec struct {
	Expect     *DNSExpect `yaml:"expect"      json:"expect"      validate:"omitempty"`
	Server     string     `yaml:"server"      json:"server"      validate:"required,hostname_rfc1123|ip"`
	Name       string     `yaml:"name"        json:"name"        validate:"required,min=1"`
	RecordType string     `yaml:"record_type" json:"record_type" validate:"required,oneof=A AAAA CNAME TXT MX NS SRV"`
}

type TLSSpec struct {
	Host        string        `yaml:"host"         json:"host"         validate:"required,hostname_rfc1123|ip"`
	ServerName  string        `yaml:"server_name"  json:"server_name"  validate:"omitempty,min=1"`
	Port        int           `yaml:"port"         json:"port"         validate:"required,gte=1,lte=65535"`
	MinValidity time.Duration `yaml:"min_validity" json:"min_validity" validate:"required,gte=1h"`
}
