package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// Default constructs the production engine using passive sources only:
// public DNS resolution, public Certificate Transparency logs (crt.sh) and
// reverse DNS for authorized CIDR ranges.
func Default(timeout time.Duration) Engine {
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	resolver := &net.Resolver{}
	client := &http.Client{Timeout: timeout}
	return New(2000,
		dnsSource{resolver: resolver},
		ctLogSource{client: client, endpoint: "https://crt.sh"},
		reverseDNSSource{resolver: resolver, maxHosts: 256},
	)
}

func normalizeHost(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if i := strings.Index(v, "://"); i >= 0 {
		v = v[i+3:]
	}
	v = strings.TrimSuffix(v, "/")
	v = strings.TrimSuffix(v, ".")
	v = strings.TrimPrefix(v, "*.")
	return v
}

// dnsSource performs passive DNS resolution (A/AAAA/CNAME/MX/NS/TXT).
type dnsSource struct{ resolver *net.Resolver }

func (dnsSource) Name() string { return "dns" }

func (dnsSource) Supports(t models.AssetType) bool {
	return t == models.AssetDomain || t == models.AssetSubdomain
}

func (s dnsSource) Discover(ctx context.Context, in Input) ([]Finding, error) {
	host := normalizeHost(in.Value)
	if host == "" {
		return nil, fmt.Errorf("invalid host %q", in.Value)
	}
	var out []Finding
	record := func(recordType, value string) {
		out = append(out, Finding{
			Type:   models.AssetDNSRecord,
			Value:  fmt.Sprintf("%s %s %s", host, recordType, value),
			Source: s.Name(),
			Attributes: map[string]any{
				"name":        host,
				"record_type": recordType,
				"value":       value,
			},
		})
	}

	if ips, err := s.resolver.LookupIP(ctx, "ip", host); err == nil {
		for _, ip := range ips {
			rt := "A"
			if ip.To4() == nil {
				rt = "AAAA"
			}
			record(rt, ip.String())
		}
	}
	if cname, err := s.resolver.LookupCNAME(ctx, host); err == nil {
		cname = strings.TrimSuffix(strings.ToLower(cname), ".")
		if cname != "" && cname != host {
			record("CNAME", cname)
		}
	}
	if mxs, err := s.resolver.LookupMX(ctx, host); err == nil {
		for _, mx := range mxs {
			record("MX", strings.TrimSuffix(mx.Host, "."))
		}
	}
	if nss, err := s.resolver.LookupNS(ctx, host); err == nil {
		for _, ns := range nss {
			record("NS", strings.TrimSuffix(ns.Host, "."))
		}
	}
	if txts, err := s.resolver.LookupTXT(ctx, host); err == nil {
		for _, txt := range txts {
			record("TXT", txt)
		}
	}
	return out, nil
}

// ctLogSource queries public Certificate Transparency logs (crt.sh) to passively
// enumerate subdomains and observed certificates.
type ctLogSource struct {
	client   *http.Client
	endpoint string
}

func (ctLogSource) Name() string { return "ct_logs" }

func (ctLogSource) Supports(t models.AssetType) bool {
	return t == models.AssetDomain || t == models.AssetSubdomain
}

type ctEntry struct {
	NameValue  string `json:"name_value"`
	CommonName string `json:"common_name"`
	IssuerName string `json:"issuer_name"`
	NotBefore  string `json:"not_before"`
	NotAfter   string `json:"not_after"`
	Serial     string `json:"serial_number"`
}

func (s ctLogSource) Discover(ctx context.Context, in Input) ([]Finding, error) {
	domain := normalizeHost(in.Value)
	if domain == "" {
		return nil, fmt.Errorf("invalid domain %q", in.Value)
	}
	url := fmt.Sprintf("%s/?q=%%25.%s&output=json", s.endpoint, domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RedIntel-Sentinel/passive-discovery")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ct log returned status %d", resp.StatusCode)
	}

	var entries []ctEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	var out []Finding
	for _, e := range entries {
		for _, name := range strings.Split(e.NameValue, "\n") {
			name = normalizeHost(name)
			if name == "" || strings.ContainsAny(name, " *") {
				continue
			}
			if name == domain || strings.HasSuffix(name, "."+domain) {
				out = append(out, Finding{
					Type:       models.AssetSubdomain,
					Value:      name,
					Source:     s.Name(),
					Attributes: map[string]any{"discovered_via": "certificate_transparency"},
				})
			}
		}
		cn := normalizeHost(e.CommonName)
		if cn != "" {
			out = append(out, Finding{
				Type:   models.AssetCertificate,
				Value:  fmt.Sprintf("%s (%s)", cn, strings.TrimSpace(e.Serial)),
				Source: s.Name(),
				Attributes: map[string]any{
					"common_name": cn,
					"issuer":      strings.TrimSpace(e.IssuerName),
					"not_before":  strings.TrimSpace(e.NotBefore),
					"not_after":   strings.TrimSpace(e.NotAfter),
					"serial":      strings.TrimSpace(e.Serial),
				},
			})
		}
	}
	return out, nil
}

// reverseDNSSource performs passive reverse DNS (PTR) over an authorized CIDR.
type reverseDNSSource struct {
	resolver *net.Resolver
	maxHosts int
}

func (reverseDNSSource) Name() string { return "reverse_dns" }

func (reverseDNSSource) Supports(t models.AssetType) bool { return t == models.AssetCIDR }

func (s reverseDNSSource) Discover(ctx context.Context, in Input) ([]Finding, error) {
	_, ipnet, err := net.ParseCIDR(strings.TrimSpace(in.Value))
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR %q: %w", in.Value, err)
	}
	max := s.maxHosts
	if max <= 0 {
		max = 256
	}
	var out []Finding
	count := 0
	ip := cloneIP(ipnet.IP.Mask(ipnet.Mask))
	for ipnet.Contains(ip) && count < max {
		if err := ctx.Err(); err != nil {
			break
		}
		names, err := s.resolver.LookupAddr(ctx, ip.String())
		if err == nil {
			for _, n := range names {
				n = normalizeHost(n)
				if n == "" {
					continue
				}
				out = append(out, Finding{
					Type:       models.AssetDNSRecord,
					Value:      fmt.Sprintf("%s PTR %s", ip.String(), n),
					Source:     s.Name(),
					Attributes: map[string]any{"name": ip.String(), "record_type": "PTR", "value": n},
				})
				out = append(out, Finding{
					Type:       models.AssetSubdomain,
					Value:      n,
					Source:     s.Name(),
					Attributes: map[string]any{"discovered_via": "reverse_dns", "ip": ip.String()},
				})
			}
		}
		incIP(ip)
		count++
	}
	return out, nil
}

func cloneIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func incIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}
