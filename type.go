package gowss

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type subMessage struct {
	Attrs []string `json:"attrs"`
}

type MsgBody struct {
	To   string      `json:"to"`
	Body interface{} `json:"body"`
}

type response struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
}

func (m *MsgBody) BodyHash() string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%v", m.Body)))
	return hex.EncodeToString(h.Sum(nil))
}

func (m *MsgBody) GetBody() []byte {
	re := response{
		Type: m.To,
		Body: m.Body,
	}
	data, _ := json.Marshal(re)
	return data
}
