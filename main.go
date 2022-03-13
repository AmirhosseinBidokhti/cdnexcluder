package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

// api response structs
// https://transform.tools/json-to-go

var CDNProviders = map[string]string{
	"AMAZON":     "https://ip-ranges.amazonaws.com/ip-ranges.json",
	"CLOUDFLARE": "https://www.cloudflare.com/ips-v4",
	"GOOGLE":     "https://www.gstatic.com/ipranges/cloud.json",
	"FASTLY":     "https://api.fastly.com/public-ip-list",
	"CACHEFLY":   "https://cachefly.cachefly.net/ips/rproxy.txt",
}

var ASNs = map[string][]string{
	"AKAMAI":      {"AS12222", "AS16625"},
	"DDOSGUARD":   {"AS57724"},
	"QRATOR":      {"AS200449"},
	"STACKPATH":   {"AS12989"},
	"STORMWALL":   {"AS59796"},
	"SUCURI":      {"AS30148"},
	"X4B":         {"AS136165"},
	"CDNNETWORKS": {"AS36408"},
}

type AmazonCDNResponse struct {
	SyncToken  string `json:"syncToken"`
	CreateDate string `json:"createDate"`
	Prefixes   []struct {
		Ip_prefix            string `json:"ip_prefix"`
		Region               string `json:"region"`
		Service              string `json:"service"`
		Network_border_group string `json:"network_border_group"`
	} `json:"prefixes"`
}

type FastlyCDNResponse struct {
	Addresses      []string `json:"Addresses"`
	Ipv6_addresses []string `json:"ipv6_addresses"`
}

type GoogleCDNResponse struct {
	SyncToken    string `json:"syncToken"`
	CreationTime string `json:"creationTime"`
	Prefixes     []struct {
		Ipv4Prefix string `json:"ipv4Prefix"`
		Ipv6Prefix string `json:"ipv6Prefix"`
		Service    string `json:"service"`
		Scope      string `json:"scope"`
	} `json:"prefixes"`
}

type bgpviewResponse struct {
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
	Data          struct {
		Ipv4Prefixes []struct {
			Prefix      string `json:"prefix"`
			IP          string `json:"ip"`
			Cidr        int    `json:"cidr"`
			RoaStatus   string `json:"roa_status"`
			Name        string `json:"name"`
			Description string `json:"description"`
			CountryCode string `json:"country_code"`
			Parent      struct {
				Prefix           string `json:"prefix"`
				IP               string `json:"ip"`
				Cidr             int    `json:"cidr"`
				RirName          string `json:"rir_name"`
				AllocationStatus string `json:"allocation_status"`
			} `json:"parent"`
		} `json:"ipv4_prefixes"`
		Ipv6Prefixes []interface{} `json:"ipv6_prefixes"`
	} `json:"data"`
	Meta struct {
		TimeZone      string `json:"time_zone"`
		APIVersion    int    `json:"api_version"`
		ExecutionTime string `json:"execution_time"`
	} `json:"@meta"`
}

func AmazonCDN() []string {
	var ips []string

	resp, err := http.Get(CDNProviders["AMAZON"])
	if err != nil {
		fmt.Println("No response from request")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	var result AmazonCDNResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	for _, rec := range result.Prefixes {
		ips = append(ips, rec.Ip_prefix)
	}
	return ips
}

func FastlyCDN() []string {
	var ips []string

	resp, err := http.Get(CDNProviders["FASTLY"])
	if err != nil {
		fmt.Println("No response from request")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	var result FastlyCDNResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}

	ips = append(ips, result.Addresses...)

	return ips
}

func GoogleCDN() []string {
	var ips []string

	resp, err := http.Get(CDNProviders["GOOGLE"])
	if err != nil {
		fmt.Println("No response from request")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	var result GoogleCDNResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}

	for _, rec := range result.Prefixes {
		if len(rec.Ipv4Prefix) != 0 {
			ips = append(ips, rec.Ipv4Prefix)
		}
	}
	return ips
}

func CloudflareCDN() []string {
	var ips []string

	resp, err := http.Get(CDNProviders["CLOUDFLARE"])
	if err != nil {
		fmt.Println("No response from request")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	ips = strings.Split(string(body), "\n")
	return ips
}

func CacheflyCDN() []string {
	var ips []string

	resp, err := http.Get(CDNProviders["CACHEFLY"])
	if err != nil {
		fmt.Println("No response from request")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	ips = strings.Split(string(body), "\n")
	return ips
}
func bgpviewCheck(asnBlock []string) []string {
	var ips []string

	for _, asn := range asnBlock {
		bgpViewURL := fmt.Sprintf("https://api.bgpview.io/asn/%s/prefixes", asn)

		resp, err := http.Get(bgpViewURL)
		if err != nil {
			fmt.Println("No response from request")
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		var result bgpviewResponse
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Println("Can not unmarshal JSON")
		}

		for _, rec := range result.Data.Ipv4Prefixes {
			if len(rec.Prefix) != 0 {
				ips = append(ips, rec.Prefix)
			}
		}
	}
	return ips
}

func checkIPCIDR(ip string, cidr string) bool {

	// TODO: Check IP validity first
	// if net.ParseIP(ip) == nil {
	// 	fmt.Printf("Invalid IP Address: %s\n", ip)
	// 	return false
	// }

	_, subnet, _ := net.ParseCIDR(cidr)
	IP := net.ParseIP(ip)
	if subnet.Contains(IP) {
		//fmt.Printf("%v is in subnet %v\n", IP, subnet)
		return true
	} else {
		return false
	}
}

// https://dabase.com/e/15006/
func delete_empty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func main() {
	var ips []string
	var CDNIPs []string
	var foundCDNs []string

	var realIps bool
	var cdnIps bool

	flag.BoolVar(&realIps, "real", false, "print the real ips, CDN IPs excluded")
	flag.BoolVar(&cdnIps, "cdn", false, "print the CDN IPs")
	flag.Parse()

	var file *os.File
	file = os.Stdin

	sc := bufio.NewScanner(file)
	for sc.Scan() {
		ips = append(ips, sc.Text())
	}
	if err := sc.Err(); err != nil {
		panic(err)
	}

	var amazon = AmazonCDN()
	var fastly = FastlyCDN()
	var google = GoogleCDN()
	var cloudflare = CloudflareCDN()
	var cachefly = CacheflyCDN()

	var akamai = bgpviewCheck(ASNs["AKAMAI"])
	var ddosguard = bgpviewCheck(ASNs["DDOSGUARD"])
	var qrator = bgpviewCheck(ASNs["QRATOR"])
	var stackpath = bgpviewCheck(ASNs["STACKPATH"])
	var stormwall = bgpviewCheck(ASNs["STORMWALL"])
	var sucuri = bgpviewCheck(ASNs["SUCURI"])
	var x4b = bgpviewCheck(ASNs["X4B"])
	var cdnnetworks = bgpviewCheck(ASNs["CDNNETWORKS"])

	CDNIPs = append(CDNIPs, amazon...)
	CDNIPs = append(CDNIPs, fastly...)
	CDNIPs = append(CDNIPs, google...)
	CDNIPs = append(CDNIPs, cloudflare...)
	CDNIPs = append(CDNIPs, cachefly...)
	CDNIPs = append(CDNIPs, akamai...)
	CDNIPs = append(CDNIPs, ddosguard...)
	CDNIPs = append(CDNIPs, qrator...)
	CDNIPs = append(CDNIPs, stackpath...)
	CDNIPs = append(CDNIPs, stormwall...)
	CDNIPs = append(CDNIPs, sucuri...)
	CDNIPs = append(CDNIPs, x4b...)
	CDNIPs = append(CDNIPs, cdnnetworks...)

	CDNIPs = delete_empty(CDNIPs)

	for i := 0; i < len(ips); i++ {
		for j := 0; j < len(CDNIPs); j++ {
			if checkIPCIDR(ips[i], CDNIPs[j]) {
				foundCDNs = append(foundCDNs, ips[i])
				break
			}
		}
	}

	if realIps {
		for _, ip := range ips {
			var isCDN bool
			for _, cdn := range foundCDNs {
				if ip == cdn {
					isCDN = true
				}
			}
			if !isCDN {
				fmt.Println(ip)
			}
		}
	}

	if cdnIps {
		for _, cdnip := range foundCDNs {
			fmt.Println(cdnip)
		}
	}

}
