package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DownloadFile 下载文件并显示进度
func DownloadFile(downloadURL, savePath string, headers map[string]string) error {
	// 检查文件是否已经存在
	if _, err := os.Stat(savePath); err == nil {
		return fmt.Errorf("文件 %s 已存在，跳过下载。", savePath) // 文件已存在，跳过下载
	} else if !os.IsNotExist(err) {
		// 如果发生了其他错误（而不是文件不存在的错误），则返回错误
		return fmt.Errorf("检查文件 %s 是否存在时出错: %v", savePath, err)
	}

	client := &http.Client{Timeout: 60 * time.Second}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//totalSize, _ := strconv.Atoi(resp.Header.Get("content-length"))
	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	downloaded := 0
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		downloaded += n
		_, err = out.Write(buffer[:n])
		if err != nil {
			return err
		}

		//fmt.Printf("\r%s: %.2f%%", filepath.Base(savePath), (float64(downloaded)/float64(totalSize))*100)
	}

	return nil
}

// PrepareDownloadFolder 创建保存文件夹
func PrepareDownloadFolder(folder string) string {
	saveFolder := "downloads"
	if folder != "" {
		saveFolder = filepath.Join("downloads", folder)
	}
	os.MkdirAll(saveFolder, os.ModePerm)
	return saveFolder
}
