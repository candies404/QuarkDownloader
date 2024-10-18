package quark

import (
	"QuarkDownloader/internal/util"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// TaskResponseData 用于解析任务状态的响应数据
type TaskStatusResponseData struct {
	Status    int    `json:"status"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int    `json:"timestamp"`
	Data      struct {
		TaskID    string `json:"task_id"`
		TaskType  int    `json:"task_type"`
		TaskTitle string `json:"task_title"`
		Status    int    `json:"status"` // 2:成功
		Share     struct {
		} `json:"share"`
		SaveAs struct {
			ToPdirFid     string        `json:"to_pdir_fid"`
			IsPack        string        `json:"is_pack"`
			SaveAsTopFids []interface{} `json:"save_as_top_fids"`
		} `json:"save_as"`
	} `json:"data"`
	Metadata struct {
		TqGap int `json:"tq_gap"`
	} `json:"metadata"`
}

// QuarkGetTaskStatus 获取任务状态
func (q *FileManager) QuarkGetTaskStatus(taskID string) error {
	// 请求URL
	taskAPI := "https://drive-pc.quark.cn/1/clouddrive/task"

	// 最大重试次数，避免无限循环
	maxRetries := 10
	retryCount := 0

	for {
		// 构建查询参数
		params := map[string]string{
			"pr":           "ucpro",
			"fr":           "pc",
			"uc_param_str": "",
			"task_id":      taskID,
			"retry_index":  strconv.Itoa(1),
			"__dt":         strconv.Itoa(60000 + rand.Intn(999999)), // 随机生成 __dt 参数
			"__t":          strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond))),
		}
		// 发送GET请求
		resp, err := util.SendRequest(http.MethodGet, taskAPI, params, nil, q.Headers)
		if err != nil {
			return fmt.Errorf("获取任务状态失败: %v", err)
		}
		defer resp.Body.Close()

		// 解析响应
		var responseData TaskStatusResponseData
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return fmt.Errorf("解析任务状态响应失败: %v", err)
		}

		// 检查响应状态
		if responseData.Code != 0 {
			return fmt.Errorf("获取任务状态失败: %s", responseData.Message)
		}

		// 如果状态为1或2，任务完成，退出轮询
		if responseData.Data.Status == 3 {
			// 输出任务状态信息
			log.Printf("任务名称: %s, 当前状态: %d\n", responseData.Data.TaskTitle, responseData.Data.Status)
			log.Printf("任务完成，状态: 失败\n")
		} else if responseData.Data.Status == 0 || responseData.Data.Status == 1 {
			retryCount++
			if retryCount > maxRetries {
				return fmt.Errorf("任务状态检查超时，状态仍为:%d", responseData.Data.Status)
			}
		}

		return nil
	}
}
