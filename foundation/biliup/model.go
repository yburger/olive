package biliup

type CookieData struct {
	IsNew        bool   `json:"is_new"`
	Mid          int64  `json:"mid"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenInfo    struct {
		Mid          int64  `json:"mid"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	} `json:"token_info"`
	CookieInfo struct {
		Cookies []struct {
			Name     string `json:"name"`
			Value    string `json:"value"`
			HTTPOnly int64  `json:"http_only"`
			Expires  int64  `json:"expires"`
			Secure   int64  `json:"secure"`
		} `json:"cookies"`
		Domains []string `json:"domains"`
	} `json:"cookie_info"`
	Sso []string `json:"sso"`
}

type PreuploadInfo struct {
	OK              int64       `json:"OK"`
	Auth            string      `json:"auth"`
	BizID           int64       `json:"biz_id"`
	ChunkRetry      int64       `json:"chunk_retry"`
	ChunkRetryDelay int64       `json:"chunk_retry_delay"`
	ChunkSize       int64       `json:"chunk_size"`
	Endpoint        string      `json:"endpoint"`
	Endpoints       []string    `json:"endpoints"`
	ExposeParams    interface{} `json:"expose_params"`
	PutQuery        string      `json:"put_query"`
	Threads         int64       `json:"threads"`
	Timeout         int64       `json:"timeout"`
	Uip             string      `json:"uip"`
	UposURI         string      `json:"upos_uri"`
}

type UploadInfo struct {
	Location string `json:"location"`
	Etag     string `json:"etag"`
	OK       int64  `json:"OK"`
	Bucket   string `json:"bucket"`
	Key      string `json:"key"`
	UploadID string `json:"upload_id"`
}

type ReqJSON struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	PartNumber int64  `json:"partNumber"`
	ETag       string `json:"eTag"`
}

type AddReqJSON struct {
	Copyright        int64    `json:"copyright"` // 1:原创 2:转载
	Cover            string   `json:"cover"`     // 封面url
	Title            string   `json:"title"`     // 视频标题
	Tid              int64    `json:"tid"`       // 分区id
	Tag              string   `json:"tag"`       // 标签 , 分割
	DescFormatID     int64    `json:"desc_format_id"`
	Desc             string   `json:"desc"`             // 简介
	Source           string   `json:"source,omitempty"` // 来源
	Dynamic          string   `json:"dynamic"`
	Interactive      int64    `json:"interactive"`
	Videos           []Video  `json:"videos"`
	ActReserveCreate int64    `json:"act_reserve_create"`
	NoDisturbance    int64    `json:"no_disturbance"`
	NoReprint        int64    `json:"no_reprint"`
	Subtitle         Subtitle `json:"subtitle"`
	Dolby            int64    `json:"dolby"`
	LosslessMusic    int64    `json:"lossless_music"`
	Csrf             string   `json:"csrf"`
}

type Video struct {
	Filename string `json:"filename"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	Cid      int64  `json:"cid"`
}

type Subtitle struct {
	Open int64  `json:"open"`
	Lan  string `json:"lan"`
}
