package main

var SuccessResponse = &StringMessage{Data: "sucess"}

type ErrorMessage struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type StringMessage struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

type ClueInfo struct {
	Title      string   `json:"title"`
	Desc       string   `json:"desc"`
	Count      int      `json:"count"`
	Number     string   `json:"number"`
	Attachment []string `json:"attachment"`
}

type ClueListResponse struct {
	Code int `json:"code"`
	Data []ClueInfo
}

type ClueInfoResponse struct {
	Code int       `json:"code"`
	Data *ClueInfo `json:"data"`
}

type BlobResponse struct {
	Code int               `json:"code"`
	Data map[string]string `json:"data"`
}
