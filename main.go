package main

import (
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

func showUsage() {
	fmt.Println(`
	ACTIONS:
		add 
	
	`)
}
func isDir(path string) {

}

const (
	IDX_NAME    = "filenote"
	PRIMARY_KEY = "hash"
)

var cindex *meilisearch.Index
var (
	client = meilisearch.NewClient(meilisearch.ClientConfig{
		Host: "http://127.0.0.1:7700",
		// APIKey: "masterKey",
	})
	accepted_extensions = []string{".txt", ".md"}
)

func accepted(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, v := range accepted_extensions {
		if strings.Compare(v, ext) == 0 {
			return true
		}
	}
	return false
}

func init() {
	var err error

	task, err := client.CreateIndex(&meilisearch.IndexConfig{Uid: IDX_NAME, PrimaryKey: PRIMARY_KEY})
	if err != nil {
		fmt.Println("creating index error:", err)
		os.Exit(1)
	}
	wait_task, _ := client.WaitForTask(task, meilisearch.WaitParams{
		Context:  context.Background(),
		Interval: time.Duration(1) * time.Second,
	})
	fmt.Println("wait task status:", wait_task.Status, wait_task.Error, wait_task.Type)
	cindex, err = client.GetIndex(IDX_NAME)
	if err != nil {
		fmt.Println("getting index error:", err)
		os.Exit(1)
	}
}
func submitFile1(path string) {
	fmt.Println("submitting:", path)
}
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

//submit file to meilisearch engine.
func submitFile(path string) {
	var err error
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("error occurs when reading file:", path)
		return
	}
	st, _ := os.Stat(path)
	h := hash(path)
	documents := []map[string]interface{}{
		{"hash": h, "path": path, "updated": st.ModTime(), "content": string(content)},
	}
	if cindex == nil {
		fmt.Println("cindex is nil")
		os.Exit(1)
	}
	task, _ := cindex.AddDocuments(documents)
	wait_task, _ := client.WaitForTask(task, meilisearch.WaitParams{
		Context:  context.Background(),
		Interval: time.Duration(1) * time.Second,
	})
	fmt.Println("wait task status:", wait_task.Status, wait_task.Error)

}
func main() {
	flag.Parse()
	fmt.Println(flag.Args())
	for _, p := range flag.Args() {
		_, err := os.Stat(p)
		if err != nil && os.IsNotExist(err) {
			fmt.Printf("path:%s not exists\n", p)
			continue
		} else {
			ap, _ := filepath.Abs(p)
			filepath.Walk(ap, func(path string, info fs.FileInfo, err error) error {

				if err != nil {
					fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
					return err
				}
				fmt.Println("fucking ", info.Name())
				//ignore hidden file or directory
				if strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}
				if info.Mode().IsRegular() && accepted(path) {
					fmt.Println("scaning path:", path)
					submitFile(path)
				}
				return nil

			})
		}
	}
	if err := recover(); err != nil {

	}
}
