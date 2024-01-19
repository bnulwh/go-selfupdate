package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kr/binarydist"
	"github.com/bnulwh/go-selfupdate/selfupdate"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type current struct {
	Version string
	Sha256  []byte
}

type gzReader struct {
	z, r io.ReadCloser
}

func (g *gzReader) Read(p []byte) (int, error) {
	return g.z.Read(p)
}

func (g *gzReader) Close() error {
	g.z.Close()
	return g.r.Close()
}

func newGzReader(r io.ReadCloser) io.ReadCloser {
	var err error
	g := new(gzReader)
	g.r = r
	g.z, err = gzip.NewReader(r)
	if err != nil {
		panic(err)
	}
	return g
}

var servePath = flag.String("dir", "./public", "path to serve")

func main() {
	gin.SetMode(gin.ReleaseMode)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	router := gin.Default()
	//pprof.Register(router, "/pprof")
	router.StaticFS("/", http.Dir(*servePath))
	router.POST("/upload", PostUpload)
	router.Run(":8080")

}

func PostUpload(ctx *gin.Context) {
	file, _ := ctx.FormFile("file")
	filePath := "/tmp/" + file.Filename
	err := ctx.SaveUploadedFile(file, filePath)
	if err != nil {
		log.Println("save file failed:", err)
		ctx.String(http.StatusOK, "上传失败")
		return
	}
	version := ctx.PostForm("version")
	signature := ctx.PostForm("signature")
	platform := ctx.PostForm("platform")
	//filename := ctx.PostForm("filename")
	sign2 := selfupdate.GenerateSha256(filePath)
	if sign2 != signature {
		log.Println("get file failed: wrong signature")
		ctx.String(http.StatusOK, "校验失败")
		return
	}
	createUpdate(filePath, platform, version, signature)
	ctx.String(http.StatusOK, "上传成功")
}

func createUpdate(path, platform, version, signature string) {
	genDir := filepath.Join(*servePath, filepath.Base(path))
	os.MkdirAll(genDir, 0755)
	sh256, err := base64.URLEncoding.DecodeString(signature)
	c := current{Version: version, Sha256: sh256}

	b, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		fmt.Println("error:", err)
	}
	err = ioutil.WriteFile(filepath.Join(genDir, platform+".json"), b, 0755)
	if err != nil {
		panic(err)
	}

	os.MkdirAll(filepath.Join(genDir, version), 0755)

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	f, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	w.Write(f)
	w.Close() // You must close this first to flush the bytes to the buffer.
	err = ioutil.WriteFile(filepath.Join(genDir, version, platform+".gz"), buf.Bytes(), 0755)

	files, err := ioutil.ReadDir(genDir)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		if file.IsDir() == false {
			continue
		}
		if file.Name() == version {
			continue
		}

		os.Mkdir(filepath.Join(genDir, file.Name(), version), 0755)

		fName := filepath.Join(genDir, file.Name(), platform+".gz")
		old, err := os.Open(fName)
		if err != nil {
			// Don't have an old release for this os/arch, continue on
			continue
		}

		fName = filepath.Join(genDir, version, platform+".gz")
		newF, err := os.Open(fName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't open %s: error: %s\n", fName, err)
			os.Exit(1)
		}

		ar := newGzReader(old)
		defer ar.Close()
		br := newGzReader(newF)
		defer br.Close()
		patch := new(bytes.Buffer)
		if err := binarydist.Diff(ar, br, patch); err != nil {
			panic(err)
		}
		ioutil.WriteFile(filepath.Join(genDir, file.Name(), version, platform), patch.Bytes(), 0755)
	}
}
