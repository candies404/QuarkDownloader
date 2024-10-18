package quark

import (
	"QuarkDownloader/internal/util"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TokenRequestData 是发送请求时的 JSON 数据结构
type TokenRequestData struct {
	PwdID    string `json:"pwd_id"`
	Passcode string `json:"passcode"`
}

// TokenResponseData 是返回的 JSON 数据结构
type TokenResponseData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Stoken string `json:"stoken"`
		Author struct {
			MemberType string `json:"member_type"`
			AvatarURL  string `json:"avatar_url"`
			NickName   string `json:"nick_name"`
		} `json:"author"`
		ExpiredType int    `json:"expired_type"`
		ExpiredAt   int64  `json:"expired_at"`
		Title       string `json:"title"`
	} `json:"data"`
}

// QuarkGetShares 获取分享页面的 Token
func (q *FileManager) QuarkGetShares() {
	// 请求URL
	tokenAPI := "https://drive-h.quark.cn/1/clouddrive/share/sharepage/token"

	log.Println("开始获取分享连接信息")

	for _, share := range q.Shares {
		// 从URL中提取 pwdId
		share.PwdId = strings.Replace(share.Url, "https://pan.quark.cn/s/", "", -1)

		// 发送POST请求
		resp, err := util.SendRequest(
			http.MethodPost,
			tokenAPI,
			map[string]string{
				"pr":           "ucpro",
				"fr":           "pc",
				"uc_param_str": "",
				"__dt":         strconv.Itoa(100 + rand.Intn(900)),
				"__t":          strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond))),
			},
			TokenRequestData{
				PwdID:    share.PwdId,
				Passcode: share.Passcode,
			},
			q.Headers)
		if err != nil {
			// 错误日志输出并继续下一个分享
			log.Printf("发送请求失败 (Share: %s, Error: %v)\n", share.PwdId, err)
			continue
		}
		defer resp.Body.Close()

		// 解析响应
		var responseData TokenResponseData
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			// 解析响应出错，记录日志并继续下一个分享
			log.Printf("解析响应失败 (Share: %s, Error: %v)\n", share.PwdId, err)
			continue
		}

		// 检查响应状态
		if responseData.Code != 0 {
			// 返回状态不为 0，记录错误信息并继续下一个分享
			log.Printf("获取Token失败 (Share: %s, Message: %s)\n", share.PwdId, responseData.Message)
			continue
		}

		// 成功获取 Token，设置到 Share 结构中
		share.SToken = responseData.Data.Stoken
		share.Title = responseData.Data.Title
	}
	for _, share := range q.Shares {
		log.Printf("%-40s ---- %s\n", share.Title, share.SToken)
	}
}
