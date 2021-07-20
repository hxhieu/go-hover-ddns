package main

type HoverDnsDomainEntry struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	Ttl       int32  `json:"ttl"`
	IsDefault bool   `json:"is_default"`
	CanRevert bool   `json:"can_revert"`
}

type HoverDnsDomain struct {
	Name    string                `json:"domain_name"`
	Id      string                `json:"id"`
	Active  bool                  `json:"active"`
	Entries []HoverDnsDomainEntry `json:"entries"`
}

type HoverDnsResult struct {
	Succeeded bool             `json:"succeeded"`
	Error     string           `json:"error"`
	ErrorCode string           `json:"error_code"`
	Domains   []HoverDnsDomain `json:"domains"`
}
