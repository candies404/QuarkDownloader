package quark

import (
	"QuarkDownloader/config"
	"log"
)

// FileManager 管理夸克云盘文件的结构体
type FileManager struct {
	Quark   Quark
	Headers map[string]string
	Shares  []*Share
}

type Share struct {
	Url      string
	Passcode string
	PwdId    string
	SToken   string
	Title    string
}

type Quark struct {
	NickName      string
	SaveDir       SaveDir
	TotalCapacity int64
	FreeCapacity  int64
	RootDir       *DirectoryNode // 存储网盘的目录结构树，根节点
}
type SaveDir struct {
	PdirID  string
	DirName string
}
type DirectoryNode struct {
	Name     string // 文件或文件夹的名称
	PwdId    string
	IsDir    bool             // 是否是文件夹
	SizeMB   float64          // 文件大小，单位为MB（如果是文件夹则为0）
	Children []*DirectoryNode // 子节点，表示该文件夹下的子文件或子目录
}

// NewQuarkFileManager 初始化 FileManager 的方法
func NewQuarkFileManager() *FileManager {
	manager := &FileManager{
		Headers: map[string]string{
			"origin":          "https://pan.quark.cn",
			"referer":         "https://pan.quark.cn/",
			"accept-language": "zh-CN,zh;q=0.9",
			"user-agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
			"cookie":          config.Cfg.Quark.Cookie,
		},
		Shares: []*Share{},
		Quark: Quark{
			SaveDir: SaveDir{
				DirName: config.Cfg.Quark.UseDirName,
			},
			RootDir: &DirectoryNode{
				Name:     "根目录",
				PwdId:    "0",
				IsDir:    true,
				SizeMB:   0,
				Children: nil,
			},
		},
	}
	for _, link := range config.Cfg.SharesLinks {
		manager.Shares = append(manager.Shares, &Share{
			Url:      link.Link.URL,
			Passcode: link.Link.PassCode,
			PwdId:    "",
			SToken:   "",
			Title:    "",
		})
	}
	err := manager.QuarkGetPanInfo()
	if err == nil {
		log.Printf("登录用户： %s ，容量：  %d / %d GB\n", manager.Quark.NickName, (manager.Quark.TotalCapacity-manager.Quark.FreeCapacity)/(1024*1024*1024), manager.Quark.TotalCapacity/(1024*1024*1024))
		err := manager.QuarkGetFileList(false, "0", manager.Quark.RootDir, 0)
		if manager.Quark.SaveDir.PdirID == "" {
			fid, err := manager.QuarkCreateFolder("0", manager.Quark.SaveDir.DirName)
			if err != nil {
				panic(err)
				return nil
			}
			manager.Quark.SaveDir.PdirID = fid
		}
		if err != nil {
			return nil
		}
		manager.QuarkGetShares()
		return manager
	}
	return nil
}
