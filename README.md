# gcrsync  
[![Build Status](https://www.travis-ci.org/latelee/gcrsync.svg?branch=master)](https://www.travis-ci.org/latelee/gcrsync)

A docker image sync tool for Google container registry (gcr.io)

gcr.io(Google container registry)docker镜像同步工具


## 安装
工具使用go语言编写，需要预先准备好go编译环境，在本地构建命令如下：
```bash
cd $(GOPATH)/src
go get github.com/latelee/gcrsync
cd gcrsync
go build
```
注意！由于gcr.io一般无法直接访问，因此，需要在诸如travis-ci等服务器（位于国外）才能正常执行本工具。  
本工具仅在[travis-ci](https://www.travis-ci.org/latelee/gcrsync)测试通过，本地仅做编译工作（以保证构建成功）。
## 使用
编译成功得到的可执行文件为gcrsync，使用帮助信息如下：  
```bash
A docker image sync tool for Google container registry (gcr.io).

Usage:
  gcrsync [flags]
  gcrsync [command]

Available Commands:
  compare     Compare gcr registry and private registry
  help        Help about any command
  monitor     Monitor sync images
  sync        Sync gcr images
  test        Test sync

Flags:
      --debug                   debug mode
      --dockerpassword string   docker registry user password
      --dockeruser string       docker registry user
      --githubemail string      github commit email (default "li@latelee.org")
      --githubrepo string       github commit repo (default "latelee/gcr.io")
      --githubtoken string      github commit token(must specify)
      --githubuser string       github commit user name (default "Late Lee")
  -h, --help                    help for gcrsync
      --httptimeout duration    http request timeout (default 10s)
      --namespace string        google container registry namespace (default "google-containers")
      --processcount int        image process total count(-1 means all in grc.io) (default 100)
      --processlimit int        image process limit (default 20)
      --proxy string            http client proxy
      --querylimit int          http query limit (default 50)

Use "gcrsync [command] --help" for more information about a command.
```
注意，--dockeruser、--dockerpassword、--githubrepo、--githubtoken等，务必根据实际情况修改。

### sync 命令
该命令进行真正的同步操作，目前仅对此命令进行全面测试，其工作流程如下:  
- 首先使用 `--githubtoken` 给定的 token 克隆 `--githubrepo` 给定的仓库到本地(这个仓库即为元数据仓库)
- 获取 `gcr.io` 下由 `--namespace` 给定命名空间下的所有镜像，但是过滤掉测试标签（带alpha/beta/rc等字眼），也过滤arm/ppc等平台版本
- 读取克隆的元数据仓库内对应的命名空间的ImageList文件
- 对比两者差异，得出待同步镜像
- 执行 `pull`、`tag`、`push` 操作，将其推送到由 `--user` 给定的Docker Hub用户仓库中
- 追加元数据仓库内对应的命名空间 的ImageList文件
- 生成 README.md 并推送元数据到指定远程仓库

本工具实际使用命令如下：
```
- gcrsync sync --namespace google-containers --querylimit 20 --processlimit 50 --httptimeout 10s --processcount 200 --dockeruser gcrcontainer --dockerpassword ${DOCKER_PASSWORD} --githubrepo latelee/gcr.io --githubtoken ${GITHUB_TOKEN}
```
其中，${DOCKER_PASSWORD}、${GITHUB_TOKEN}因为是敏感信息，使用travis-ci的环境变量（参考travis-ci官方使用文档），所以不要泄漏。--processcount指定为200，表示一次执行处理的gcr.io镜像为200个，笔者实际执行中发现，超过此数，travis-ci会执行失败，如果要全部同步，将其置为-1即可。  

### compare 命令

该命令用于对比 Docker Hub 指定用户镜像与 `gcr.io` 指定 namesapce 下的镜像差异，同时生成已经同步的 json 文件

### monitor 命令

该命令与 compare 类似，只不过不生成任何文件，实时在控制台显示对比差异；一般用于监测同步进度

### test 命令

该命令与 sync 命令基本行为一致，只不过不进行真正的同步，会生成 CHANGELOG，但不会推送到远程

## 其他说明
最终镜像文件位于：[https://hub.docker.com/r/gcrcontainer/]<https://hub.docker.com/r/gcrcontainer/>。  
镜像列表说明位于：[https://github.com/latelee/gcr.io]<https://github.com/latelee/gcr.io>，gcr.io的命名空间与应该仓库的目录一一对应，内有README.md，列出镜像及其标签（tag），并链接到Docker Hub账号对应的镜像地址。  
traivs-ci一次处理数据不多，gcr.io的google-containers镜像不到3000个，一共执行了很多次才全部同步完成。  
本工具每天定时执行一次，但不能保证gcr.io指定命名空间下所有镜像都能同步到Docker Hub上。请谨慎使用。    

## 致谢
本工具原作者为[mritd](https://github.com/mritd)，本仓库源码由[李迟](https://github.com/latleee)根据使用情况做了适当的修改。  