package gpdns

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

// API reference:
//   https://developers.google.com/speed/public-dns/docs/dns-over-https#api_specification

var apiURI = "https://dns.google.com/resolve"

type question struct {
	Name string `json:"name"`
	Type uint16 `json:"type"`
}

type answer struct {
	Name string `json:"name"`
	Type uint16 `json:"type"`
	TTL  int    `json:"TTL"`
	Data string `json:"data"`
}

type response struct {
	Status           int        `json:"Status"`
	TC               bool       `json:"TC"`
	RD               bool       `json:"RD"`
	RA               bool       `json:"RA"`
	AD               bool       `json:"AD"`
	CD               bool       `json:"CD"`
	Question         []question `json:"Question"`
	Answer           []answer   `json:"Answer"`
	Additional       []answer   `json:"Additional"`
	EDNSClientSubnet string     `json:"edns_client_subnet"`
	Comment          string     `json:"Comment"`
}

type Client struct {
	ednsSubnet string
	client     *http.Client
}

func parseQuestion(qs []question) []dns.Question {
	dnsQs := []dns.Question{}
	for _, q := range qs {
		dnsQs = append(dnsQs, dns.Question{
			Name:   q.Name,
			Qtype:  q.Type,
			Qclass: dns.ClassINET, // Only supports IN
		})
	}
	return dnsQs
}

func parseAnswer(as []answer) ([]dns.RR, error) {
	dnsAs := []dns.RR{}
	for _, a := range as {
		rr, err := dns.NewRR(fmt.Sprintf("%s %d IN %s %s", a.Name, a.TTL, dns.TypeToString[a.Type], a.Data))
		if err != nil {
			return nil, err
		}
		dnsAs = append(dnsAs, rr)
	}
	return dnsAs, nil
}

// Exchange replicates the miekg/dns.Client.Exchange method
func (c *Client) Exchange(msg *dns.Msg, _ string) (*dns.Msg, time.Duration, error) {
	started := time.Now()
	req, err := http.NewRequest("GET", apiURI, nil)
	if err != nil {
		return nil, 0, err
	}
	if len(msg.Question) == 0 {
		return nil, 0, errors.New("gpdns: must ask a question")
	}
	query := make(url.Values)
	query.Add("name", msg.Question[0].Name)
	query.Add("type", strconv.Itoa(int(msg.Question[0].Qtype)))
	if msg.CheckingDisabled {
		query.Add("cd", "true")
	}
	if c.ednsSubnet != "" {
		query.Add("edns_client_subnet", c.ednsSubnet)
	}
	req.URL.RawQuery = query.Encode()
	apiResp, err := c.client.Do(req)
	if err != nil {
		return nil, time.Since(started), err
	}
	defer apiResp.Body.Close()
	body, err := ioutil.ReadAll(apiResp.Body)
	if err != nil {
		return nil, time.Since(started), err
	}
	var respObj response
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return nil, time.Since(started), err
	}
	resp := new(dns.Msg)
	resp.SetReply(msg)
	resp.Rcode = respObj.Status
	resp.Truncated = respObj.TC
	resp.RecursionDesired = respObj.RD
	resp.RecursionAvailable = respObj.RA
	resp.AuthenticatedData = respObj.AD
	resp.CheckingDisabled = respObj.CD
	resp.Question = parseQuestion(respObj.Question)
	resp.Answer, err = parseAnswer(respObj.Answer)
	if err != nil {
		return nil, time.Since(started), err
	}
	resp.Extra, err = parseAnswer(respObj.Additional)
	if err != nil {
		return nil, time.Since(started), err
	}
	return resp, time.Since(started), nil
}
