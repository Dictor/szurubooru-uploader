package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func execUpload(cmd *cobra.Command, args []string) {
	var (
		err error
	)

	filePaths := []string{}
	err = filepath.Walk(args[0], func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		for _, e := range ImageExtension {
			if filepath.Ext(info.Name()) == e {
				filePaths = append(filePaths, path)
			}
		}
		return nil
	})
	if err != nil {
		Logger.WithError(err).Errorln("error caused during walking directory")
	}
	Logger.Infof("%d images will be uploaded", len(filePaths))

	logError := func(cur, total int, err error, path, action string) {
		Logger.WithFields(logrus.Fields{
			"error": err,
			"path":  path,
		}).Errorf("(%d/%d) error : %s\n", cur+1, total, action)
	}

	for i, path := range filePaths {
		// upload temporary image file
		ftok, err := uploadFile(host, userToken, path)
		if err != nil {
			logError(i, len(filePaths), err, path, "upload file")
			continue
		}
		// request reverse search
		rev, err := reverseSearch(host, userToken, ftok)
		if err != nil {
			logError(i, len(filePaths), err, path, "search similar post")
			continue
		}
		// create post
		if err := createPost(host, userToken, ftok, args[1], safety, rev); err != nil {
			logError(i, len(filePaths), err, path, "create post")
			continue
		}
		Logger.Infof("(%d/%d) uploaded : %s", i+1, len(filePaths), path)
	}
}

func execBatchUpload(cmd *cobra.Command, args []string) {
	var (
		Folders []BatchUploadFolder = []BatchUploadFolder{}
	)

	filepath.WalkDir(args[0], func(path string, d fs.DirEntry, err error) error {
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
			Folders = append(Folders, BatchUploadFolder{
				Name:   Name,
				Number: Number,
				Path:   path,
			})
		}
		return nil
	})

	fmt.Printf("%d folders are parsed\n", len(Folders))

	for _, f := range Folders {
		execUpload(cmd, []string{f.Path, fmt.Sprintf("%s(%d)", strings.Replace(f.Name, " ", "_", -1), f.Number)})
	}
}
