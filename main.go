package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

var (
	isDebug        bool
	client         = resty.New()
	Logger         = logrus.New()
	ImageExtension = []string{".jpg", ".jpeg", ".png", ".gif", ".webm", ".mp4"}
)

func Request() *resty.Request {
	return client.R().SetHeader("Accept", "application/json").SetHeader("Content-Type", "application/json")
}

func main() {
	var (
		userToken              string
		host, userId, userPass string
		directory, safety, tag string
		err                    error
	)

	flag.StringVar(&host, "host", "http://localhost", "address of host")
	flag.StringVar(&directory, "dir", "", "directory to upload")
	flag.StringVar(&safety, "safety", "unsafe", "safety of images in directory")
	flag.StringVar(&tag, "tag", "", "tag which will be assigned to images")
	flag.BoolVar(&isDebug, "debug", false, "print debug log")
	flag.Parse()

	if isDebug {
		Logger.SetLevel(logrus.DebugLevel)
	}

	fmt.Print("enter user id : ")
	fmt.Scanln(&userId)
	fmt.Print("enter user password : ")
	fmt.Scanln(&userPass)

	if userToken, err = login(host, userId, userId); err != nil {
		Logger.WithError(err).Fatalln("fail to log in")
	}

	filePaths := []string{}
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
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

	for _, path := range filePaths {
		ftok, err := uploadFile(host, userToken, path)
		if err != nil {
			Logger.WithError(err).Errorln("error caused during upload file")
			continue
		}

		rev, err := reverseSearch(host, userToken, ftok)
		if err != nil {
			Logger.WithError(err).Errorln("error caused during reverse search")
			continue
		}

		if err := createPost(host, userToken, ftok, tag, safety, rev); err != nil {
			Logger.WithError(err).Errorln("error caused during create post")
			continue
		}
		Logger.Infof("uploaded : %s", path)
	}
}

func login(host, userId, userPass string) (string, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", userId, userPass)))
	resp, err := Request().SetHeader("Authorization", "Basic "+auth).Post(host + "/api/user-token/" + userId)
	if err != nil {
		return "", err
	}
	ret := map[string]interface{}{}
	if err := json.Unmarshal(resp.Body(), &ret); err != nil {
		return "", err
	}
	if _, exist := ret["token"]; !exist {
		Logger.Debugf("response: %s\n", string(resp.Body()))
		return "", fmt.Errorf("request error: no token response")
	}
	return auth, nil
}

func uploadFile(host, userToken, filePath string) (string, error) {
	resp, err := Request().SetFile("content", filePath).SetHeader("Authorization", "Basic "+userToken).Post(host + "/api/uploads")
	if err != nil {
		return "", err
	}
	ret := map[string]string{}
	if err := json.Unmarshal(resp.Body(), &ret); err != nil {
		return "", err
	}
	if _, exist := ret["token"]; !exist {
		Logger.Debugf("response: %s\n", string(resp.Body()))
		return "", fmt.Errorf("request error: no token response")
	}
	return ret["token"], nil
}

func createPost(host, userToken, fileToken, tag, safety string, reverseSearch *ReverseSearchResponse) error {
	payload := map[string]interface{}{
		"contentToken": fileToken,
		"safety":       safety,
		"tags":         []string{tag},
	}
	if reverseSearch != nil && len(reverseSearch.SimilarPosts) > 0 {
		payload["relationCount"] = len(reverseSearch.SimilarPosts)
		similarPost := []int{}
		for _, p := range reverseSearch.SimilarPosts {
			similarPost = append(similarPost, p.Post.Id)
		}
		payload["relations"] = similarPost
	}
	m, _ := json.Marshal(payload)
	Logger.Debugln(string(m))
	resp, err := Request().SetHeader("Authorization", "Basic "+userToken).SetBody(payload).Post(host + "/api/posts")
	Logger.WithField("code", resp.StatusCode()).Debugf("response: %s\n", string(resp.Body()))
	if resp.StatusCode() != 200 {
		return fmt.Errorf("status code is %d", resp.StatusCode())
	}
	return err
}

func reverseSearch(host, userToken, fileToken string) (*ReverseSearchResponse, error) {
	resp, err := Request().SetHeader("Authorization", "Basic "+userToken).SetBody(map[string]string{"contentToken": fileToken}).Post(host + "/api/posts/reverse-search")
	if err != nil {
		return nil, err
	}
	Logger.WithField("code", resp.StatusCode()).Debugf("response: %s\n", string(resp.Body()))
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("status code is %d", resp.StatusCode())
	}
	result := ReverseSearchResponse{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
