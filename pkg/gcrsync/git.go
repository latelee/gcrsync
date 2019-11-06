/*
 * Copyright © 2018 mritd <mritd1234@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package gcrsync

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/json-iterator/go"
	"github.com/Sirupsen/logrus"
	"github.com/latelee/gcrsync/pkg/utils"
)

/*
组装文件：README.md和ImageList，使用markdown格式。
*/
func (g *Gcr) Commit(images []string) {
	repoDir := strings.Split(g.GithubRepo, "/")[1]
	readmeFile := filepath.Join(repoDir, g.NameSpace)
    err := os.MkdirAll(readmeFile, 0755)
    if err != nil {
        logrus.Errorln(err)
    }
    readmeFile = filepath.Join(readmeFile, ReadmeFile)
	repoUpdateFile := filepath.Join(repoDir, g.NameSpace)
    err = os.MkdirAll(repoUpdateFile, 0755)
    if err != nil {
        logrus.Errorln(err)
    }
    repoUpdateFile = filepath.Join(repoUpdateFile, ImageListFile)

    logrus.Infof("file: %s %s", readmeFile, repoUpdateFile)

	var content []byte
	chgLog, err := os.Open(readmeFile)
    defer chgLog.Close()
    // 如果能打开，则读取已有内容(否则以前的记录会没有)
    if err == nil {
		content, err = ioutil.ReadAll(chgLog)
		utils.CheckAndExit(err)
	}
    // 带创建功能的打开方式
	chgLog, err = os.OpenFile(readmeFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	utils.CheckAndExit(err)
	defer chgLog.Close()

    // 转换成东八区时间
	loc, _ := time.LoadLocation("Asia/Shanghai")
    updateTime := time.Now().In(loc).Format("2006-01-02 15:04:05")
	updateInfo := fmt.Sprintf("### %s Update(num: %d):\n\n", updateTime, len(images))
	for _, imageName := range images {
        // 分离镜像名称和标签
        // TODO：将同一个镜像的所有标签放到一起，不用一一分开，但目前未想到
        tmpImage := strings.Split(imageName, ":")[0]
        //tmpTag := strings.Split(imageName, ":")[1]
        // 添加超链接到hub.docker上，方便查看
		updateInfo += "- " + fmt.Sprintf("[gcr.io/%s/%s](https://hub.docker.com/r/%s/%s/tags)", g.NameSpace, imageName, g.DockerUser, tmpImage) + "\n"
        //updateInfo += "- " + fmt.Sprintf("<a href=\"https://hub.docker.com/r/%s/%s/tags\" target=\"_blank\">gcr.io/%s/%s</a>", g.DockerUser, tmpImage, g.NameSpace, imageName) + "\n"
        //updateInfo += fmt.Sprintf("Tags: [%s]\n", tmpTag)
	}
	chgLog.WriteString(updateInfo + string(content))

    // 如果不存在，则创建文件
	var synchronizedImages []string
	updateFile, err := os.Open(repoUpdateFile)
    defer updateFile.Close()
    content = []byte("[]") // 使用默认的'[]'，否则json解析出错
	if err == nil {
		content, _ = ioutil.ReadAll(updateFile)
	}

	utils.CheckAndExit(jsoniter.Unmarshal(content, &synchronizedImages))
	synchronizedImages = append(synchronizedImages, images...)
	sort.Strings(synchronizedImages)
	buf, err := jsoniter.MarshalIndent(synchronizedImages, "", "    ")
	utils.CheckAndExit(err)
	newUpdateFile, err := os.OpenFile(repoUpdateFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	utils.CheckAndExit(err)
	defer newUpdateFile.Close()
	newUpdateFile.Write(buf)

    logrus.Infof("will commit to github using %s %s\n", g.GithubUser, g.GithubEmail)
	utils.GitCmd(repoDir, "config", "--global", "push.default", "simple")
	utils.GitCmd(repoDir, "config", "--global", "user.name", g.GithubUser)
	utils.GitCmd(repoDir, "config", "--global", "user.email", g.GithubEmail)
	utils.GitCmd(repoDir, "add", ".")
    utils.GitCmd(repoDir, "add", ".", "-u")
	utils.GitCmd(repoDir, "commit", "-m", fmt.Sprintf("Auto sync at %s", updateTime))
	utils.GitCmd(repoDir, "push", "--force", g.commitURL, "master")
}

func (g *Gcr) Clone() {
	os.RemoveAll(strings.Split(g.GithubRepo, "/")[1])
	utils.GitCmd("", "clone", g.commitURL)
}
