package quark

import (
	"QuarkDownloader/internal/util"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type FileListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List []struct {
			Fid      string `json:"fid"`
			FileName string `json:"file_name"`
			PdirFid  string `json:"pdir_fid"`
			Size     int    `json:"size"`
			Dir      bool   `json:"dir"` // 是否是文件夹
		} `json:"list"`
	} `json:"data"`
}

// QuarkGetFileList 获取网盘文件列表，并根据 downLoad 参数判断是创建文件夹还是下载文件
func (q *FileManager) QuarkGetFileList(isPrint bool, pdirFid string, parentNode *DirectoryNode, indentLevel int) error {
	fileListAPI := "https://drive-pc.quark.cn/1/clouddrive/file/sort"
	page := 1
	pageSize := 50

	// 树形结构的缩进
	indent := strings.Repeat("  ", indentLevel)

	// 分页处理，持续请求直到没有更多文件
	for {
		// 发送GET请求
		resp, err := util.SendRequest(http.MethodGet, fileListAPI,
			map[string]string{
				"pr":              "ucpro",
				"fr":              "pc",
				"uc_param_str":    "",
				"pdir_fid":        pdirFid,                         // 当前文件夹ID
				"_page":           strconv.Itoa(page),              // 当前页
				"_size":           strconv.Itoa(pageSize),          // 每页的文件数
				"_fetch_total":    "1",                             // 是否获取总数
				"_fetch_sub_dirs": "0",                             // 是否获取子目录
				"_sort":           "file_type:asc,updated_at:desc", // 排序方式
			}, nil, q.Headers)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		// 解析响应
		var responseData FileListResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return err
		}

		// 检查响应状态
		if responseData.Code != 0 {
			return fmt.Errorf("获取网盘文件列表失败: %s", responseData.Message)
		}

		// 如果没有文件了，退出分页循环
		if len(responseData.Data.List) == 0 {
			break
		}

		// 遍历文件/文件夹列表
		for _, file := range responseData.Data.List {
			// 创建当前文件/文件夹的 DirectoryNode
			currentNode := &DirectoryNode{
				Name:     file.FileName,
				PwdId:    file.Fid,
				IsDir:    file.Dir,
				SizeMB:   0,   // 默认文件夹大小为 0
				Children: nil, // 先设为空
			}
			if file.Dir {
				// 文件夹处理逻辑
				if q.Quark.SaveDir.DirName == file.FileName {
					q.Quark.SaveDir.PdirID = file.Fid
				}
				if isPrint {
					log.Printf("%s📁 %s\n", indent, file.FileName)
					// 初始化子节点切片
					currentNode.Children = []*DirectoryNode{}

					// 递归获取子文件夹内容，增加缩进，同时传递当前路径
					err := q.QuarkGetFileList(isPrint, file.Fid, currentNode, indentLevel+1)
					if err != nil {
						log.Printf("获取子文件夹失败: %s\n", err)
					}
				}
			} else {
				// 文件处理逻辑，将字节转换为MB并设置大小
				if isPrint {
					sizeInMB := float64(file.Size) / (1024 * 1024)
					currentNode.SizeMB = sizeInMB
					log.Printf("%s📄 %s - %.2f MB\n", indent, file.FileName, sizeInMB)
				}
			}

			// 将当前节点加入到父节点的子节点列表中
			parentNode.Children = append(parentNode.Children, currentNode)
		}

		// 翻页
		page++
	}

	return nil
}

// QuarkDownloadAndClear 实现多线程下载多个文件并清理
func (q *FileManager) QuarkDownloadAndClear(pdirFid string, currentPath string) error {
	fileListAPI := "https://drive-pc.quark.cn/1/clouddrive/file/sort"
	page := 1
	pageSize := 50

	// 使用 WaitGroup 等待所有下载任务完成
	var wg sync.WaitGroup
	// 用于限制并发下载数量
	sem := make(chan struct{}, 20)

	// 分页处理，持续请求直到没有更多文件
	for {
		// 发送GET请求
		resp, err := util.SendRequest(http.MethodGet, fileListAPI,
			map[string]string{
				"pr":              "ucpro",
				"fr":              "pc",
				"uc_param_str":    "",
				"pdir_fid":        pdirFid,                         // 当前文件夹ID
				"_page":           strconv.Itoa(page),              // 当前页
				"_size":           strconv.Itoa(pageSize),          // 每页的文件数
				"_fetch_total":    "1",                             // 是否获取总数
				"_fetch_sub_dirs": "0",                             // 是否获取子目录
				"_sort":           "file_type:asc,updated_at:desc", // 排序方式
			}, nil, q.Headers)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		// 解析响应
		var responseData FileListResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return err
		}

		// 检查响应状态
		if responseData.Code != 0 {
			return fmt.Errorf("获取网盘文件列表失败: %s", responseData.Message)
		}

		// 如果没有文件了，退出分页循环
		if len(responseData.Data.List) == 0 {
			break
		}

		crtPath := currentPath
		// 遍历文件/文件夹列表
		for _, file := range responseData.Data.List {
			// 创建当前文件/文件夹的 DirectoryNode
			currentNode := &DirectoryNode{
				Name:     file.FileName,
				PwdId:    file.Fid,
				IsDir:    file.Dir,
				SizeMB:   0,   // 默认文件夹大小为 0
				Children: nil, // 先设为空
			}
			if file.Dir {
				// 初始化子节点切片
				currentNode.Children = []*DirectoryNode{}

				crtPath := filepath.Join(currentPath, file.FileName)
				err := os.MkdirAll(crtPath, os.ModePerm)
				if err != nil {
					log.Printf("创建文件夹失败: %s\n", err)
				}

				err = q.QuarkDownloadAndClear(file.Fid, crtPath)
				if err != nil {
					log.Printf("获取子文件夹失败: %s\n", err)
				}
			} else {
				// 文件下载使用多线程，增加 WaitGroup 和 Semaphore 控制并发
				wg.Add(1)
				sem <- struct{}{} // 占用一个并发槽位
				go func(fileFid, path string) {
					defer wg.Done()
					defer func() { <-sem }() // 释放并发槽位

					// 下载文件
					q.FileDownload([]string{fileFid}, path)
				}(file.Fid, crtPath)
			}
		}

		// 翻页
		page++
	}

	// 等待所有文件下载完成
	wg.Wait()

	// 执行清理
	q.Clear()

	return nil
}
