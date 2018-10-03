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
	"net/http"
	"sync"
	"time"
    "strings"

	"github.com/docker/docker/client"

	"github.com/Sirupsen/logrus"
	"github.com/json-iterator/go"
	"github.com/latelee/gcrsync/pkg/utils"
)

type Image struct {
	Name string
	Tags []string
}

type Gcr struct {
	Proxy          string
	DockerUser     string
	DockerPassword string
	NameSpace      string
	GithubToken    string
	GithubRepo     string
    GithubUser     string
    GithubEmail    string
	CommitMsg      string
	MonitorCount   int
	TestMode       bool
	MonitorMode    bool
	Debug          bool
	QueryLimit     chan int
	ProcessLimit   chan int
    ProcessCount   int
	HttpTimeOut    time.Duration
	httpClient     *http.Client
	dockerClient   *client.Client
	dockerHubToken string
	update         chan string
	commitURL      string
}

func (g *Gcr) gcrImageList() []string {

	var images []string
	publicImageNames := g.gcrPublicImageNames()

	logrus.Infof("gcrImageList() Number of gcr images: %d", len(publicImageNames))

	imgNameCh := make(chan string, 20)
	imgGetWg := new(sync.WaitGroup)
	imgGetWg.Add(len(publicImageNames))

	for _, imageName := range publicImageNames {

		tmpImageName := imageName

		go func() {
			defer func() {
				g.QueryLimit <- 1
				imgGetWg.Done()
			}()

			select {
			case <-g.QueryLimit:
				req, err := http.NewRequest("GET", fmt.Sprintf(GcrImageTags, g.NameSpace, tmpImageName), nil)
				utils.CheckAndExit(err)

				resp, err := g.httpClient.Do(req)
				utils.CheckAndExit(err)

				b, err := ioutil.ReadAll(resp.Body)
				utils.CheckAndExit(err)
				resp.Body.Close()

				var tags []string
				jsoniter.UnmarshalFromString(jsoniter.Get(b, "tags").ToString(), &tags)
                //logrus.Infof("gcrImageList() 102 image %s, tags:%s", tmpImageName, tags)
                // 去掉一些标签：包括alpha、beta、rc，等等，这些认为是测试版本
				for _, tag := range tags {
                    if len(tag) > 12 || strings.Contains(tag, "alpha") || 
                    strings.Contains(tag, "beta") || strings.Contains(tag, "rc") || 
                    strings.Contains(tag, "test") {
                    //logrus.Infof("gcrImageList() 107 image %s", tag)
                    continue
                    }
					imgNameCh <- tmpImageName + ":" + tag
                    //logrus.Infof("gcrImageList() 109 image %s", tag)
				}
			}
		}()
	}

	var imgReceiveWg sync.WaitGroup
	imgReceiveWg.Add(1)
	go func() {
		defer imgReceiveWg.Done()
		for {
			select {
			case imageName, ok := <-imgNameCh:
				if ok {
					images = append(images, imageName)
				} else {
					goto imgSetExit
				}
			}
		}
	imgSetExit:
	}()

	imgGetWg.Wait()
	close(imgNameCh)
	imgReceiveWg.Wait()
    
    return images
    /*
    i := 0
    var out []string
    for _, tmp := range images {
        i++
        //logrus.Infof("cnt: %d\n", i)
        if i > 3 {
            break
        }
        out = append(out , tmp)
        
    }
    //logrus.Infof("output: %s", out)
	return out
    */
}

/*
获取指定命名空间（g.NameSpace）下的镜像名称（只是名称，还没有tag）
*/
func (g *Gcr) gcrPublicImageNames() []string {

    // GET请求，https://gcr.io/v2/google-containers/tags/list
    // 与curl https://gcr.io/v2/google-containers/tags/list的结果应该是一样的
	req, err := http.NewRequest("GET", fmt.Sprintf(GcrImages, g.NameSpace), nil)
	utils.CheckAndExit(err)

	resp, err := g.httpClient.Do(req)
	utils.CheckAndExit(err)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	utils.CheckAndExit(err)

	var imageNames []string
	jsoniter.UnmarshalFromString(jsoniter.Get(b, "child").ToString(), &imageNames)
    
    logrus.Infof("gcrPublicImageNames() Number of gcr images: %d", len(imageNames))

    // 去掉arm、ppc、s390x版本的镜像，——因为加上这些，镜像会非常多
    var outtmp []string
    for _, tmp := range imageNames {
        tmpImageName := tmp
        if strings.Contains(tmpImageName, "-arm") || strings.Contains(tmpImageName, "-ppc") ||
        strings.Contains(tmpImageName, "-s390x") {
            //logrus.Infof("gcrPublicImageNames() 188 image %s", tmpImageName)
            continue
            }
        outtmp = append(outtmp , tmp)
    }
    
    return outtmp
}
