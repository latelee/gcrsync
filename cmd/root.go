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

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/latelee/gcrsync/pkg/gcrsync"

	"github.com/spf13/cobra"
)

var debug, test bool
var proxy, dockerUser, dockerPassword, nameSpace string
var githubRepo, githubUser, githubEmail, githubToken string
var queryLimit, processLimit, processCount, monitorCount int
var httpTimeout time.Duration

var rootCmd = &cobra.Command{
	Use:   "gcrsync",
	Short: "A docker image sync tool for Google container registry (gcr.io)",
	Long: `
A docker image sync tool for Google container registry (gcr.io).`,
	Run: func(cmd *cobra.Command, args []string) {

		gcr := &gcrsync.Gcr{
			Proxy:          proxy,
			DockerUser:     dockerUser,
			DockerPassword: dockerPassword,
			NameSpace:      nameSpace,
			QueryLimit:     make(chan int, queryLimit),
			ProcessLimit:   make(chan int, processLimit),
            ProcessCount:   processCount,
			HttpTimeOut:    httpTimeout,
			GithubRepo:     githubRepo,
            GithubUser:     githubUser,
            GithubEmail:    githubEmail,
			GithubToken:    githubToken,
			MonitorCount:   monitorCount,
			TestMode:       test,
			Debug:          debug,
		}
		gcr.Init()
		gcr.Sync()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug mode")
	rootCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "http client proxy")
	rootCmd.PersistentFlags().StringVar(&dockerUser, "dockeruser", "", "docker registry user")
	rootCmd.PersistentFlags().StringVar(&dockerPassword, "dockerpassword", "", "docker registry user password")
	rootCmd.PersistentFlags().StringVar(&nameSpace, "namespace", "google-containers", "google container registry namespace")
	rootCmd.PersistentFlags().IntVar(&queryLimit, "querylimit", 50, "http query limit")
	rootCmd.PersistentFlags().DurationVar(&httpTimeout, "httptimeout", 10*time.Second, "http request timeout")
	rootCmd.PersistentFlags().IntVar(&processLimit, "processlimit", 20, "image process limit")
    rootCmd.PersistentFlags().IntVar(&processCount, "processcount", 100, "image process total count(-1 means all in grc.io)")
	rootCmd.PersistentFlags().StringVar(&githubRepo, "githubrepo", "latelee/gcr.io", "github commit repo")
    rootCmd.PersistentFlags().StringVar(&githubUser, "githubuser", "Late Lee", "github commit user name")
    rootCmd.PersistentFlags().StringVar(&githubEmail, "githubemail", "li@latelee.org", "github commit email")
	rootCmd.PersistentFlags().StringVar(&githubToken, "githubtoken", "", "github commit token(must specify)")
}
