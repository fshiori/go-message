//package message is a wrapper for message.net.tw's API.
package message

import (
	"bytes"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"

	"github.com/parnurzeal/gorequest"
)

// ServerURL is API endpoint.
const ServerURL = "http://api.message.net.tw/"

// SendPath is send message API path.
const SendPath = "send.php"

// QueryLogPath is query sended message log API path.
const QueryLogPath = "query.php"

// ReserveDelPath is delete reserve message API path.
const ReserveDelPath = "del.php"

// Message is base type for all api call.
type Message struct {
	ID       string
	Password string
}

// SendResp is respone type for send message API call.
type SendResp struct {
	ErrorCode int
	LeftCount int
	MessageID map[string]string
}

// QueryLogResp is respone type for query sended message log API call.
type QueryLogResp struct {
	ErrorCode int
	LeftCount int
	Item      map[string]map[string]string
}

// ReserveDelResp is respone type for delete reserve message API call.
type ReserveDelResp struct {
	ErrorCode     int
	LeftCount     int
	MessageStatus map[string]int
}

// NewMessage creates a new MessageAPI instance
func NewMessage(id string, password string) *Message {
	return &Message{ID: id, Password: password}
}

// Send makes a request to send message.
func (msg Message) Send(longSMS int, sendDate string, tel []string, message string) (resp []SendResp, errs [][]error) {
	var buffer bytes.Buffer
	buffer.WriteString(ServerURL)
	buffer.WriteString(SendPath)

	telCount := len(tel)
	sendTimes := telCount / 100
	if telCount%100 != 0 {
		sendTimes++
	}

	for i := 0; i < sendTimes; i++ {
		tResp := SendResp{}
		v := url.Values{}
		if longSMS != 0 {
			v.Add("longsms", strconv.Itoa(longSMS))
		}
		v.Add("id", msg.ID)
		v.Add("password", msg.Password)
		if sendDate != "" {
			v.Add("sdate", sendDate)
		}
		var sendTel []string
		if i == sendTimes-1 {
			sendTel = tel[i*100:]
		} else {
			sendTel = tel[i*100 : (i+1)*100]
		}
		v.Add("tel", strings.Join(sendTel, ";"))
		v.Add("msg", message)
		v.Add("encoding", "urlencode_utf8")

		request := gorequest.New()
		_, body, tErrs := request.Post(buffer.String()).Send(v.Encode()).End()
		errs = append(errs, tErrs)
		if len(tErrs) != 0 {
			continue
		}
		result := strings.Split(body, "\n")
		tResp.MessageID = map[string]string{}
		for _, value := range result {
			data := strings.Split(strings.TrimSpace(value), "=")
			if len(data) != 2 {
				continue
			}
			switch {
			case data[0] == "ErrorCode":
				tResp.ErrorCode, _ = strconv.Atoi(data[1])
			case data[0] == "LCount":
				tResp.LeftCount, _ = strconv.Atoi(data[1])
			default:
				tResp.MessageID[data[0]] = data[1]
			}
		}
		resp = append(resp, tResp)
	}
	return
}

// QueryLog makes a request to query sended message log.
func (msg Message) QueryLog(longSMS int, queryColumns []string, startDate string, endDate string, startIndex int, msgCount int, mid []string) (resp QueryLogResp, errs []error) {
	var buffer bytes.Buffer
	buffer.WriteString(ServerURL)
	buffer.WriteString(QueryLogPath)

	v := url.Values{}
	if longSMS != 0 {
		v.Add("longsms", strconv.Itoa(longSMS))
	}
	v.Add("id", msg.ID)
	v.Add("password", msg.Password)
	if len(queryColumns) > 0 {
		v.Add("columns", strings.Join(queryColumns, ";"))
	}
	if startDate != "" {
		v.Add("startdate", startDate)
	}
	if endDate != "" {
		v.Add("enddate", endDate)
	}
	if startIndex != 0 {
		v.Add("sindex", strconv.Itoa(startIndex))
	}
	if msgCount != 1000 && msgCount != 0 {
		v.Add("mcount", strconv.Itoa(msgCount))
	}
	if len(mid) > 0 {
		v.Add("mid", strings.Join(mid, ";"))
	}

	request := gorequest.New()
	_, body, errs := request.Post(buffer.String()).Send(v.Encode()).End()
	if len(errs) != 0 {
		return
	}
	result := strings.Split(body, "\n")
	var caption []string
	var item []string
	resp.Item = map[string]map[string]string{}
	for _, value := range result {
		data := strings.Split(strings.TrimSpace(value), "=")
		switch {
		case data[0] == "ErrorCode":
			resp.ErrorCode, _ = strconv.Atoi(data[1])
		case data[0] == "LCount":
			resp.LeftCount, _ = strconv.Atoi(data[1])
		case data[0] == "caption":
			caption = strings.Split(data[1], ",")
		case strings.HasPrefix(data[0], "item"):
			item = strings.Split(data[1], ",")
			if len(caption) == len(item) {
				itemMap := map[string]string{}
				itemMap["id"] = strings.TrimLeft(data[0], "item")
				for i := 0; i < len(caption); i++ {
					mapKey := strings.Trim(caption[i], "\"")
					mapValue := strings.Trim(item[i], "\"")
					if mapKey == "prms" || mapKey == "msg" {
						mapValue, _, _ = transform.String(traditionalchinese.Big5.NewDecoder(), mapValue)
					}
					itemMap[mapKey] = mapValue
				}
				resp.Item[data[0]] = itemMap
			}
		}
	}
	return
}

// ReserveDel makes a request to delete reserve message.
func (msg Message) ReserveDel(msgID []string) (resp ReserveDelResp, errs []error) {
	var buffer bytes.Buffer
	buffer.WriteString(ServerURL)
	buffer.WriteString(ReserveDelPath)

	v := url.Values{}
	v.Add("id", msg.ID)
	v.Add("password", msg.Password)
	if len(msgID) > 0 {
		v.Add("msgid", strings.Join(msgID, ";"))
	} else {
		v.Add("msgid", "all")
	}

	request := gorequest.New()
	_, body, errs := request.Post(buffer.String()).Send(v.Encode()).End()
	if len(errs) != 0 {
		return
	}
	result := strings.Split(body, "\n")
	resp.MessageStatus = map[string]int{}
	for _, value := range result {
		data := strings.Split(strings.TrimSpace(value), "=")
		if len(data) != 2 {
			continue
		}
		switch {
		case data[0] == "ErrorCode":
			resp.ErrorCode, _ = strconv.Atoi(data[1])
		case data[0] == "LCount":
			resp.LeftCount, _ = strconv.Atoi(data[1])
		default:
			resp.MessageStatus[data[0]], _ = strconv.Atoi(data[1])
		}
	}
	return
}
