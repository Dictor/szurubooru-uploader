package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
)

type (
	Folder struct {
		Name   string
		Number int
		Path   string
	}
)

func main() {
	var (
		targetDir        string
		executeDir       string
		host             string
		safety           string
		userId, userPass string
		Folders          []Folder = []Folder{}
	)

	flag.StringVar(&targetDir, "dir", "", "target directory path")
	flag.StringVar(&executeDir, "exe", "./szurubooru-uploader", "path of szurubooru-uploader execute")
	flag.StringVar(&host, "host", "localhost", "path to szurubooru server")
	flag.StringVar(&safety, "safety", "unsafe", "safety level for uploding images")
	flag.StringVar(&userId, "uid", "", "user's id")
	flag.StringVar(&userPass, "upw", "", "user's password")
	flag.Parse()

	filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		var (
			Name   string
			Number int
		)
		if d.IsDir() {
			n, err := fmt.Sscanf(d.Name(), "%s (%d)", &Name, &Number)
			if n != 2 || err != nil {
				fmt.Printf("fail to parse path '%s'\n", path)
				return nil
			}
			Folders = append(Folders, Folder{
				Name:   Name,
				Number: Number,
				Path:   path,
			})
		}
		return nil
	})

	fmt.Printf("%d folders are parsed\n", len(Folders))

	for _, f := range Folders {
		cmd := exec.Command(executeDir, "-dir", f.Path, "-host", host, "-safety", safety, "-tag", fmt.Sprintf("%s(%d)", strings.Replace(f.Name, " ", "_", -1), f.Number), "-uid", userId, "-upw", userPass)
		fmt.Println(cmd.String())
	}
}
