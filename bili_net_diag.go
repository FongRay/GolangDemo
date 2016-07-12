package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	GetDNSHost = ".trace.term.chinacache.com"
	GetDNSURL  = "http://%s/getdns.php"
	GetIPLocal = "http://interface.bilibili.com/ip.json"
)

type IPLocal struct {
	Country  string
	Province string
	City     string
	ISP      string
}

var g_bili_result string

func (loc IPLocal) GetLocalStr() string {
	return fmt.Sprintf("Country: %s | Province: %s | City: %s | ISP: %s\n", loc.Country, loc.Province, loc.City, loc.ISP)
}

func get_ip_and_dns() {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	prefix := make([]byte, 4)
	for i := 0; i < 4; i++ {
		prefix[i] = byte(r.Int31n(26) + int32('a'))
	}
	randhost := fmt.Sprintf("%s%d%s", prefix, time.Now().UnixNano()/1000000, GetDNSHost)
	get_dns_url := fmt.Sprintf(GetDNSURL, randhost)
	fmt.Println(get_dns_url)

	v := url.Values{}
	v.Add("randhost", randhost)
	v.Add("user", "bilibili")
	fmt.Println(v.Encode())

	resp, err := http.PostForm(get_dns_url, v)
	if err != nil {
		fmt.Println("ERR:: get ip & dns err!")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s\n", body)

	type IP_ADDR struct {
		Ip       net.IP
		Dns      net.IP
		Ipaddr   string
		Dnsaddr  string
		Dnsmatch string
	}
	var resp_ip_dns IP_ADDR
	err = json.Unmarshal(body, &resp_ip_dns)
	if err != nil {
		fmt.Println("ERR:: unmarshal resp failed!")
	}
	fmt.Println(resp_ip_dns)
	ip_local := make(chan IPLocal)
	dns_local := make(chan IPLocal)
	go getlocal(resp_ip_dns.Ip, ip_local)
	go getlocal(resp_ip_dns.Dns, dns_local)
	ip_loc := <-ip_local
	dns_loc := <-dns_local

	g_bili_result += fmt.Sprintf("IP:\t%s - %s\n", resp_ip_dns.Ip, resp_ip_dns.Ipaddr)
	g_bili_result += fmt.Sprintf("\t%s\n\n", ip_loc.GetLocalStr())
	g_bili_result += fmt.Sprintf("DNS:\t%s - %s\n", resp_ip_dns.Dns, resp_ip_dns.Dnsaddr)
	g_bili_result += fmt.Sprintf("\t%s\n\n", dns_loc.GetLocalStr())
	g_bili_result += fmt.Sprintf("DNS Match:\t%s", func() string {
		if resp_ip_dns.Dnsmatch == "y" {
			return "True"
		} else {
			return "False"
		}
	}())
	fmt.Println(g_bili_result)
}

func getlocal(ip net.IP, loc chan IPLocal) {
	get_local_url := GetIPLocal + "?ip=" + fmt.Sprintf("%s", ip)
	fmt.Println(get_local_url)
	resp, err := http.Get(get_local_url)
	if err != nil {
		fmt.Printf("ERR:: get ip's local failed: %s\n", ip)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var local IPLocal
	err = json.Unmarshal(body, &local)
	loc <- local
}

func main() {
	get_ip_and_dns()
}
