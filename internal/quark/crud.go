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

// SaveRequestData 保存分享文件所需的请求体数据结构
type SaveRequestData struct {
	FidList      []string `json:"fid_list"`
	FidTokenList []string `json:"fid_token_list"`
	ToPdirFid    string   `json:"to_pdir_fid"`
	PwdID        string   `json:"pwd_id"`
	SToken       string   `json:"stoken"`
	PdirFid      string   `json:"pdir_fid"`
	Scene        string   `json:"scene"`
}

// DeleteFileRequestData 定义删除文件的请求体数据结构
type DeleteFileRequestData struct {
	ActionType  int      `json:"action_type"`  // 操作类型，2表示删除
	FileList    []string `json:"filelist"`     // 要删除的文件ID列表
	ExcludeFids []string `json:"exclude_fids"` // 排除的文件ID（可以为空）
}

// CreateFolderRequestData 定义创建文件夹的请求体数据结构
type CreateFolderRequestData struct {
	PdirFid     string `json:"pdir_fid"`      // 父目录ID
	FileName    string `json:"file_name"`     // 新建文件夹的名称
	DirPath     string `json:"dir_path"`      // 目录路径（可以为空）
	DirInitLock bool   `json:"dir_init_lock"` // 是否锁定文件夹
}

// CreateFolderResponseData 定义创建文件夹的响应体数据结构
type CreateFolderResponseData struct {
	Status    int    `json:"status"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Data      struct {
		Finish bool   `json:"finish"`
		Fid    string `json:"fid"`
	} `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
}

// TaskResponseData 保存分享文件的响应结构
type TaskResponseData struct {
	Status    int    `json:"status"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int    `json:"timestamp"`
	Data      struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
	Metadata struct {
		TqGap int `json:"tq_gap"`
	} `json:"metadata"`
}

func (q *FileManager) doTask(resp *http.Response) error {
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求Task失败，HTTP状态码: %d", resp.StatusCode)
	}
	// 解析响应体
	var responseData TaskResponseData
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应状态
	if responseData.Code != 0 {
		if responseData.Code == 41030 { // 账户被封
			log.Fatal("账户被封，无法转存，请重新获取cookie： https://pan.quark.cn/")
		} else {
			return fmt.Errorf("检查响应状态: %s", responseData.Message)
		}
	}
	err := q.QuarkGetTaskStatus(responseData.Data.TaskID)
	if err != nil {
		return fmt.Errorf("任务状态获取失败: %s", err.Error())
	}

	return nil
}

// QuarkSaveShareFiles 保存分享的文件到网盘
func (q *FileManager) QuarkSaveShareFiles(fidList, fidTokenList []string, pdirFid, toPdirFid string, shareNo int) error {
	// 请求URL
	saveAPI := "https://drive-pc.quark.cn/1/clouddrive/share/sharepage/save"
	share := q.Shares[shareNo]

	// 发送POST请求
	resp, err := util.SendRequest(http.MethodPost, saveAPI,
		map[string]string{
			"pr":           "ucpro",
			"fr":           "pc",
			"uc_param_str": "",
			"__dt":         strconv.Itoa(1000000 + rand.Intn(999999)), // 模拟的 __dt 参数
			"__t":          strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond))),
		},
		SaveRequestData{
			FidList:      fidList,
			FidTokenList: fidTokenList,
			ToPdirFid:    toPdirFid,
			PwdID:        share.PwdId,
			SToken:       share.SToken,
			PdirFid:      pdirFid,
			Scene:        "link",
		}, q.Headers)
	if err != nil {
		return fmt.Errorf("保存文件请求失败: %v", err)
	}
	defer resp.Body.Close()
	err = q.doTask(resp)
	if err != nil {
		return err
	}
	return nil
}

// QuarkDeleteFile 执行删除文件操作
func (q *FileManager) QuarkDeleteFile(fileIDs []string) error {
	// 请求URL
	deleteAPI := "https://drive-pc.quark.cn/1/clouddrive/file/delete"

	// 构建请求体数据
	requestData := DeleteFileRequestData{
		ActionType:  2,          // 2表示删除操作
		FileList:    fileIDs,    // 要删除的文件ID列表
		ExcludeFids: []string{}, // 没有排除的文件ID
	}

	// 发送POST请求
	resp, err := util.SendRequest(http.MethodPost, deleteAPI,
		map[string]string{
			"pr":           "ucpro",
			"fr":           "pc",
			"uc_param_str": "",
		}, requestData, q.Headers)
	if err != nil {
		return fmt.Errorf("删除文件请求失败: %v", err)
	}
	defer resp.Body.Close()
	err = q.doTask(resp)
	if err != nil {
		return err
	}
	return nil
}

// QuarkCreateFolder 执行创建文件夹操作
// QuarkCreateFolder 执行创建文件夹操作并获取FID
func (q *FileManager) QuarkCreateFolder(parentDirID, folderName string) (string, error) {
	// 请求URL
	createFolderAPI := "https://drive-pc.quark.cn/1/clouddrive/file"

	// 发送POST请求
	resp, err := util.SendRequest(http.MethodPost, createFolderAPI,
		map[string]string{
			"pr":           "ucpro",
			"fr":           "pc",
			"uc_param_str": "",
		},
		map[string]interface{}{
			"pdir_fid":      parentDirID, // 父目录ID，根目录ID为"0"
			"file_name":     folderName,  // 要创建的文件夹名称
			"dir_path":      "",          // 目录路径，可以为空
			"dir_init_lock": false,       // 文件夹不需要初始化锁定
		}, q.Headers)
	if err != nil {
		return "", fmt.Errorf("创建文件夹请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("创建文件夹失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 解析响应体
	var responseBody CreateFolderResponseData
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("解析创建文件夹响应失败: %v", err)
	}

	// 检查返回的状态码是否为0，表示操作成功
	if responseBody.Code != 0 {
		return "", fmt.Errorf("创建文件夹失败，错误信息: %s", responseBody.Message)
	}

	// 检查操作是否完成并返回FID
	if responseBody.Data.Finish {
		log.Printf("文件夹创建成功，FID: %s\n", responseBody.Data.Fid)
		return responseBody.Data.Fid, nil
	}

	return "", fmt.Errorf("文件夹创建未完成")
}

// Clear 清空 Quark.RootDir 下的文件/文件夹
func (q *FileManager) Clear() {
	err := q.QuarkDeleteFile([]string{q.Quark.SaveDir.PdirID})
	if err != nil {
		println("删除文件时出错: ", err.Error())
		return
	}
	q.Quark.SaveDir.PdirID = ""
	q.Quark.SaveDir.PdirID, err = q.QuarkCreateFolder("0", q.Quark.SaveDir.DirName)
	if err != nil {
		println("创建文件时出错: ", err.Error())
	}
}
