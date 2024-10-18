package quark

import (
	"QuarkDownloader/internal/util"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type MemberResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TotalCapacity int64 `json:"total_capacity"`
		UseCapacity   int64 `json:"use_capacity"`
	} `json:"data"`
}

type InfoResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Nickname  string `json:"nickname"`
		AvatarURI string `json:"avatarUri"`
	} `json:"data"`
	Code string `json:"code"`
}

const (
	memberAPI = "https://drive-pc.quark.cn/1/clouddrive/member"
	infoAPI   = "https://pan.quark.cn/account/info"
)

// QuarkGetPanInfo 获取网盘的文件夹和文件详情及用户信息
func (q *FileManager) QuarkGetPanInfo() error {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// 并发任务 1: 获取网盘容量
	wg.Add(1)
	go func() {
		defer wg.Done()

		resp, err := util.SendRequest(http.MethodGet, memberAPI,
			map[string]string{
				"pr":              "ucpro",
				"fr":              "pc",
				"uc_param_str":    "",
				"fetch_subscribe": "true",
				"_ch":             "home",
				"fetch_identity":  "true",
			}, nil, q.Headers)
		if err != nil {
			errChan <- fmt.Errorf("请求网盘容量失败: %v", err)
			return
		}
		defer resp.Body.Close()

		var responseData MemberResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			errChan <- fmt.Errorf("解析网盘容量响应失败: %v", err)
			return
		}

		if responseData.Code != 0 {
			errChan <- fmt.Errorf("获取网盘容量失败: %s", responseData.Message)
			return
		}

		q.Quark.TotalCapacity = responseData.Data.TotalCapacity
		q.Quark.FreeCapacity = responseData.Data.TotalCapacity - responseData.Data.UseCapacity
	}()

	// 并发任务 2: 获取用户昵称
	wg.Add(1)
	go func() {
		defer wg.Done()

		resp, err := util.SendRequest(http.MethodGet, infoAPI,
			map[string]string{
				"fr":       "pc",
				"platform": "pc",
			}, nil, q.Headers)
		if err != nil {
			errChan <- fmt.Errorf("请求用户信息失败: %v", err)
			return
		}
		defer resp.Body.Close()

		var responseData InfoResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			errChan <- fmt.Errorf("解析用户信息响应失败: %v", err)
			return
		}

		if responseData.Code != "OK" {
			errChan <- fmt.Errorf("获取用户信息失败: %s", responseData.Code)
			return
		}

		q.Quark.NickName = responseData.Data.Nickname
	}()

	// 等待所有goroutine完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误发生
	allError := true
	for err := range errChan {
		if err != nil {
			log.Fatalf(err.Error())
		} else {
			allError = false
		}
	}
	if len(errChan) != 0 && allError {
		return fmt.Errorf("获取网盘详情接口出错")
	}
	return nil
}
