package config

import (
	"errors"
	"maps"
	"reflect"
	"slices"
	"time"
)

type CheckType string
type GRPCHealthStatus string
type RecordType string
type GRPCHealthService string
type GRPCHealthMethod string
type StatusImpact string

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

	ARecord     RecordType = "A"
	AAAARecord  RecordType = "AAAA"
	CNAMERecord RecordType = "CNAME"
	TXTRecord   RecordType = "TXT"
	MXRecord    RecordType = "MX"
	NSRecord    RecordType = "NS"
	SRVRecord   RecordType = "SRV"

	HealthService GRPCHealthService = "grpc.health.v1.Health"
	HealthMethod  GRPCHealthMethod  = "Check"

	StatusImpactNone     StatusImpact = "none"
	StatusImpactMinor    StatusImpact = "minor"
	StatusImpactMajor    StatusImpact = "major"
	StatusImpactCritical StatusImpact = "critical"
)

type Comparer[T any] interface {
	Cmp(other T) bool
}

type StringExpect struct {
	Contains string `yaml:"contains" json:"contains"`
	Equals   string `yaml:"equals"   json:"equals"`
}

func (s *StringExpect) Validate() error {
	if s.Contains != "" && s.Equals != "" {
		return errors.New("cannot specify both 'contains' and 'equals'")
	}

	if s.Contains == "" && s.Equals == "" {
		return errors.New("must specify either 'contains' or 'equals'")
	}

	if s.Contains != "" {
		return validateString(s.Contains)
	}

	if s.Equals != "" {
		return validateString(s.Equals)
	}

	return nil
}

type DNSExpect struct {
	Contains []string `yaml:"contains" json:"contains"`
	Equals   []string `yaml:"equals"   json:"equals"`
}

func (s *DNSExpect) Validate() error {
	if len(s.Contains) == 0 && len(s.Equals) == 0 {
		return errors.New("must specify at least one of 'contains' or 'equals'")
	}

	if len(s.Contains) > 0 && len(s.Equals) > 0 {
		return errors.New("cannot specify both 'contains' and 'equals'")
	}

	for _, v := range s.Contains {
		if err := validateString(v); err != nil {
			return err
		}
	}

	for _, v := range s.Equals {
		if err := validateString(v); err != nil {
			return err
		}
	}

	return nil
}

type GRPCHealthRequest struct {
	Service string `yaml:"service" json:"service"`
}

type ServiceSet struct {
	Services []Service `yaml:"services" json:"services" validate:"required,min=1,dive"`
}

type Service struct {
	ID          string `yaml:"id"          json:"id"          validate:"required,min=1"`
	Name        string `yaml:"name"        json:"name"        validate:"required,min=1"`
	Description string `yaml:"description" json:"description" validate:"omitempty,min=1"`
}

//nolint:lll
type CheckFields struct {
	StatusImpact     StatusImpact  `yaml:"status_impact"     json:"status_impact"     validate:"required,oneof=none minor major critical"`
	Name             string        `yaml:"name"              json:"name"              validate:"required,min=1"`
	Service          string        `yaml:"service"           json:"service"           validate:"required,min=1"`
	Type             CheckType     `yaml:"type"              json:"type"              validate:"required,oneof=http tcp grpc tls dns"`
	ID               string        `yaml:"id"                json:"id"                validate:"required,min=1"`
	AllowedBuckets   []string      `yaml:"allowed_buckets"   json:"allowed_buckets"   validate:"omitempty,dive,oneof=second minute hour day"`
	Jitter           time.Duration `yaml:"jitter"            json:"jitter"            validate:"gte=0"`
	FailureThreshold int           `yaml:"failure_threshold" json:"failure_threshold" validate:"required,gte=1"`
	Interval         time.Duration `yaml:"interval"          json:"interval"          validate:"required,gt=0ms"`
	Timeout          time.Duration `yaml:"timeout"           json:"timeout"           validate:"required,gt=0ms"`
	Retries          int           `yaml:"retries"           json:"retries"           validate:"required,gte=0"`
	Enabled          bool          `yaml:"enabled"           json:"enabled"`
}

type SpecField[T any] struct {
	Spec T `yaml:"spec" json:"spec" validate:"required"`
}

func (s *SpecField[T]) Validate() error {
	v, ok := any(&s.Spec).(Validator)
	if !ok {
		return nil
	}

	return v.Validate()
}

type Check struct {
	SpecField[any] `yaml:",inline" json:",inline"`
	CheckFields    `yaml:",inline" json:",inline"`
}

type TypedCheck[T any] struct {
	SpecField[T] `yaml:",inline" json:",inline"`
	CheckFields  `yaml:",inline" json:",inline"`
}

func (c *TypedCheck[T]) Validate() error {
	v, ok := any(&c.Spec).(Validator)
	if !ok {
		return nil
	}

	return v.Validate()
}

type CheckSet struct {
	Checks []Check `yaml:"checks" json:"checks" validate:"required,min=1,dive"`
}

type HttpSpec struct {
	Headers         map[string]string `yaml:"headers"          json:"headers"          validate:"omitempty,min=1"`
	Payload         map[string]string `yaml:"payload"          json:"payload"          validate:"omitempty,min=1"`
	ExpectedBody    *StringExpect     `yaml:"expected_body"    json:"expected_body"    validate:"omitempty"`
	URL             string            `yaml:"url"              json:"url"              validate:"required,env_url"`
	Method          string            `yaml:"method"           json:"method"           validate:"required,env_http_method"`
	SuccessCodes    []int             `yaml:"success_codes"    json:"success_codes"    validate:"required,min=1"`
	FollowRedirects bool              `yaml:"follow_redirects" json:"follow_redirects" validate:"omitempty"`
}

func (o *HttpSpec) Validate() error {
	if err := validateMap(o.Headers); err != nil {
		return err
	}

	if err := validateMap(o.Payload); err != nil {
		return err
	}

	if err := validateString(o.Method); err != nil {
		return err
	}

	if o.ExpectedBody != nil {
		if err := o.ExpectedBody.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type TCPSpec struct {
	Expect *StringExpect `yaml:"expect" json:"expect" validate:"omitempty"`
	Host   string        `yaml:"host"   json:"host"   validate:"required,env_host"`
	Send   string        `yaml:"send"   json:"send"   validate:"omitempty,min=1"`
	Port   int           `yaml:"port"   json:"port"   validate:"required,gte=1,lte=65535"`
}

func (o *TCPSpec) Validate() error {
	if o.Expect != nil {
		if err := o.Expect.Validate(); err != nil {
			return err
		}
	}

	if o.Send != "" {
		if err := validateString(o.Send); err != nil {
			return err
		}
	}

	return nil
}

type GRPCSpec struct {
	Metadata map[string]string  `yaml:"metadata" json:"metadata" validate:"omitempty,min=1"`
	Host     string             `yaml:"host"     json:"host"     validate:"required,env_host"`
	Service  GRPCHealthService  `yaml:"service"  json:"service"  validate:"required,oneof=grpc.health.v1.Health"`
	Method   GRPCHealthMethod   `yaml:"method"   json:"method"   validate:"required,oneof=Check"`
	Request  *GRPCHealthRequest `yaml:"request"  json:"request"  validate:"omitempty"`

	//nolint:lll,nolintlint
	ExpectedHealthStatus GRPCHealthStatus `yaml:"expected_status" json:"expected_status" validate:"required,oneof=UNKNOWN SERVING NOT_SERVING SERVICE_UNKNOWN"`

	Port int `yaml:"port" json:"port" validate:"required,gte=1,lte=65535"`
}

func (o *GRPCSpec) Validate() error {
	if err := validateMap(o.Metadata); err != nil {
		return err
	}

	if o.Request != nil {
		if o.Request.Service != "" {
			if err := validateString(o.Request.Service); err != nil {
				return err
			}
		}
	}

	return nil
}

type DNSSpec struct {
	Expect     *DNSExpect `yaml:"expect"      json:"expect"      validate:"omitempty"`
	Server     string     `yaml:"server"      json:"server"      validate:"omitempty,env_host"`
	Name       string     `yaml:"name"        json:"name"        validate:"required,min=1"`
	RecordType RecordType `yaml:"record_type" json:"record_type" validate:"required,oneof=A AAAA CNAME TXT MX NS SRV"`
}

func (o *DNSSpec) Validate() error {
	if o.Expect != nil {
		if err := o.Expect.Validate(); err != nil {
			return err
		}
	}

	if err := validateString(o.Name); err != nil {
		return err
	}

	return nil
}

type TLSSpec struct {
	Host        string        `yaml:"host"         json:"host"         validate:"required,env_host"`
	ServerName  string        `yaml:"server_name"  json:"server_name"  validate:"omitempty,env_host"`
	Port        int           `yaml:"port"         json:"port"         validate:"required,gte=1,lte=65535"`
	MinValidity time.Duration `yaml:"min_validity" json:"min_validity" validate:"required,gte=1h"`
}

func (o *TLSSpec) Validate() error {
	if o.ServerName != "" {
		if err := validateString(o.ServerName); err != nil {
			return err
		}
	}

	return nil
}

//nolint:gocyclo,cyclop
func (o *HttpSpec) Cmp(other *HttpSpec) bool {
	if o == nil || other == nil {
		return o == other
	}

	if o.ExpectedBody == nil && other.ExpectedBody != nil {
		return false
	}

	if o.ExpectedBody != nil && other.ExpectedBody == nil {
		return false
	}

	if o.ExpectedBody != nil && other.ExpectedBody != nil {
		if o.ExpectedBody.Equals != other.ExpectedBody.Equals {
			return false
		}

		if o.ExpectedBody.Contains != other.ExpectedBody.Contains {
			return false
		}
	}

	return o.URL == other.URL &&
		o.Method == other.Method &&
		o.FollowRedirects == other.FollowRedirects &&
		slices.Equal(o.SuccessCodes, other.SuccessCodes) &&
		maps.Equal(o.Headers, other.Headers) &&
		reflect.DeepEqual(o.Payload, other.Payload)
}

func (o *TCPSpec) Cmp(other *TCPSpec) bool {
	if o == nil || other == nil {
		return o == other
	}

	if o.Expect == nil && other.Expect != nil {
		return false
	}

	if o.Expect != nil && other.Expect == nil {
		return false
	}

	if o.Expect != nil && other.Expect != nil {
		if o.Expect.Equals != other.Expect.Equals {
			return false
		}

		if o.Expect.Contains != other.Expect.Contains {
			return false
		}
	}

	return o.Host == other.Host &&
		o.Port == other.Port &&
		o.Send == other.Send
}

func (o *GRPCSpec) Cmp(other *GRPCSpec) bool {
	if o == nil || other == nil {
		return o == other
	}

	if o.Request == nil && other.Request != nil {
		return false
	}

	if o.Request != nil && other.Request == nil {
		return false
	}

	if o.Request != nil && other.Request != nil {
		if o.Request.Service != other.Request.Service {
			return false
		}
	}

	return o.Host == other.Host &&
		o.Port == other.Port &&
		o.Service == other.Service &&
		o.Method == other.Method &&
		o.ExpectedHealthStatus == other.ExpectedHealthStatus &&
		maps.Equal(o.Metadata, other.Metadata)
}

//nolint:gocognit,gocyclo,cyclop
func (o *DNSSpec) Cmp(other *DNSSpec) bool {
	if o == nil || other == nil {
		return o == other
	}

	if o.Expect == nil && other.Expect != nil {
		return false
	}

	if o.Expect != nil && other.Expect == nil {
		return false
	}

	if o.Expect != nil && other.Expect != nil {
		if len(o.Expect.Equals) != len(other.Expect.Equals) {
			return false
		}

		for i := range o.Expect.Equals {
			if o.Expect.Equals[i] != other.Expect.Equals[i] {
				return false
			}
		}

		if len(o.Expect.Contains) != len(other.Expect.Contains) {
			return false
		}

		for i := range o.Expect.Contains {
			if o.Expect.Contains[i] != other.Expect.Contains[i] {
				return false
			}
		}
	}

	return o.Server == other.Server && o.Name == other.Name && o.RecordType == other.RecordType
}

func (o *TLSSpec) Cmp(other *TLSSpec) bool {
	if o == nil || other == nil {
		return o == other
	}

	return o.Host == other.Host &&
		o.ServerName == other.ServerName &&
		o.Port == other.Port &&
		o.MinValidity == other.MinValidity
}
