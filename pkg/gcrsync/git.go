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

func (g *Gcr) Commit(images []string) {
	repoDir := strings.Split(g.GithubRepo, "/")[1]
	repoChangeLog := filepath.Join(repoDir, g.NameSpace)
    logrus.Infof("file111: %s", repoChangeLog)
    err := os.MkdirAll(repoChangeLog, 0755)
    if err != nil {
        logrus.Errorln(err)
    }
    repoChangeLog = filepath.Join(repoChangeLog, ChangeLog)
    logrus.Infof("file222: %s", repoChangeLog)
    
	repoUpdateFile := filepath.Join(repoDir, g.NameSpace)
    err = os.MkdirAll(repoUpdateFile, 0755)
    if err != nil {
        logrus.Errorln(err)
    }
    repoUpdateFile = filepath.Join(repoUpdateFile, g.NameSpace)

    logrus.Infof("file: %s %s", repoChangeLog, repoUpdateFile)

	var content []byte
	chgLog, err := os.Open(repoChangeLog)
    logrus.Errorln(err)
    defer chgLog.Close()
    // 如果能打开，则读取已有内容
    if err == nil {
		content, err = ioutil.ReadAll(chgLog)
		utils.CheckAndExit(err)
	}
    // 带创建功能的打开方式
	chgLog, err = os.OpenFile(repoChangeLog, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	utils.CheckAndExit(err)
	defer chgLog.Close()

	loc, _ := time.LoadLocation("Asia/Shanghai")
    updateTime := time.Now().In(loc).Format("2006-01-02 15:04:05")
	updateInfo := fmt.Sprintf("### %s Update:\n\n", updateTime)
	for _, imageName := range images {
		updateInfo += "- " + fmt.Sprintf(GcrRegistryTpl, g.NameSpace, imageName) + "\n"
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

    logrus.Infof("will commit to github %s %s\n", g.GithubUser, g.GithubEmail)
	utils.GitCmd(repoDir, "config", "--global", "push.default", "simple")
	utils.GitCmd(repoDir, "config", "--global", "user.email", g.GithubUser)
	utils.GitCmd(repoDir, "config", "--global", "user.name", g.GithubEmail)
	utils.GitCmd(repoDir, "add", ".")
    utils.GitCmd(repoDir, "add", ".", "-u")
	utils.GitCmd(repoDir, "commit", "-m", fmt.Sprintf("Auto sync at %s", updateTime))
	utils.GitCmd(repoDir, "push", "--force", g.commitURL, "master")
}

func (g *Gcr) Clone() {
	os.RemoveAll(strings.Split(g.GithubRepo, "/")[1])
	utils.GitCmd("", "clone", g.commitURL)
}
