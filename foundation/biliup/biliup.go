// Package biliup provides support for uploading videos to bilibili.
package biliup

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/imroc/req/v3"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

type Biliup struct {
	Config Config

	client *req.Client

	cookieInfo     CookieInfo
	videoMetadata  VideoMetadata
	uploadMetadata UploadMetadata
}

type Config struct {
	CookieFilepath string
	VideoFilepath  string
	Threads        int64
}

func New(cfg Config) *Biliup {
	return &Biliup{
		Config: cfg,
	}
}

func (b *Biliup) Upload() error {
	err := b.newClient(b.Config.CookieFilepath)
	if err != nil {
		return err
	}
	err = b.newVideoMetadata(b.Config.VideoFilepath)
	if err != nil {
		return err
	}
	err = b.preUpload()
	if err != nil {
		return err
	}
	err = b.periUpload()
	if err != nil {
		return err
	}
	err = b.postUpload()
	if err != nil {
		return err
	}
	return nil
}

type CookieInfo struct {
	Cookie string
	Csrf   string
}

type VideoMetadata struct {
	Filepath string // 视频路径
	Filename string // 视频名称
	Filesize int64  // 视频大小
	Title    string // 视频标题
	Desc     string // 视频简介
	UpType   int64  // 1:原创 2:转载
	Cover    string // 封面路径
	Tid      int64  // 分区id
	Tag      string // 标签 `,`分割
	Source   string // 来源
}

type UploadMetadata struct {
	Auth      string
	BaseURL   string
	FileName  string
	ChunkSize int64
	BizID     int64

	UploadID string
}

func (b *Biliup) newClient(cookieFilepath string) error {
	cookieBytes, err := os.ReadFile(cookieFilepath)
	if err != nil {
		return err
	}
	var cd CookieData
	err = jsoniter.Unmarshal(cookieBytes, &cd)
	if err != nil {
		return err
	}

	var ci CookieInfo
	for _, v := range cd.CookieInfo.Cookies {
		ci.Cookie += v.Name + "=" + v.Value + ";"
		if v.Name == "bili_jct" {
			ci.Csrf = v.Value
		}
	}

	client := req.C().SetCommonHeaders(map[string]string{
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36 Edg/105.0.1343.53",
		"cookie":     ci.Cookie,
		"Connection": "keep-alive",
	})
	resp, err := client.R().Get("https://api.bilibili.com/x/web-interface/nav")
	if err != nil {
		return err
	}
	uname := gjson.ParseBytes(resp.Bytes()).Get("data.uname").String()
	if uname == "" {
		return errors.New("cookie失效,请重新登录")
	}

	b.client = client
	b.cookieInfo = ci

	return nil
}

func (b *Biliup) newVideoMetadata(videoFilepath string) error {
	file, err := os.Open(videoFilepath)
	if err != nil {
		return err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	filename := filepath.Base(videoFilepath)
	title := strings.TrimSuffix(filename, filepath.Ext(videoFilepath))

	b.videoMetadata = VideoMetadata{
		Filepath: videoFilepath,
		Filesize: fileInfo.Size(),
		Filename: filename,
		Title:    title,
		Desc:     "",
		UpType:   2,
		Cover:    "",
		Tid:      21, // 日常
		Tag:      "日常,生活,vlog",
		Source:   "www",
	}

	return nil
}

func (b *Biliup) preUpload() error {
	var info PreuploadInfo
	b.client.R().SetQueryParams(map[string]string{
		"probe_version": "20211012",
		"upcdn":         "bda2",
		"zone":          "cs",
		"name":          b.videoMetadata.Filename,
		"r":             "upos",
		"profile":       "ugcfx/bup",
		"ssl":           "0",
		"version":       "2.10.4.0",
		"build":         "2100400",
		"size":          strconv.FormatInt(b.videoMetadata.Filesize, 10),
		"webVersion":    "2.0.0",
	}).SetResult(&info).Get("https://member.bilibili.com/preupload")

	UposURISlice := strings.Split(info.UposURI, "//")
	if len(UposURISlice) < 2 {
		return fmt.Errorf("invalid UposUri[%s]", info.UposURI)
	}

	UposURISlice2 := strings.Split(UposURISlice[1], "/")
	if len(UposURISlice2) < 2 {
		return fmt.Errorf("invalid UposUri[%s]", info.UposURI)
	}

	b.uploadMetadata = UploadMetadata{
		BaseURL:   fmt.Sprintf("https:%s/%s", info.Endpoint, UposURISlice[1]),
		FileName:  strings.Split(UposURISlice2[1], ".")[0],
		ChunkSize: info.ChunkSize,
		Auth:      info.Auth,
		BizID:     info.BizID,
	}

	return nil
}

func (b *Biliup) periUpload() (err error) {
	var upinfo UploadInfo
	b.client.SetCommonHeader(
		"X-Upos-Auth", b.uploadMetadata.Auth).R().
		SetQueryParams(map[string]string{
			"uploads":       "",
			"output":        "json",
			"profile":       "ugcfx/bup",
			"filesize":      strconv.FormatInt(b.videoMetadata.Filesize, 10),
			"partsize":      strconv.FormatInt(b.uploadMetadata.ChunkSize, 10),
			"biz_id":        strconv.FormatInt(b.uploadMetadata.BizID, 10),
			"meta_upos_uri": b.getMetaUposURI(),
		}).SetResult(&upinfo).Post(b.uploadMetadata.BaseURL)

	b.uploadMetadata.UploadID = upinfo.UploadID

	chunks := int64(math.Ceil(float64(b.videoMetadata.Filesize) / float64(b.uploadMetadata.ChunkSize)))
	// log.Printf("total chunks %d", chunks)

	parts := make([]Part, chunks)

	file, err := os.Open(b.videoMetadata.Filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	chunk := 0
	start := 0

	concurrentGoroutines := make(chan struct{}, b.Config.Threads)

	var wg sync.WaitGroup
	for {
		buf := GetBytes(int(b.uploadMetadata.ChunkSize))
		size, err := file.Read(buf)
		if err != nil && err != io.EOF {
			break
		}
		if size > 0 {
			wg.Add(1)
			go func(chunk int, start, size int, buf []byte) {
				defer wg.Done()
				concurrentGoroutines <- struct{}{}
				// log.Println("doing chunk", chunk)

				_, err := b.client.R().SetHeaders(map[string]string{
					"Content-Type":   "application/octet-stream",
					"Content-Length": strconv.Itoa(size),
				}).SetQueryParams(map[string]string{
					"partNumber": strconv.Itoa(chunk + 1),
					"uploadId":   b.uploadMetadata.UploadID,
					"chunk":      strconv.Itoa(chunk),
					"chunks":     strconv.Itoa(int(chunks)),
					"size":       strconv.Itoa(size),
					"start":      strconv.Itoa(start),
					"end":        strconv.Itoa(start + size),
					"total":      strconv.FormatInt(b.videoMetadata.Filesize, 10),
				}).SetBodyBytes(buf[:size]).SetRetryCount(5).AddRetryHook(func(resp *req.Response, err error) {
					// log.Println("重试发送分片", chunk)
				}).
					AddRetryCondition(func(resp *req.Response, err error) bool {
						return err != nil || resp.StatusCode != 200
					}).Put(b.uploadMetadata.BaseURL)

				PutBytes(buf)

				if err != nil {
					err = fmt.Errorf("视频文件[%s]分片[%d]上传失败[err msg = %s]分片大小[%d]", b.videoMetadata.Filename, chunk, err.Error(), size)
				}
				parts[chunk] = Part{
					PartNumber: int64(chunk + 1),
					ETag:       "etag",
				}

				<-concurrentGoroutines
			}(chunk, start, size, buf)

			start += size
			chunk++
		}

		if err == io.EOF {
			break
		}
	}
	wg.Wait()

	reqJSON := ReqJSON{
		Parts: parts,
	}
	reqStr, _ := jsoniter.MarshalToString(reqJSON)
	b.client.R().SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"Origin":       "https://member.bilibili.com",
		"Referer":      "https://member.bilibili.com/",
	}).SetQueryParams(map[string]string{
		"output":   "json",
		"profile":  "ugcfx/bup",
		"name":     b.videoMetadata.Filename,
		"uploadId": b.uploadMetadata.UploadID,
		"biz_id":   strconv.FormatInt(b.uploadMetadata.BizID, 10),
	}).SetBodyString(reqStr).SetResult(&upinfo).SetRetryCount(5).AddRetryHook(func(resp *req.Response, err error) {
		// log.Println("重试发送分片确认请求")
	}).
		AddRetryCondition(func(resp *req.Response, err error) bool {
			return err != nil || resp.StatusCode != 200
		}).Post(b.uploadMetadata.BaseURL)

	return nil
}

func (b *Biliup) getMetaUposURI() string {
	var info PreuploadInfo
	b.client.R().SetQueryParams(map[string]string{
		"name":       "file_meta.txt",
		"size":       "2000",
		"r":          "upos",
		"profile":    "fxmeta/bup",
		"ssl":        "0",
		"version":    "2.10.4",
		"build":      "2100400",
		"webVersion": "2.0.0",
	}).SetResult(&info).Get("https://member.bilibili.com/preupload")

	return info.UposURI
}

func (b *Biliup) postUpload() error {
	var addreq = AddReqJSON{
		Copyright:    b.videoMetadata.UpType,
		Cover:        "",
		Title:        b.videoMetadata.Title,
		Tid:          b.videoMetadata.Tid,
		Tag:          b.videoMetadata.Tag,
		DescFormatID: 16,
		Desc:         b.videoMetadata.Desc,
		Source:       b.videoMetadata.Source,
		Dynamic:      "",
		Interactive:  0,
		Videos: []Video{
			{
				Filename: b.uploadMetadata.FileName,
				Title:    b.videoMetadata.Filename,
				Desc:     "",
				Cid:      b.uploadMetadata.BizID,
			},
		},
		ActReserveCreate: 0,
		NoDisturbance:    0,
		NoReprint:        1,
		Subtitle: Subtitle{
			Open: 0,
			Lan:  "",
		},
		Dolby:         0,
		LosslessMusic: 0,
		Csrf:          b.cookieInfo.Csrf,
	}

	// {"code":0,"message":"0","ttl":1,"data":{"aid":000,"bvid":"Bxx"}}
	resp, err := b.client.R().SetQueryParams(map[string]string{
		"csrf": b.cookieInfo.Csrf,
	}).SetBodyJsonMarshal(addreq).Post("https://member.bilibili.com/x/vu/web/add/v3")

	if err != nil {
		return err
	}

	if gjson.ParseBytes([]byte(resp.Bytes())).Get("code").String() != "0" {
		return fmt.Errorf("视频文件[%s]投稿失败[%s]", b.videoMetadata.Filename, resp.String())
	}

	return nil
}
