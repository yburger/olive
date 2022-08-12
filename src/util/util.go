package util

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
)

func GetURLContent(url string) (string, error) {
	webUserAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", webUserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	content := string(respBody)
	return content, nil
}

func Match(pattern, content string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	submatch := re.FindAllStringSubmatch(content, -1)
	res := make([]string, 0)
	for _, v := range submatch {
		res = append(res, string(v[1]))
	}
	if len(res) < 1 {
		return "", errors.New("pattern not found")
	}
	return res[0], nil
}

func MoveFile(src, dst string) error {
	if runtime.GOOS == "windows" {
		return MoveFileWindows(src, dst)
	}
	return os.Rename(src, dst)
}

func MoveFileWindows(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}

	err = out.Sync()
	if err != nil {
		return fmt.Errorf("Sync error: %s", err)
	}

	si, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("Stat error: %s", err)
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return fmt.Errorf("Chmod error: %s", err)
	}

	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}
