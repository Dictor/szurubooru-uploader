package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

/*
args = [target directory, tags]
*/
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
		ptag := strings.Split(args[1], ",")
		if err := createPost(host, userToken, ftok, ptag, safety, rev); err != nil {
			logError(i, len(filePaths), err, path, "create post")
			continue
		}
		Logger.Infof("(%d/%d) uploaded : %s", i+1, len(filePaths), path)
	}
}

/*
args = [target directory]
*/
func execBatchUpload(cmd *cobra.Command, args []string) {
	var (
		Folders []BatchUploadFolder = []BatchUploadFolder{}
	)

	err := filepath.WalkDir(args[0], func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == d.Name() {
			return nil
		}
		var (
			Name   string
			Number int
		)
		if d.IsDir() {
			switch args[1] {
			case "pixiv":
				n, err := fmt.Sscanf(d.Name(), "%s (%d)", &Name, &Number)
				if n != 2 || err != nil {
					fmt.Printf("fail to parse path '%s'\n", path)
					return nil
				}
			case "name":
				n, err := fmt.Sscanf(d.Name(), "%s", &Name)
				Number = 0
				if n != 1 || err != nil {
					fmt.Printf("fail to parse path '%s'\n", path)
					return nil
				}
			case "split":
				Name = d.Name()
			default:
				return fmt.Errorf("unknown handler name: %s", args[1])
			}
			Folders = append(Folders, BatchUploadFolder{
				Name:   Name,
				Number: Number,
				Path:   path,
			})
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	fmt.Printf("%d folders are parsed\n", len(Folders))

	tag := ""
	for _, f := range Folders {
		switch args[1] {
		case "pixiv":
			tag = fmt.Sprintf("%s(%d)", strings.Replace(f.Name, " ", "_", -1), f.Number)
		case "name":
			tag = f.Name
		case "split":
			tag = strings.Replace(f.Name, " ", ",", -1)
		default:
			continue
		}
		Logger.WithFields(logrus.Fields{"path": f.Path, "tag": tag}).Infoln("upload folder")
		execUpload(cmd, []string{f.Path, tag})
	}
}

/*
args = [query, except favorite (bool)]
*/
func execDelete(cmd *cobra.Command, args []string) {
	logError := func(err error) {
		Logger.WithFields(logrus.Fields{
			"error": err,
			"query": args[0],
		}).Errorln("error caused during querying posts")
	}
	res, err := queryPost(host, userToken, args[0], 0)
	if err != nil {
		logError(err)
	}
	Logger.Infof("%d posts are found, start recursive posts retrieving\n", res.Total)

	posts := []Post{}
	posts = append(posts, res.Results...)
	currentPosition := len(res.Results)
	if res.Total > len(res.Results) {
		for {
			if currentPosition >= res.Total {
				break
			}
			res, err := queryPost(host, userToken, args[0], currentPosition)
			if err != nil {
				logError(err)
				return
			}
			posts = append(posts, res.Results...)
			currentPosition += len(res.Results)
		}
	}
	Logger.Infof("posts retrieving complete. %d posts are expected, %d posts are retrieved\n", res.Total, len(posts))
	fmt.Print("if want to continue, press enter (else, press ctrl + c)")
	fmt.Scanln()
	for i, p := range posts {
		if args[1] == "true" && p.FavoriteCount > 0 {
			Logger.Infof("(%d/%d) skipped : %d", i+1, len(posts), p.Id)
			continue
		}
		if err := deletePost(host, userToken, p); err != nil {
			Logger.WithFields(logrus.Fields{
				"error": err,
				"id":    p.Id,
			}).Errorf("(%d/%d) error : %d\n", i+1, len(posts), p.Id)
			continue
		}
		Logger.Infof("(%d/%d) deleted : %d", i+1, len(posts), p.Id)
	}
}

/*
args = [implication]
*/
func execBatchDelete(cmd *cobra.Command, args []string) {
	/* retrieving first page of tag  */
	logError := func(err error) {
		Logger.WithFields(logrus.Fields{
			"error": err,
			"query": args[0],
		}).Errorln("error caused during querying posts")
	}
	res, err := queryTag(host, userToken, "", 0)
	if err != nil {
		logError(err)
	}
	Logger.Infof("%d tags are found, start recursive tags retrieving\n", res.Total)

	/* retrieving whole tag pages */
	tags := []Tag{}
	tags = append(tags, res.Results...)
	currentPosition := len(res.Results)
	if res.Total > len(res.Results) {
		for {
			if currentPosition >= res.Total {
				break
			}
			res, err := queryTag(host, userToken, "", currentPosition)
			if err != nil {
				logError(err)
				return
			}
			tags = append(tags, res.Results...)
			currentPosition += len(res.Results)
		}
	}
	Logger.Infof("tags retrieving complete. %d tags are expected, %d tags are retrieved\n", res.Total, len(tags))

	/* filter tags by imlicate tag */
	implicateCandiate := []Tag{}
	for _, t := range tags {
		implications := lo.Map[Tag, string](t.Implications, func(implication Tag, index int) string {
			return implication.Names[0]
		})
		if lo.Contains(implications, args[0]) {
			implicateCandiate = append(implicateCandiate, t)
		}
	}
	Logger.Infof("%d sanitize tag candidate found", len(implicateCandiate))

	/* retrieving posts from candidate */
	for i, it := range implicateCandiate {
		/* exec delete command */
		Logger.Infof("------ batch (%d/%d) %s", i, len(implicateCandiate), it.Names[0])
		execDelete(nil, []string{it.Names[0], "true"})

		/* remove target tag and add "suspected" tag */
		mit := it
		hasTargetTag := false
		mit.Implications = lo.Filter[Tag](mit.Implications, func(item Tag, index int) bool {
			if item.Names[0] != args[0] {
				return true
			} else {
				hasTargetTag = true
				return false
			}
		})
		hasSuspected := lo.ContainsBy[Tag](mit.Implications, func(item Tag) bool {
			return item.Names[0] == "suspected"
		})
		if !hasSuspected {
			mit.Implications = append(mit.Implications, Tag{
				Names: []string{"suspected"},
			})
		}
		Logger.Infof("implication tags status: has '%s'? = %t / has 'suspected'? = %t", args[0], hasTargetTag, hasSuspected)
		req := ImplicationUpdateRequest{
			Version: mit.Version,
			Implications: lo.Map[Tag, string](mit.Implications, func(item Tag, index int) string {
				return item.Names[0]
			}),
		}
		if _, err = updateTagImplications(host, userToken, mit.Names[0], req); err != nil {
			Logger.WithError(err).Errorf("failed to update tag implications")
		}
	}
}
