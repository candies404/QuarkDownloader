package quark

import (
	"QuarkDownloader/internal/util"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// FileDownload 获取下载地址并下载文件
func (q *FileManager) FileDownload(fids []string, folder string) {
	downloadAPI := "https://drive-pc.quark.cn/1/clouddrive/file/download"

	resp, err := util.SendRequest(
		http.MethodPost,
		downloadAPI,
		map[string]string{
			"pr":           "ucpro",
			"fr":           "pc",
			"uc_param_str": "",
			"__dt":         strconv.Itoa(600 + rand.Intn(9399)),
			"__t":          strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond))),
		},
		map[string]interface{}{
			"fids": fids,
		},
		q.Headers)
	if err != nil {
		log.Printf("下载 \033[31m失败\033[0m : 获取下载连接\033[31m失败\033[0m\n")
		return
	}
	defer resp.Body.Close()

	// 解析响应数据
	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Printf("下载 \033[31m失败\033[0m : 解析响应\033[31m失败\033[0m\n")
		return
	}

	if status, ok := responseData["status"].(float64); ok && int(status) != 200 {
		log.Printf("下载 \033[31m失败\033[0m : 获取下载URL列表失败: %s\n", responseData["message"])
		return
	}

	dataList, ok := responseData["data"].([]interface{})
	if !ok {
		log.Printf("下载 \033[31m失败\033[0m :无效的文件列表\n")
		return
	}

	saveFolder := util.PrepareDownloadFolder(folder)

	var wg sync.WaitGroup
	for i, data := range dataList {
		wg.Add(1)

		go func(i int, fileData interface{}) {
			defer wg.Done()

			fileInfo := fileData.(map[string]interface{})
			fileName, _ := fileInfo["file_name"].(string)
			downloadURL, _ := fileInfo["download_url"].(string)

			savePath := filepath.Join(saveFolder, fileName)
			if err := util.DownloadFile(downloadURL, savePath, q.Headers); err != nil {
				log.Printf("下载\u001B[31m失败\u001B[0m %-40s: %v\n", fileName, err)
			} else {
				log.Printf("下载\u001B[32m成功\u001B[0m %-40s\n", fileName)
			}
		}(i, data)
	}
	wg.Wait()
}
