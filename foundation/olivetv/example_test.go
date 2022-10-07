package olivetv_test

import (
	"fmt"
	"testing"

	"github.com/go-olive/olive/foundation/olivetv"
)

func ExampleTV() {
	t, err := olivetv.New("huya", "518512")
	if err != nil {
		println(err.Error())
		return
	}

	t.Snap()
	fmt.Println(t)
}

func ExampleSetCookie() {
	douyinCookie := "__ac_nonce=06245c89100e7ab2dd536; __ac_signature=_02B4Z6wo00f01LjBMSAAAIDBwA.aJ.c4z1C44TWAAEx696;"
	t, err := olivetv.New("douyin", "600571451250", olivetv.SetCookie(douyinCookie))
	if err != nil {
		println(err.Error())
		return
	}

	t.Snap()
	fmt.Println(t)
}

func ExampleNewWithURL() {
	t, err := olivetv.NewWithURL("https://www.huya.com/518512")
	if err != nil {
		println(err.Error())
		return
	}

	t.Snap()
	fmt.Println(t)
}

func TestExampleTv(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleTV()
}

func TestExampleSetCookie(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleSetCookie()
}

func TestExampleNewWithUrl(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleNewWithURL()
}
