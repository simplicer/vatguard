package infrastructure

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/antonio-villamarin/vat-validator/internal/vatvalidation/application"
)

const (
	viesURL      = "https://ec.europa.eu/taxation_customs/vies/services/checkVatService"
	viesTimeout  = 5 * time.Second
	soapAction   = ""
	soapEnvelope = `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:ec.europa.eu:taxud:vies:services:checkVat">
  <soapenv:Header/>
  <soapenv:Body>
    <urn:checkVat>
      <urn:countryCode>%s</urn:countryCode>
      <urn:vatNumber>%s</urn:vatNumber>
    </urn:checkVat>
  </soapenv:Body>
</soapenv:Envelope>`
)

// VIESClient implements application.VIESChecker using the VIES SOAP service.
type VIESClient struct {
	httpClient *http.Client
	url        string
}

// NewVIESClient creates a new VIES SOAP client with the default endpoint.
func NewVIESClient() *VIESClient {
	return &VIESClient{
		httpClient: &http.Client{
			Timeout: viesTimeout,
		},
		url: viesURL,
	}
}

// NewVIESClientWithURL creates a VIES client with a custom URL (for testing).
func NewVIESClientWithURL(url string) *VIESClient {
	return &VIESClient{
		httpClient: &http.Client{
			Timeout: viesTimeout,
		},
		url: url,
	}
}

// Check sends a SOAP request to VIES and returns the result.
func (c *VIESClient) Check(countryCode, nationalNumber string) (*application.VIESResult, error) {
	body := fmt.Sprintf(soapEnvelope, xmlEscape(countryCode), xmlEscape(nationalNumber))

	req, err := http.NewRequest(http.MethodPost, c.url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating VIES request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	if soapAction != "" {
		req.Header.Set("SOAPAction", soapAction)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling VIES service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VIES returned HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading VIES response: %w", err)
	}

	return parseVIESResponse(data)
}

// SOAP response types

type envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    body     `xml:"Body"`
}

type body struct {
	Response *checkVatResponse `xml:"checkVatResponse"`
	Fault    *soapFault        `xml:"Fault"`
}

type checkVatResponse struct {
	CountryCode string `xml:"countryCode"`
	VATNumber   string `xml:"vatNumber"`
	Valid       bool   `xml:"valid"`
	Name        string `xml:"name"`
	Address     string `xml:"address"`
}

type soapFault struct {
	FaultString string `xml:"faultstring"`
}

func parseVIESResponse(data []byte) (*application.VIESResult, error) {
	var env envelope
	if err := xml.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parsing VIES XML response: %w", err)
	}

	if env.Body.Fault != nil {
		return nil, fmt.Errorf("VIES SOAP fault: %s", env.Body.Fault.FaultString)
	}

	if env.Body.Response == nil {
		return nil, fmt.Errorf("VIES returned empty response")
	}

	r := env.Body.Response
	return &application.VIESResult{
		Valid:       r.Valid,
		CompanyName: r.Name,
		Address:     r.Address,
	}, nil
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
