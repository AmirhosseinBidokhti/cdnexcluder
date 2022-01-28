package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var CDNProviders = map[string]string{
  "AMAZON":  "https://ip-ranges.amazonaws.com/ip-ranges.json",
  "CLOUDFLARE": "https://www.cloudflare.com/ips-v4",
  "GOOGLE": "https://www.gstatic.com/ipranges/cloud.json",
  "FASTLY": "https://api.fastly.com/public-ip-list",
  "CACHEFLY": "https://cachefly.cachefly.net/ips/rproxy.txt",
}

type AmazonCDNResponse struct {
  SyncToken string `json:"syncToken"`
  CreateDate string `json:"createDate"`
  Prefixes []struct {
    Ip_prefix string `json:"ip_prefix"`
    Region string `json:"region"`
    Service string `json:"service"`
    Network_border_group string `json:"network_border_group"`
  } `json:"prefixes"`
}

type FastlyCDNResponse struct {
  Addresses []string `json:"Addresses"`
  Ipv6_addresses []string `json:"ipv6_addresses"`
}

type GoogleCDNResponse struct {
  SyncToken string `json:"syncToken"`
  CreationTime string `json:"creationTime"`
  Prefixes []struct {
    Ipv4Prefix string `json:"ipv4Prefix"`
    Ipv6Prefix string `json:"ipv6Prefix"`
    Service string `json:"service"`
    Scope string `json:"scope"`
  } `json:"prefixes"`
}


func AmazonCDN()[]string {
  var ips []string

  resp, err := http.Get(CDNProviders["AMAZON"])
  if err != nil {
      fmt.Println("No response from request")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body) // response body is []byte
  
  var result AmazonCDNResponse
  if err := json.Unmarshal(body, &result); err != nil {  // Parse []byte to the go struct pointer
      fmt.Println("Can not unmarshal JSON")
  }
  
  for _, rec := range result.Prefixes {
      ips = append(ips, rec.Ip_prefix)
  }
  return ips;
}

func FastlyCDN()[]string {
  var ips []string

  resp, err := http.Get(CDNProviders["FASTLY"])
  if err != nil {
      fmt.Println("No response from request")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body) 
  
  var result FastlyCDNResponse
  if err := json.Unmarshal(body, &result); err != nil {  
      fmt.Println("Can not unmarshal JSON")
  }
  
  for _, rec := range result.Addresses {
      ips = append(ips, rec)
    
  }
  return ips;
}

func GoogleCDN()[]string {
  var ips []string

  resp, err := http.Get(CDNProviders["GOOGLE"])
  if err != nil {
      fmt.Println("No response from request")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body) 
  
  var result GoogleCDNResponse
  if err := json.Unmarshal(body, &result); err != nil {  
      fmt.Println("Can not unmarshal JSON")
  }
  
  for _, rec := range result.Prefixes {
      if len(rec.Ipv4Prefix) !=0 {
      ips = append(ips, rec.Ipv4Prefix)
      }
  }
  return ips;
}

func CloudflareCDN()[]string {
  var ips []string

  resp, err := http.Get(CDNProviders["CLOUDFLARE"])
  if err != nil {
      fmt.Println("No response from request")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body) 
  
  ips = strings.Split(string(body), "\n")
  return ips;
}

func CacheflyCDN()[]string {
  var ips []string

  resp, err := http.Get(CDNProviders["CACHEFLY"])
  if err != nil {
      fmt.Println("No response from request")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body) 
  
  ips = strings.Split(string(body), "\n")
  return ips;
}


func main() {

  var CDNIPs []string

  //fmt.Print(amazon)
  // fmt.Println(amazon)
  // fmt.Println(fastly)
  // fmt.Println(google)

  var amazon = AmazonCDN()
  var fastly = FastlyCDN()
  var google = GoogleCDN()
  var cloudflare = CloudflareCDN()
  var cachefly = CacheflyCDN()

  CDNIPs = append(CDNIPs, amazon...)
  CDNIPs = append(CDNIPs, fastly...)
  CDNIPs = append(CDNIPs, google...)
  CDNIPs = append(CDNIPs, cloudflare...)
  CDNIPs = append(CDNIPs, cachefly...)


  for _, rec := range CDNIPs {
    fmt.Println(rec)
  }

  fmt.Print(len(CDNIPs))

}