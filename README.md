# Quark File Downloader

### 不能突破大小限制，只能实现小文件的转存、下载
### 不能突破大小限制，只能实现小文件的转存、下载
### 不能突破大小限制，只能实现小文件的转存、下载

### 尽量使用没有存储资源的网盘，操作不慎可能导致网盘重要文件丢失
### 尽量使用没有存储资源的网盘，操作不慎可能导致网盘重要文件丢失
### 尽量使用没有存储资源的网盘，操作不慎可能导致网盘重要文件丢失

[<img src="https://api.gitsponsors.com/api/badge/img?id=874895549" height="90">](https://api.gitsponsors.com/api/badge/link?p=fz6/r2Ybi1Xr7qbqvFyX1mYyGYIjE1bFtv0spH/mDSxTkvG7LjPQy2AdwL0VKpm23E/wjjcTGaNXyLFyI2QF7svduM4VSHkI06lIKWjGNbk/9j11Acsxo/m3A5417J80nvZO7H/qkOstvJTDZG6eLw==)

## 前沿
之前在找音乐资源的时候，发现别人分享的连接 **目录结果复杂**、**文件众多**、**云盘大小的限制**，导致下载过程十分繁琐
本项目的初衷是获取他人分享的云盘连接

## 概述
本项目是一个自动化下载工具，用于从 Quark 云盘上批量下载文件，并根据配置文件中的过滤规则下载特定文件类型。下载后的文件将被保存到本地指定的目录中。

## 配置说明

配置文件`config.yaml`结构如下：

```yaml
quark:
    useDirName: "临时存储文件"      # Quark 云盘临时存储文件的文件夹名（最好是没有的文件名，因为下载过程中会清空这个文件夹）
    cookie: ""                      # 你的 Quark 云盘 cookie

sharesLinks:
    - link:
        url: "https://pan.quark.cn/s/787b447799dc"  # Quark 分享链接
        passCode: ""                               # 如果有密码，填写密码

localSaveDir: "download"        # 本地保存下载文件的目录
delay: 0.5                      # 每个文件下载之间的延迟时间（秒）
downloadFilter:
    include: ""                   # 包含的文件类型过滤正则表达式
    exclude: ".jpg"               # 排除的文件类型过滤正则表达式
proxy: ""                       # 全局代理配置
```

## 使用方法

1. 配置好 `config.yaml` 文件，确保填入正确的 Quark 云盘 cookie 和分享链接。
2. 运行程序，程序会自动根据配置下载符合过滤条件的文件。
