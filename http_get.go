package main

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	BiliPassHost = "https://xxx.bilibili.com"
	BiliArcHost  = "http://xxx.bilibili.com"
	AppKey       = "xxxxxxxxxxxxx"
	AppSecret    = "xxxxxxxxxxxxxxxxxxxxxxxxx"

	BiliGetKey = BiliPassHost + "/api/xxx"
	BiliLogin  = BiliPassHost + "/api/xxx"
)

var (
	userName  string
	accessKey string
	Hash      string
	PubKey    string
)

func addSignParam(v url.Values) string {
	v.Set("appkey", AppKey)
	v.Set("ts", strconv.FormatInt(time.Now().Unix(), 10))
	sign := md5.Sum([]byte(v.Encode() + AppSecret))
	param := v.Encode()
	param += fmt.Sprintf("&sign=%x", sign)
	return param
}

func getKey() (string, string) {
	get_key_url := BiliGetKey + "?" + addSignParam(url.Values{})
	//fmt.Println(get_key_url)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	r, err := client.Get(get_key_url)
	if err != nil {
		fmt.Println(err.Error())
		return "", ""
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	type Resp struct {
		Ts   int    `json:"ts"`
		Hash string `json:"hash"`
		Key  string `json:"key"`
	}
	var res Resp
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println(err.Error())
		return "", ""
	}
	return res.Hash, res.Key
}

func rsaEncrypt(key []byte, data []byte) ([]byte, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("Public key error!")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, data)
}

func logIn(user string, pwd string) string {
	p, _ := rsaEncrypt([]byte(PubKey), []byte(Hash+pwd))
	pwd = base64.StdEncoding.EncodeToString(p)

	login_url := BiliLogin + "?"
	v := url.Values{}
	v.Set("userid", user)
	v.Set("pwd", pwd)
	login_url += addSignParam(v)
	//fmt.Println(login_url)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	r, err := client.Get(login_url)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	fmt.Printf("%s", body)
	return ""
}

func main() {
	fmt.Println("Please Input UserName:")
	fmt.Scanln(&userName)
	fmt.Println("Please Input Password:")
	var passWord string
	fmt.Scanln(&passWord)

	// Login
	Hash, PubKey = getKey()
	accessKey = logIn(userName, passWord)

	fmt.Printf("Hi, %s!\n", userName)
}
