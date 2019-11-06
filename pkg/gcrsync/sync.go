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
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "sync"
    "time"

    "github.com/json-iterator/go"

    "github.com/Sirupsen/logrus"
    "github.com/docker/docker/client"

    "github.com/latelee/gcrsync/pkg/utils"
)

const (
    ImageListFile   = "ImageList"
    ReadmeFile      = "README.md"
    GcrRegistryTpl = "gcr.io/%s/%s"
    GcrImages      = "https://gcr.io/v2/%s/tags/list"
    GcrImageTags   = "https://gcr.io/v2/%s/%s/tags/list"
    RegistryTag    = "https://hub.docker.com/v2/repositories/%s/%s/tags/%s/"
)

var CntIter int
var CntTotal int

func (g *Gcr) Sync() {

    gcrImages := g.gcrImageList()

    needSyncImages := g.compareCache(gcrImages)

    //logrus.Infof("Sync() Google container registry images total: %d %s", len(gcrImages), gcrImages)
    logrus.Infof("Sync() Google container registry images total: %d", len(gcrImages))
    

    // 考虑到travis-ci一次性处理不了那么多（所有gcr.io镜像可达几千个），次数可以由命令行传递进来
    // 如果为-1，则处理所有的（但可能会失败）
    if g.ProcessCount == -1 {
        CntTotal = len(needSyncImages)
    } else {
        if len(needSyncImages) < g.ProcessCount {
            CntTotal = len(needSyncImages)
        } else {
            CntTotal = g.ProcessCount
        }
    }
    processWg := new(sync.WaitGroup)
    processWg.Add(CntTotal)

    logrus.Infof("Sync() Number of images waiting to be processed: %d", CntTotal)
    
    i := 0
    var out []string
    for _, tmp := range needSyncImages {
        i++
        //logrus.Infof("cnt: %d\n", i)
        if i > CntTotal {
            break
        }
        out = append(out , tmp)
    }
    
    for _, imageName := range out {
        tmpImageName := imageName
        go func() {
            defer func() {
                g.ProcessLimit <- 1
                processWg.Done()
            }()
            select {
            case <-g.ProcessLimit:
                g.Process(tmpImageName)
            }
        }()
    }

    logrus.Infof("done process, will generate doc")
    // doc gen
    chgWg := new(sync.WaitGroup)
    chgWg.Add(1)
    go func() {
        defer chgWg.Done()

        var images []string
        for {
            select {
            case imageName, ok := <-g.update:
                if ok {
                    images = append(images, imageName)
                } else {
                    goto ReadmeFileDone
                }
            }
        }
    ReadmeFileDone:
        if len(images) > 0 && !g.TestMode {
            g.Commit(images)
        }
    }()

    processWg.Wait()
    close(g.update)
    chgWg.Wait()

}

func (g *Gcr) Monitor() {

    if g.MonitorCount == -1 {
        for {
            select {
            case <-time.Tick(5 * time.Second):
                gcrImages := g.gcrImageList()
                needSyncImages := g.compareCache(gcrImages)
                logrus.Infof("Gcr images: %d    Waiting process: %d", len(gcrImages), len(needSyncImages))
            }
        }
    } else {
        for i := 0; i < g.MonitorCount; i++ {
            select {
            case <-time.Tick(5 * time.Second):
                gcrImages := g.gcrImageList()
                needSyncImages := g.compareCache(gcrImages)
                logrus.Infof("Gcr images: %d    Waiting process: %d", len(gcrImages), len(needSyncImages))
            }
        }
    }

}

func (g *Gcr) Compare() {
    gcrImages := g.gcrImageList()
    needSyncImages := g.needProcessImages(gcrImages)

    logrus.Infof("Compare() Google container registry images total: %d %s", len(gcrImages), gcrImages)
    logrus.Infof("222 Number of images waiting to be processed: %d", len(needSyncImages))

    diff := utils.SliceDiff(gcrImages, needSyncImages)
    sort.Strings(diff)
    repoDir := strings.Split(g.GithubRepo, "/")[1]
    f, err := os.OpenFile(filepath.Join(repoDir, g.NameSpace), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
    utils.CheckAndExit(err)
    defer f.Close()
    b, err := jsoniter.MarshalIndent(diff, "", "    ")
    utils.CheckAndExit(err)
    f.Write(b)
}

func (g *Gcr) Init() {

    if g.Debug {
        logrus.SetLevel(logrus.DebugLevel)
    }

    logrus.Infoln("111 Init http client.")
    g.httpClient = &http.Client{
        Timeout: g.HttpTimeOut,
    }
    if g.Proxy != "" {
        p := func(_ *http.Request) (*url.URL, error) {
            return url.Parse(g.Proxy)
        }
        g.httpClient.Transport = &http.Transport{Proxy: p}
    }

    logrus.Infoln("Init docker client.")
    dockerClient, err := client.NewEnvClient()
    utils.CheckAndExit(err)
    g.dockerClient = dockerClient

    logrus.Infoln("Init limit channel.")
    for i := 0; i < cap(g.QueryLimit); i++ {
        g.QueryLimit <- 1
    }
    for i := 0; i < cap(g.ProcessLimit); i++ {
        g.ProcessLimit <- 1
    }

    logrus.Infoln("Init update channel.")
    g.update = make(chan string, 20)

    logrus.Infof("Init commit repo: %s", g.GithubRepo)
    if g.GithubToken == "" {
        utils.ErrorExit("Github Token is blank!", 1)
    }
    g.commitURL = "https://" + g.GithubToken + "@github.com/" + g.GithubRepo + ".git"
    g.Clone()

    logrus.Infoln("Init success...")
}
