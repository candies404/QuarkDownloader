package quark

import (
	"QuarkDownloader/internal/util"
	"encoding/json"
	"fmt"
	"net/http"
)

// QuarkGetDetails 获取网盘的文件夹和文件详情
func (q *FileManager) QuarkGetDetails(folderID string) (map[string]interface{}, error) {
	detailsAPI := "https://drive-pc.quark.cn/1/clouddrive/file/details"

	resp, err := util.SendRequest(http.MethodGet, detailsAPI,
		map[string]string{
			"folder_id": folderID,
			"timestamp": "",
		}, nil, q.Headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, err
	}

	if status, ok := responseData["status"].(float64); ok && int(status) != 200 {
		return nil, fmt.Errorf("获取网盘详情失败: %s", responseData["message"])
	}

	return responseData, nil
}
