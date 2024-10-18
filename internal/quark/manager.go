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
			"cookie":          "__itrace_wid=af3beb62-6606-4b07-34e6-cb2e120e8e62; b-user-id=257451cf-4b28-3d02-6330-d6f4a424bad9; _UP_A4A_11_=wb96a19fbaaa48f1985f73cc42068133; _UP_D_=pc; __pus=f9ac057c711a562a8372fd54dd7dc5bdAAT7U705+CDW/JFDvb21QehmzBwzMhzuHzHes4kkAKdFtRO3IIkTJjMTvsHpNpcmxhGmYBY1l8QocHtR7QVEPoEG; __kp=e7f70ce0-8ac6-11ef-8548-b9999184522d; __kps=AATv3HaD3AL4tr38jcVigUDB; __ktd=Now14efqEIZmhH45IyeobQ==; __uid=AATv3HaD3AL4tr38jcVigUDB; omelette-vid=289772429351748962941705; omelette-vid.sig=wguHlB6BSsRH5VPPL0v9xqDKFc1Pi5q1Z1w7wsq9yoc; __puus=4fe6bbdf1fb638dc2e7296590e38a9a1AARl2RXU31zfVWOeQRpiVS5HZI0qrnwQRNJOudNoLaw3ZdK/mVm57sbn+Dm8DyclfKyUJFb5bYUcYZ/OxmkxpG8CQwoywlqcBwnnc82SAIr+GiJy8/woNuDcwUZNWF44InRfoj/1aUC/ayDsQGKBrKnf60MNPM20ssPcOdZmbMD+bbqC9SvIPwxRL6J4wRBYPygJ+ghUS1mrHU5lZU3kF+Gh; tfstk=gMar0jZpDaQz_qbsGj3E_G8lDL3-A4X1KyMItWVnNYDoRgEU0WNNwXw7tvrEnWHkEeYoYjct6LAkq9HE3x20NUMIrvAUdWZSETKkixcUI6hntYxemfeBZe4U9el3tJC-d_Ibw73K-O6_LNN8w5lpXC6j-6VmJ3iK5MQ_w7dDapf_mN9EPQtgEv2nqqAmTxck-b2HmmDx9XvotUf4ixHDrLDo-IAm6fpHqJ0hgSDxtE7_Vktqs0f6C0MEC1hj4AVo3FPYubDp27D2SF4400kMRx8H-rlzGYjz_FSIIlHSfXeclUuUiXyZ_8JcGxq4iznsqC8aQ4yEjzm2tFNi-o3qiPfBXxaqFJ4rm1TZAq4sj4qfDTiIzfy3y8Re-RVb1znY8OJmpuGtobyhStSy1phDlXzL4pxEqjhqCs5VX65eOm1-sntpvmDZgA1-eHKK4L8B4PcMvHnmQjk1wYC..; isg=BMbGvKLy3cD1D4k5AQd5080tF7xIJwrhmDWXiLDvienEs2bNGLU88f5Bj-9_GwL5",
			"user-agent":      config.Cfg.Quark.Cookie,
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
