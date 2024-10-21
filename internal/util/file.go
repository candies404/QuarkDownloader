package util

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

	client, err := GetHTTPClient()
	if err != nil {
		return err
	}

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

	// 解析响应数据
	if resp.StatusCode != 200 {
		if resp.StatusCode == 403 {
			log.Fatal("token似乎失效了，请重新登录")
		}
		return fmt.Errorf("下载文件失败：%s", resp.Status)
	}

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
	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		fmt.Printf("创建文件连接失败：%v", err)
		return ""
	}
	return folder
}

// PathExists 判断所给路径文件/文件夹是否存在
func PathExists(path string) (bool, int64, error) {
	f, err := os.Stat(path)
	if err == nil {
		return true, f.Size(), err
	}
	//isnotexist来判断，是不是不存在的错误
	if os.IsNotExist(err) { //如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false, 0, nil
	}
	return false, 0, err //如果有错误了，但是不是不存在的错误，所以把这个错误原封不动的返回
}
