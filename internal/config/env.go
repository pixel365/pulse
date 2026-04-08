package config

import (
	"fmt"
	"os"
)

func parseEnv(value string) string {
	return envRefPattern.ReplaceAllStringFunc(value, func(match string) string {
		name := envRefPattern.FindStringSubmatch(match)[1]
		if env, ok := os.LookupEnv(name); ok {
			return env
		}

		return match
	})
}

func resolveStringEnv(value string) (string, error) {
	var resolveErr error
	result := envRefPattern.ReplaceAllStringFunc(value, func(match string) string {
		name := envRefPattern.FindStringSubmatch(match)[1]
		env, ok := os.LookupEnv(name)
		if !ok {
			resolveErr = fmt.Errorf("environment variable %s is not set", name)
			return match
		}

		return env
	})

	if resolveErr != nil {
		return "", resolveErr
	}

	return result, nil
}

func resolveStringMapEnv(values map[string]string) (map[string]string, error) {
	if values == nil {
		return nil, nil
	}

	result := make(map[string]string, len(values))
	for key, value := range values {
		resolved, err := resolveStringEnv(value)
		if err != nil {
			return nil, fmt.Errorf("could not resolve %s: %w", key, err)
		}
		result[key] = resolved
	}

	return result, nil
}

func resolveStringExpectEnv(expect *StringExpect) (*StringExpect, error) {
	if expect == nil {
		return nil, nil
	}

	result := *expect
	var err error

	if result.Contains != "" {
		result.Contains, err = resolveStringEnv(result.Contains)
		if err != nil {
			return nil, err
		}
	}

	if result.Equals != "" {
		result.Equals, err = resolveStringEnv(result.Equals)
		if err != nil {
			return nil, err
		}
	}

	return &result, nil
}

func resolveDNSExpectEnv(expect *DNSExpect) (*DNSExpect, error) {
	if expect == nil {
		return nil, nil
	}

	result := DNSExpect{
		Contains: make([]string, 0, len(expect.Contains)),
		Equals:   make([]string, 0, len(expect.Equals)),
	}

	for _, value := range expect.Contains {
		resolved, err := resolveStringEnv(value)
		if err != nil {
			return nil, err
		}
		result.Contains = append(result.Contains, resolved)
	}

	for _, value := range expect.Equals {
		resolved, err := resolveStringEnv(value)
		if err != nil {
			return nil, err
		}
		result.Equals = append(result.Equals, resolved)
	}

	return &result, nil
}

func ResolveHTTPSpecEnv(spec HttpSpec) (HttpSpec, error) {
	var err error

	spec.URL, err = resolveStringEnv(spec.URL)
	if err != nil {
		return HttpSpec{}, err
	}

	spec.Method, err = resolveStringEnv(spec.Method)
	if err != nil {
		return HttpSpec{}, err
	}

	spec.Headers, err = resolveStringMapEnv(spec.Headers)
	if err != nil {
		return HttpSpec{}, err
	}

	spec.Payload, err = resolveStringMapEnv(spec.Payload)
	if err != nil {
		return HttpSpec{}, err
	}

	spec.ExpectedBody, err = resolveStringExpectEnv(spec.ExpectedBody)
	if err != nil {
		return HttpSpec{}, err
	}

	return spec, nil
}

func ResolveTCPSpecEnv(spec TCPSpec) (TCPSpec, error) {
	var err error

	spec.Host, err = resolveStringEnv(spec.Host)
	if err != nil {
		return TCPSpec{}, err
	}

	spec.Send, err = resolveStringEnv(spec.Send)
	if err != nil {
		return TCPSpec{}, err
	}

	spec.Expect, err = resolveStringExpectEnv(spec.Expect)
	if err != nil {
		return TCPSpec{}, err
	}

	return spec, nil
}

func ResolveGRPCSpecEnv(spec GRPCSpec) (GRPCSpec, error) {
	var err error

	spec.Host, err = resolveStringEnv(spec.Host)
	if err != nil {
		return GRPCSpec{}, err
	}

	spec.Metadata, err = resolveStringMapEnv(spec.Metadata)
	if err != nil {
		return GRPCSpec{}, err
	}

	if spec.Request != nil {
		request := *spec.Request
		request.Service, err = resolveStringEnv(request.Service)
		if err != nil {
			return GRPCSpec{}, err
		}
		spec.Request = &request
	}

	return spec, nil
}

func ResolveDNSSpecEnv(spec DNSSpec) (DNSSpec, error) {
	var err error

	spec.Server, err = resolveStringEnv(spec.Server)
	if err != nil {
		return DNSSpec{}, err
	}

	spec.Name, err = resolveStringEnv(spec.Name)
	if err != nil {
		return DNSSpec{}, err
	}

	spec.Expect, err = resolveDNSExpectEnv(spec.Expect)
	if err != nil {
		return DNSSpec{}, err
	}

	return spec, nil
}

func ResolveTLSSpecEnv(spec TLSSpec) (TLSSpec, error) {
	var err error

	spec.Host, err = resolveStringEnv(spec.Host)
	if err != nil {
		return TLSSpec{}, err
	}

	spec.ServerName, err = resolveStringEnv(spec.ServerName)
	if err != nil {
		return TLSSpec{}, err
	}

	return spec, nil
}
