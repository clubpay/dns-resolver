package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/miekg/dns"
	"net/http"
)

var (
	dnsTypeMap = map[string]uint16{
		"txt":   dns.TypeTXT,
		"a":     dns.TypeA,
		"aaaa":  dns.TypeAAAA,
		"cname": dns.TypeCNAME,
		"mx":    dns.TypeMX,
	}
)

type Record struct {
	Name string `json:"name"`
	TTL  uint32 `json:"ttl"`
	Type string `json:"type"`
	TXT  struct {
		Targets []string `json:"targets,omitempty"`
	} `json:"txt,omitempty"`
	MX struct {
		Target     string `json:"target,omitempty"`
		Preference uint16 `json:"preference,omitempty"`
	} `json:"mx,omitempty"`
	Value string `json:"value,omitempty"`
}

type Response struct {
	Records []Record `json:"records"`
}

func main() {
	app := fiber.New()
	dnsClient := &dns.Client{
		SingleInflight: true,
	}

	app.Get("/:typ/:domain", func(c *fiber.Ctx) error {
		req := &dns.Msg{}
		req.SetQuestion(c.Params("domain"), dnsTypeMap[c.Params("typ")])
		req.RecursionDesired = true
		r, _, err := dnsClient.Exchange(req, "1.1.1.1:53")
		if err != nil {
			fmt.Println(err)
			return c.SendStatus(http.StatusNotFound)
		}
		if r.Rcode != dns.RcodeSuccess {
			return c.SendStatus(http.StatusExpectationFailed)
		}
		res := &Response{}
		for _, a := range r.Answer {
			hdr := a.Header()
			if hdr == nil {
				continue
			}
			record := Record{
				Name: hdr.Name,
				TTL:  hdr.Ttl,
				Type: dns.Type(hdr.Rrtype).String(),
			}

			switch a := a.(type) {
			case *dns.CNAME:
				record.Value = a.Target
			case *dns.MX:
				record.MX.Preference = a.Preference
				record.MX.Target = a.Mx
			case *dns.A:
				record.Value = a.A.String()
			case *dns.TXT:
				record.TXT.Targets = a.Txt
			case *dns.AAAA:
				record.Value = a.AAAA.String()
			}

			res.Records = append(res.Records, record)
		}

		if len(res.Records) == 0 {
			return c.SendStatus(http.StatusNotFound)
		}

		resBytes, _ := json.Marshal(res)
		return c.Send(resBytes)
	})

	_ = app.Listen(":80")
}
