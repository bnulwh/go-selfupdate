package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/bnulwh/go-selfupdate/selfupdate"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func uploadFile(url, filename, platform, version string) error {
	var buf bytes.Buffer
	signature := selfupdate.GenerateSha256(filename)
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		log.Println("创建表单失败: ", err)
		return err
	}
	reader, err := os.Open(filename)
	if err != nil {
		log.Println("打开文件失败: ", err)
		return err
	}
	defer reader.Close()

	_, err = io.Copy(part, reader)
	if err != nil {
		log.Println("复制文件内容失败:", err)
		return err
	}
	writer.WriteField("folder", filepath.Dir(filename))
	writer.WriteField("filename", filename)
	writer.WriteField("platform", platform)
	writer.WriteField("version", version)
	writer.WriteField("signature", signature)
	//writer.WriteField("genDir", genDir)
	err = writer.Close()
	if err != nil {
		log.Println("关闭写入失败:", err)
		return err
	}
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		log.Println("创建请求失败:", err)
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		log.Println("发送请求失败:", err)
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取响应失败:", err)
		return err
	}
	log.Println("Response:", string(respBody))
	return nil

}

func printUsage() {
	fmt.Println("")
	fmt.Println("Positional arguments:")
	fmt.Println("\tSingle platform: go-selfupdate myapp 1.2")
	fmt.Println("\tCross platform: go-selfupdate /tmp/mybinares/ 1.2")
}

func main() {
	//outputDirFlag := flag.String("o", "public", "Output directory for writing updates")

	var defaultPlatform string
	goos := os.Getenv("GOOS")
	goarch := os.Getenv("GOARCH")
	if goos != "" && goarch != "" {
		defaultPlatform = goos + "-" + goarch
	} else {
		defaultPlatform = runtime.GOOS + "-" + runtime.GOARCH
	}
	platformFlag := flag.String("platform", defaultPlatform,
		"Target platform in the form OS-ARCH. Defaults to running os/arch or the combination of the environment variables GOOS and GOARCH if both are set.")

	uploadFlag := flag.String("dest", "http://localhost:8080/upload", "upload url")

	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		printUsage()
		os.Exit(0)
	}

	platform := *platformFlag
	appPath := flag.Arg(0)
	version := flag.Arg(1)
	uploadUrl := *uploadFlag
	//genDir = filepath.Join(*outputDirFlag, appPath)

	log.Println("platform:", platform)
	log.Println("app path:", appPath)
	log.Println("version:", version)
	//log.Println("gen dir:", genDir)
	//createBuildDir()

	// If dir is given create update for each file
	fi, err := os.Stat(appPath)
	if err != nil {
		panic(err)
	}

	if fi.IsDir() {
		files, err := ioutil.ReadDir(appPath)
		if err == nil {
			for _, file := range files {
				uploadFile(uploadUrl, filepath.Join(appPath, file.Name()), platform, version)
			}
			os.Exit(0)
		}
	}

	uploadFile(uploadUrl, appPath, platform, version)
}
