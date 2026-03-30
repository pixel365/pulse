package dns

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"

	c "github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/model"

	mdns "github.com/miekg/dns"
)

func (c *Checker) execute(ctx context.Context) model.CheckExecutionResult {
	var err error

	result := model.CheckExecutionResult{
		ExecutionID: rand.Text(),
		CheckID:     c.config.ID,
		ServiceID:   c.config.Service,
		CheckType:   c.config.Type,
		Status:      model.Success,
		StartedAt:   time.Now().UTC(),
	}

	attempts := c.config.Retries + 1
	for i := 0; i < attempts; i++ {
		result.AttemptsTotal = i + 1

		if ctx.Err() != nil {
			err = ctx.Err()
			break
		}

		err = nil
		reqErr := c.request(ctx)
		if reqErr != nil {
			err = reqErr
			continue
		}
		break
	}

	result.FinishedAt = time.Now().UTC()
	result.Duration = result.FinishedAt.Sub(result.StartedAt)

	if err != nil {
		result.Status = model.Failure
	} else {
		result.Status = model.Success
	}

	return result
}

func (c *Checker) request(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	req := new(mdns.Msg)
	req.SetQuestion(
		mdns.Fqdn(c.config.Spec.Name),
		recordTypeToQType(c.config.Spec.RecordType),
	)
	req.RecursionDesired = true

	server, err := resolveServer(c.config.Spec.Server)
	if err != nil {
		return fmt.Errorf("could not resolve dns server: %w", err)
	}

	res, err := exchangeDNS(ctx, req, server)
	if err != nil {
		return fmt.Errorf("could not exchange dns request: %w", err)
	}

	if res == nil {
		return fmt.Errorf("empty dns response")
	}

	if res.Rcode != mdns.RcodeSuccess {
		return fmt.Errorf("unexpected dns response code: %s", mdns.RcodeToString[res.Rcode])
	}

	values, err := collectAnswers(res.Answer, c.config.Spec.RecordType)
	if err != nil {
		return err
	}

	if len(values) == 0 {
		return fmt.Errorf(
			"no %s records found for %s",
			c.config.Spec.RecordType,
			c.config.Spec.Name,
		)
	}

	if err = checkAnswers(c.config.Spec.Expect, c.config.Spec.RecordType, values); err != nil {
		return err
	}

	return nil
}

func resolveServer(server string) (string, error) {
	if server != "" {
		return normalizeServer(server), nil
	}

	cfg, err := mdns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return "", fmt.Errorf("could not read /etc/resolv.conf: %w", err)
	}

	if len(cfg.Servers) == 0 {
		return "", fmt.Errorf("no nameservers configured in /etc/resolv.conf")
	}

	port := cfg.Port
	if port == "" {
		port = "53"
	}

	return net.JoinHostPort(cfg.Servers[0], port), nil
}

func exchangeDNS(ctx context.Context, req *mdns.Msg, server string) (*mdns.Msg, error) {
	udpClient := &mdns.Client{Net: "udp"}
	res, _, err := udpClient.ExchangeContext(ctx, req, server)
	if err == nil && res != nil && !res.Truncated {
		return res, nil
	}

	tcpClient := &mdns.Client{Net: "tcp"}
	tcpRes, _, tcpErr := tcpClient.ExchangeContext(ctx, req, server)
	if tcpErr == nil {
		return tcpRes, nil
	}

	if err != nil {
		return nil, fmt.Errorf("udp: %w; tcp: %w", err, tcpErr)
	}

	return nil, tcpErr
}

func recordTypeToQType(recordType c.RecordType) uint16 {
	switch recordType {
	case c.ARecord:
		return mdns.TypeA
	case c.AAAARecord:
		return mdns.TypeAAAA
	case c.CNAMERecord:
		return mdns.TypeCNAME
	case c.TXTRecord:
		return mdns.TypeTXT
	case c.MXRecord:
		return mdns.TypeMX
	case c.NSRecord:
		return mdns.TypeNS
	case c.SRVRecord:
		return mdns.TypeSRV
	default:
		return 0
	}
}

func normalizeServer(server string) string {
	if _, _, err := net.SplitHostPort(server); err == nil {
		return server
	}

	return net.JoinHostPort(server, "53")
}

func collectAnswers(records []mdns.RR, recordType c.RecordType) ([]string, error) {
	values := make([]string, 0, len(records))

	for _, record := range records {
		value, ok := recordValue(record, recordType)
		if !ok {
			continue
		}

		values = append(values, normalizeValue(recordType, value))
	}

	return uniqueSorted(values), nil
}

func recordValue(record mdns.RR, recordType c.RecordType) (string, bool) {
	switch rr := record.(type) {
	case *mdns.A:
		if recordType != c.ARecord {
			return "", false
		}
		return rr.A.String(), true
	case *mdns.AAAA:
		if recordType != c.AAAARecord {
			return "", false
		}
		return rr.AAAA.String(), true
	case *mdns.CNAME:
		if recordType != c.CNAMERecord {
			return "", false
		}
		return rr.Target, true
	case *mdns.TXT:
		if recordType != c.TXTRecord {
			return "", false
		}
		return strings.Join(rr.Txt, ""), true
	case *mdns.MX:
		if recordType != c.MXRecord {
			return "", false
		}
		return fmt.Sprintf("%d %s", rr.Preference, rr.Mx), true
	case *mdns.NS:
		if recordType != c.NSRecord {
			return "", false
		}
		return rr.Ns, true
	case *mdns.SRV:
		if recordType != c.SRVRecord {
			return "", false
		}
		return fmt.Sprintf("%d %d %d %s", rr.Priority, rr.Weight, rr.Port, rr.Target), true
	}

	return "", false
}

func checkAnswers(expect *c.DNSExpect, recordType c.RecordType, actual []string) error {
	if expect == nil {
		return nil
	}

	actualSet := make(map[string]struct{}, len(actual))
	for _, value := range actual {
		actualSet[normalizeValue(recordType, value)] = struct{}{}
	}

	for _, want := range expect.Contains {
		value := normalizeValue(recordType, want)
		if _, ok := actualSet[value]; !ok {
			return fmt.Errorf("expected dns answer to contain %q", want)
		}
	}

	if len(expect.Equals) == 0 {
		return nil
	}

	expected := make([]string, 0, len(expect.Equals))
	for _, want := range expect.Equals {
		expected = append(expected, normalizeValue(recordType, want))
	}

	expected = uniqueSorted(expected)
	if len(expected) != len(actual) {
		return fmt.Errorf("dns answers mismatch: expected %v, got %v", expected, actual)
	}

	for i := range expected {
		if expected[i] != actual[i] {
			return fmt.Errorf("dns answers mismatch: expected %v, got %v", expected, actual)
		}
	}

	return nil
}

func normalizeValue(recordType c.RecordType, value string) string {
	value = strings.TrimSpace(value)

	//nolint:exhaustive
	switch recordType {
	case c.CNAMERecord, c.NSRecord, c.MXRecord, c.SRVRecord:
		return strings.TrimSuffix(strings.ToLower(value), ".")
	default:
		return value
	}
}

func uniqueSorted(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	slices.Sort(values)

	unique := values[:1]
	for i := 1; i < len(values); i++ {
		if values[i] == values[i-1] {
			continue
		}

		unique = append(unique, values[i])
	}

	return unique
}
