package olivetv_test

import (
	"testing"

	tv "github.com/go-olive/olive/foundation/olivetv"
)

func TestDouyin_Snap(t *testing.T) {
	u := "https://live.douyin.com/278246244716"
	cookie := "__ac_nonce=062f60f640020c76e5b7c; __ac_signature=_02B4Z6wo00f011BxI-gAAIDAxKd45w5q9dNQUSdAALb1WEJMf.7Ma1NuqG0oiO7cYko5mx60CrOZ7DEMeiImZWCkLuxGYUV7nL8NBVxgyWeCnQdWfhttJNV21omp7bThIi8SVu-58ihu3EiU32;"
	dy, err := tv.NewWithUrl(u, tv.SetCookie(cookie))
	if err != nil {
		println(err.Error())
		return
	}
	dy.Snap()
	t.Log(dy)
}
