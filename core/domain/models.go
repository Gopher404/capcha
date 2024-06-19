package domain

type Capcha struct {
	Uid     string `json:"uid"`
	ImgSrc  string `json:"img_src"`
	Expires string `json:"ttl"`
}
