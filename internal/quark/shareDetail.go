package quark

import (
	"QuarkDownloader/config"
	"QuarkDownloader/internal/util"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SharePageDetailResponse 定义返回的 JSON 数据结构
type SharePageDetailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List []FileInfo `json:"list"`
	} `json:"data"`
}

type FileInfo struct {
	Fid           string `json:"fid"`
	ShareFidToken string `json:"share_fid_token"`
	FileName      string `json:"file_name"`
	Size          int    `json:"size"`
	Dir           bool   `json:"dir"`
	UpdatedAt     int64  `json:"updated_at"`
}

var includeReg, excludeReg *regexp.Regexp

func init() {
	if config.Cfg.DownloadFilter.Include != "" {
		includeReg = regexp.MustCompile(config.Cfg.DownloadFilter.Include)
	}
	if config.Cfg.DownloadFilter.Exclude != "" {
		excludeReg = regexp.MustCompile(config.Cfg.DownloadFilter.Exclude)
	}
}

// QuarkGetSharePageDetail 获取分享页面的文件夹详情
func (q *FileManager) QuarkGetSharePageDetail(pdirFid string, shareNo, indentLevel int) error {
	// 请求URL
	shareDetailAPI := "https://drive-h.quark.cn/1/clouddrive/share/sharepage/detail"
	page := 1
	pageSize := 50
	share := q.Shares[shareNo]
	// 树形结构的缩进
	indent := strings.Repeat("  ", indentLevel)

	// 分页处理，持续请求直到没有更多文件
	for {
		// 发送GET请求
		resp, err := util.SendRequest(http.MethodGet, shareDetailAPI,
			map[string]string{
				"pr":            "ucpro",
				"fr":            "pc",
				"uc_param_str":  "",
				"pwd_id":        share.PwdId,
				"stoken":        share.SToken,
				"pdir_fid":      pdirFid, // 当前访问的文件夹ID，0 表示根目录
				"force":         "0",
				"_page":         fmt.Sprintf("%d", page),     // 当前页
				"_size":         fmt.Sprintf("%d", pageSize), // 每页的文件数
				"_fetch_banner": "1",
				"_fetch_share":  "1",
				"_fetch_total":  "1",
				"_sort":         "file_type:asc,updated_at:desc", // 排序方式
				"__dt":          strconv.Itoa(600 + rand.Intn(9399)),
				"__t":           strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond))), // 需要确认这个时间戳字段的生成方式，或者使用固定值
			}, nil, q.Headers)

		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// 解析响应
		var responseData SharePageDetailResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return err
		}

		// 检查响应状态
		if responseData.Code != 0 {
			return fmt.Errorf("获取分享页面详情失败: %s", responseData.Message)
		}

		// 如果没有文件了，退出分页循环
		if len(responseData.Data.List) == 0 {
			break
		}

		// 输出文件/文件夹列表
		fidList, fidTokenList := make([]string, 0), make([]string, 0)
		for _, file := range responseData.Data.List {
			fidList = append(fidList, file.Fid)
			fidTokenList = append(fidTokenList, file.ShareFidToken)
			// 格式化文件/文件夹信息
			if file.Dir {
				log.Printf("%s📁 %s\n", indent, file.FileName) // 文件夹
				// 递归调用 QuarkGetSharePageDetail 来获取子文件夹内容，增加缩进
				if err := q.QuarkGetSharePageDetail(file.Fid, shareNo, indentLevel+1); err != nil {
					log.Printf("获取子文件夹失败: %s\n", err)
				}
			} else {
				// 将字节转换为MB
				sizeInMB := float64(file.Size) / (1024 * 1024)
				log.Printf("%s📄 %s - %.2f MB\n", indent, file.FileName, sizeInMB) // 文件
			}
		}
		err = q.QuarkSaveShareFiles(fidList, fidTokenList, pdirFid, "0", shareNo)
		if err != nil {
			return err
		}
		// 翻页
		page++
	}
	return nil
}

func (q *FileManager) QuarkGetShareAndDownload(pdirFid, crtPath string, shareNo int) error {
	// 请求URL
	shareDetailAPI := "https://drive-h.quark.cn/1/clouddrive/share/sharepage/detail"
	page := 1
	pageSize := 50
	share := q.Shares[shareNo]
	p := crtPath
	var crtSize int
	fidList, fidTokenList := make([]string, 0), make([]string, 0)

	// 分页处理，持续请求直到没有更多文件
	for {
		// 发送GET请求
		resp, err := util.SendRequest(http.MethodGet, shareDetailAPI,
			map[string]string{
				"pr":            "ucpro",
				"fr":            "pc",
				"uc_param_str":  "",
				"pwd_id":        share.PwdId,
				"stoken":        share.SToken,
				"pdir_fid":      pdirFid, // 当前访问的文件夹ID，0 表示根目录
				"force":         "0",
				"_page":         fmt.Sprintf("%d", page),     // 当前页
				"_size":         fmt.Sprintf("%d", pageSize), // 每页的文件数
				"_fetch_banner": "1",
				"_fetch_share":  "1",
				"_fetch_total":  "1",
				"_sort":         "file_type:asc,updated_at:desc", // 排序方式
				"__dt":          strconv.Itoa(600 + rand.Intn(9399)),
				"__t":           strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond))), // 需要确认这个时间戳字段的生成方式，或者使用固定值
			}, nil, q.Headers)

		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// 解析响应
		var responseData SharePageDetailResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return err
		}

		// 检查响应状态
		if responseData.Code != 0 {
			return fmt.Errorf("获取分享页面详情失败: %s", responseData.Message)
		}

		// 如果没有文件了，退出分页循环
		if len(responseData.Data.List) == 0 {
			break
		}
		for _, file := range responseData.Data.List {
			savePath := path.Join(config.Cfg.LocalSaveDir, crtPath, file.FileName)

			// 如果是文件夹，递归获取子文件夹
			if file.Dir {
				p := path.Join(crtPath, file.FileName)
				if err := q.QuarkGetShareAndDownload(file.Fid, p, shareNo); err != nil {
					log.Printf("获取子文件夹失败: %s\n", err)
				}
				continue
			}

			// 文件大小超过限制，跳过
			if file.Size > 100*1024*1024 {
				fmt.Printf("文件大小超出限制，无法下载：%s\n", file.FileName)
				continue
			}

			// 检查文件是否需要下载
			if shouldDownloadFile(file, savePath, includeReg, excludeReg) {
				crtSize += file.Size
				fidList = append(fidList, file.Fid)
				fidTokenList = append(fidTokenList, file.ShareFidToken)

				// 检查当前容量并处理下载
				if float32(crtSize)/float32(q.Quark.FreeCapacity) > 0.9 {
					err = saveAndDownloadFiles(q, fidList, fidTokenList, shareNo, pdirFid, crtPath)
					if err != nil {
						log.Printf("保存或下载文件失败：%v\n", err)
						continue
					}
					// 清空列表和计数器
					crtSize = 0
					fidList = make([]string, 0)
					fidTokenList = make([]string, 0)
				}
			}
		}
		// 翻页
		page++
	}
	if len(fidList) != 0 {
		err := q.QuarkSaveShareFiles(fidList, fidTokenList, pdirFid, q.Quark.SaveDir.PdirID, shareNo)
		if err != nil {
			return err
		}

		err = q.QuarkDownloadAndClear(q.Quark.SaveDir.PdirID, p)
		if err != nil {
			return err
		}
	}
	return nil
}

// 判断文件是否需要下载
func shouldDownloadFile(file FileInfo, savePath string, includeReg, excludeReg *regexp.Regexp) bool {
	// 过滤文件名，匹配 include 和 exclude 正则
	if (includeReg != nil && !includeReg.MatchString(file.FileName)) ||
		(excludeReg != nil && excludeReg.MatchString(file.FileName)) {
		return false
	}

	// 检查文件是否已存在，且文件大小是否一致
	if ok, size, _ := util.PathExists(savePath); ok {
		if int(size) == file.Size {
			fmt.Printf("目标文件已存在且大小相同，跳过下载：%s\n", savePath)
			return false
		}
		fmt.Printf("文件大小不一样：删除原有文件重新下载 %s\n", savePath)
		if err := os.Remove(savePath); err != nil {
			fmt.Printf("删除文件失败：%s，错误：%v\n", savePath, err)
			return false
		}
	}
	return true
}

// 保存和下载文件的逻辑
func saveAndDownloadFiles(q *FileManager, fidList, fidTokenList []string, shareNo int, pdirFid, crtPath string) error {
	p := path.Join(config.Cfg.LocalSaveDir, crtPath)
	err := q.QuarkSaveShareFiles(fidList, fidTokenList, pdirFid, q.Quark.SaveDir.PdirID, shareNo)
	if err != nil {
		return err
	}
	return q.QuarkDownloadAndClear(q.Quark.SaveDir.PdirID, p)
}
