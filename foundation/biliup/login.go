package biliup

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/tidwall/gjson"
	"rsc.io/qr"
)

const (
	qrcodeNewAPI  = "https://passport.bilibili.com/x/passport-tv-login/qrcode/auth_code"
	qrcodePollAPI = "http://passport.bilibili.com/x/passport-tv-login/qrcode/poll"
	appkey        = "4409e2ce8ffd12b8"
	appsec        = "59b43e04ad6965f34319062b478f83dd"
	filename      = "cookies.json"
)

func Login() error {
	qc, err := newQrcode()
	if err != nil {
		return err
	}
	qc.Print()
	fmt.Printf("请扫码登陆或将以下链接复制到手机B站打开:\n%s\n", qc.qrcodeURL)
	fmt.Println("登陆成功后请按回车键继续:")
	fmt.Scanf("%s", "")
	return qc.Poll()
}

type qrcode struct {
	qrcodeURL string
	authCode  string
}

func newQrcode() (*qrcode, error) {
	data := make(url.Values)
	data.Add("local_id", "0")
	data.Add("ts", fmt.Sprint(time.Now().Unix()))
	data.Add("appkey", appkey)

	bytesToSign := []byte(data.Encode() + appsec)
	hasher := md5.New()
	hasher.Write(bytesToSign)
	signBytes := hasher.Sum(nil)
	sign := hex.EncodeToString(signBytes)
	data.Add("sign", sign)

	reqBody := strings.NewReader(data.Encode())
	req, err := http.NewRequest(http.MethodPost, qrcodeNewAPI, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	code := gjson.Parse(string(respBody)).Get("code").Int()
	if code != 0 {
		return nil, fmt.Errorf("new qrcode failed, err msg: %s", respBody)
	}

	qc := &qrcode{
		qrcodeURL: gjson.Parse(string(respBody)).Get("data.url").String(),
		authCode:  gjson.Parse(string(respBody)).Get("data.auth_code").String(),
	}

	return qc, nil
}

func (qc *qrcode) Print() {
	config := qrterminal.Config{
		Level:          qr.M,
		Writer:         os.Stdout,
		HalfBlocks:     true,
		BlackChar:      qrterminal.BLACK_BLACK,
		WhiteBlackChar: qrterminal.WHITE_BLACK,
		WhiteChar:      qrterminal.WHITE_WHITE,
		BlackWhiteChar: qrterminal.BLACK_WHITE,
		QuietZone:      0,
	}
	qrterminal.GenerateWithConfig(qc.qrcodeURL, config)
}

func (qc *qrcode) Poll() error {
	data := make(url.Values)
	data.Add("auth_code", qc.authCode)
	data.Add("local_id", "0")
	data.Add("ts", fmt.Sprint(time.Now().Unix()))
	data.Add("appkey", appkey)

	bytesToSign := []byte(data.Encode() + appsec)
	hasher := md5.New()
	hasher.Write(bytesToSign)
	signBytes := hasher.Sum(nil)
	sign := hex.EncodeToString(signBytes)
	data.Add("sign", sign)

	reqBody := strings.NewReader(data.Encode())
	req, err := http.NewRequest(http.MethodPost, qrcodePollAPI, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	code := gjson.Parse(string(respBody)).Get("code").Int()
	if code != 0 {
		return fmt.Errorf("poll qrcode failed, err msg: %s", respBody)
	}

	fmt.Println("登录成功")
	cookies := gjson.Parse(string(respBody)).Get("data").String()
	err = os.WriteFile(filename, []byte(cookies), os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Println("登录成功, 信息已保存在", filename)
	return nil
}
