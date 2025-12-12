package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sbzhu/weworkapi_golang/wxbizmsgcrypt"
)

type MsgContent struct {
	ToUsername   string `xml:"ToUserName"`
	FromUsername string `xml:"FromUserName"`
	CreateTime   uint32 `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Content      string `xml:"Content"`
	Msgid        string `xml:"MsgId"`
	Agentid      uint32 `xml:"AgentId"`
}

var wxcpt *wxbizmsgcrypt.WXBizMsgCrypt

func init() {
	token := "xxxxxxxxxx"
	receiverId := "wx5823bf96d3bd56c7"
	encodingAeskey := "xxxxxxxxxxxxxxxx"
	wxcpt = wxbizmsgcrypt.NewWXBizMsgCrypt(token, encodingAeskey, receiverId, wxbizmsgcrypt.XmlType)
}

// 处理 GET 请求 - 验证回调 URL
func verifyURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析 URL 参数
	query := r.URL.Query()
	msgSignature := query.Get("msg_signature")
	timestamp := query.Get("timestamp")
	nonce := query.Get("nonce")
	echostr := query.Get("echostr")

	// URL 解码
	decodedEchostr, err := url.QueryUnescape(echostr)
	if err != nil {
		log.Printf("URL decode error: %v", err)
		http.Error(w, "Invalid echostr parameter", http.StatusBadRequest)
		return
	}

	// 验证 URL
	echoStr, cryptErr := wxcpt.VerifyURL(msgSignature, timestamp, nonce, decodedEchostr)
	if cryptErr != nil {
		log.Printf("VerifyURL fail: %v", cryptErr)
		http.Error(w, fmt.Sprintf("VerifyURL fail: %s", cryptErr.ErrMsg), http.StatusBadRequest)
		return
	}

	log.Printf("VerifyURL success, echoStr: %s", string(echoStr))
	// 返回解密后的 echostr
	w.Write(echoStr)
}

// 处理 POST 请求 - 解密消息
func decryptMsgHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析 URL 参数
	query := r.URL.Query()
	msgSignature := query.Get("msg_signature")
	timestamp := query.Get("timestamp")
	nonce := query.Get("nonce")

	// 读取 POST 数据
	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Read body error: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 解密消息
	msg, cryptErr := wxcpt.DecryptMsg(msgSignature, timestamp, nonce, reqData)
	if cryptErr != nil {
		log.Printf("DecryptMsg fail: %v", cryptErr)
		http.Error(w, fmt.Sprintf("DecryptMsg fail: %s", cryptErr.ErrMsg), http.StatusBadRequest)
		return
	}

	log.Printf("DecryptMsg success, msg: %s", string(msg))

	// 解析消息内容
	var msgContent MsgContent
	err = xml.Unmarshal(msg, &msgContent)
	if err != nil {
		log.Printf("Unmarshal fail: %v", err)
		http.Error(w, "Failed to parse message", http.StatusInternalServerError)
		return
	}

	log.Printf("Parsed message: %+v", msgContent)

	// 构造回复消息（示例：自动回复）
	replyMsg := buildReplyMessage(&msgContent)

	// 加密回复消息
	timestampStr := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr := strconv.FormatInt(time.Now().UnixNano(), 10)
	encryptMsg, cryptErr := wxcpt.EncryptMsg(replyMsg, timestampStr, nonceStr)
	if cryptErr != nil {
		log.Printf("EncryptMsg fail: %v", cryptErr)
		http.Error(w, fmt.Sprintf("EncryptMsg fail: %s", cryptErr.ErrMsg), http.StatusInternalServerError)
		return
	}

	log.Printf("EncryptMsg success")
	w.Header().Set("Content-Type", "application/xml")
	w.Write(encryptMsg)
}

// 构造回复消息
func buildReplyMessage(msgContent *MsgContent) string {
	now := time.Now().Unix()
	replyContent := fmt.Sprintf("收到您的消息: %s", msgContent.Content)

	replyMsg := fmt.Sprintf(`<xml>
<ToUserName><![CDATA[%s]]></ToUserName>
<FromUserName><![CDATA[%s]]></FromUserName>
<CreateTime>%d</CreateTime>
<MsgType><![CDATA[text]]></MsgType>
<Content><![CDATA[%s]]></Content>
<AgentID>%d</AgentID>
</xml>`, msgContent.FromUsername, msgContent.ToUsername, now, replyContent, msgContent.Agentid)

	return replyMsg
}

func main() {
	http.HandleFunc("/cgi-bin/wxpush", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			verifyURLHandler(w, r)
		} else if r.Method == http.MethodPost {
			decryptMsgHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := ":8090"
	log.Printf("Starting HTTP server on port %s", port)
	log.Printf("GET /cgi-bin/wxpush - 验证回调 URL")
	log.Printf("POST /cgi-bin/wxpush - 接收并解密消息")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
