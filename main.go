package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	flag.StringVar(&tag, "tag", "", "comma seperated tags which will be assigned to images")
	flag.StringVar(&userId, "uid", "", "user's login id")
	flag.StringVar(&userPass, "upw", "", "user's login password")
	flag.BoolVar(&isDebug, "debug", false, "print debug log")
	flag.Parse()

	if isDebug {
		Logger.SetLevel(logrus.DebugLevel)
	}
	Logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		DisableColors: false,
		ForceQuote:    false,
	})
	if len(userId) < 1 {
		fmt.Print("enter user id : ")
		fmt.Scanln(&userId)
	}
	if len(userPass) < 1 {
		fmt.Print("enter user password : ")
		fmt.Scanln(&userPass)
	}

	if userToken, err = login(host, userId, userPass); err != nil {
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

	logError := func(cur, total int, err error, path, action string) {
		Logger.WithFields(logrus.Fields{
			"error": err,
			"path":  path,
		}).Errorf("(%d/%d) error : %s\n", cur+1, total, action)
	}

	tags := strings.Split(tag, ",")
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
		if err := createPost(host, userToken, ftok, tags, safety, rev); err != nil {
			logError(i, len(filePaths), err, path, "create post")
			continue
		}
		Logger.Infof("(%d/%d) uploaded : %s", i+1, len(filePaths), path)
	}
}

func logResponse(resp *resty.Response, action string) {
	Logger.WithFields(logrus.Fields{
		"action": action,
		"code":   resp.StatusCode(),
	}).Debugf("response: %s\n", string(resp.Body()))
}

func parseErrorResponse(resp *resty.Response) string {
	var (
		name  = "unknown"
		title = "unknown"
		desc  = "unknown"
	)
	ret := map[string]interface{}{}
	if err := json.Unmarshal(resp.Body(), &ret); err != nil {
		return "fail to parse response"
	}
	if rawName, ok := ret["name"]; ok {
		name = rawName.(string)
	}
	if rawTitle, ok := ret["title"]; ok {
		title = rawTitle.(string)
	}
	if rawDesc, ok := ret["description"]; ok {
		desc = rawDesc.(string)
	}
	return fmt.Sprintf("<%s> %s : %s", name, title, desc)
}

func login(host, userId, userPass string) (string, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", userId, userPass)))
	resp, err := Request().SetHeader("Authorization", "Basic "+auth).Post(host + "/api/user-token/" + userId)
	if err != nil {
		return "", err
	}
	logResponse(resp, "login")
	ret := map[string]interface{}{}
	if err := json.Unmarshal(resp.Body(), &ret); err != nil {
		return "", err
	}
	if _, exist := ret["token"]; !exist {
		return "", fmt.Errorf("request error: no token response (%s)", parseErrorResponse(resp))
	}
	return auth, nil
}

func uploadFile(host, userToken, filePath string) (string, error) {
	resp, err := Request().SetFile("content", filePath).SetHeader("Authorization", "Basic "+userToken).Post(host + "/api/uploads")
	if err != nil {
		return "", err
	}
	logResponse(resp, "uploadFile")
	ret := map[string]string{}
	if err := json.Unmarshal(resp.Body(), &ret); err != nil {
		return "", err
	}
	if _, exist := ret["token"]; !exist {
		return "", fmt.Errorf("request error: no token response (%s)", parseErrorResponse(resp))
	}
	return ret["token"], nil
}

func createPost(host, userToken, fileToken string, tags []string, safety string, reverseSearch *ReverseSearchResponse) error {
	payload := map[string]interface{}{
		"contentToken": fileToken,
		"safety":       safety,
		"tags":         tags,
	}
	if reverseSearch != nil && len(reverseSearch.SimilarPosts) > 0 {
		payload["relationCount"] = len(reverseSearch.SimilarPosts)
		similarPost := []int{}
		for _, p := range reverseSearch.SimilarPosts {
			similarPost = append(similarPost, p.Post.Id)
		}
		payload["relations"] = similarPost
		Logger.WithFields(logrus.Fields{
			"count": payload["relationCount"],
			"posts": fmt.Sprint(payload["relations"]),
		}).Infof("file '%s' has similar posts and will apply relation between those posts.\n", fileToken)
	}
	m, _ := json.Marshal(payload)
	Logger.Debugln(string(m))
	resp, err := Request().SetHeader("Authorization", "Basic "+userToken).SetBody(payload).Post(host + "/api/posts")
	logResponse(resp, "createPost")
	if resp.StatusCode() != 200 {
		return fmt.Errorf("status code is %d (%s)", resp.StatusCode(), parseErrorResponse(resp))
	}
	return err
}

func reverseSearch(host, userToken, fileToken string) (*ReverseSearchResponse, error) {
	resp, err := Request().SetHeader("Authorization", "Basic "+userToken).SetBody(map[string]string{"contentToken": fileToken}).Post(host + "/api/posts/reverse-search")
	if err != nil {
		return nil, err
	}
	logResponse(resp, "createPost")
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("status code is %d (%s)", resp.StatusCode(), parseErrorResponse(resp))
	}
	result := ReverseSearchResponse{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
