package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

func Request() *resty.Request {
	return client.R().SetHeader("Accept", "application/json").SetHeader("Content-Type", "application/json")
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

func queryPost(host, userToken, query string, offset int) (*ListPostResponse, error) {
	const imagePerRequest = 50
	buildUrl := func(host, query string, limit, offset int) string {
		url := []string{
			host,
			"/api/posts/?query=",
			query,
			"&limit=",
			strconv.Itoa(imagePerRequest),
			"&offset=",
			strconv.Itoa(offset),
		}
		return strings.Join(url, "")
	}
	resp, err := Request().SetHeader("Authorization", "Basic "+userToken).Get(buildUrl(host, query, imagePerRequest, offset))
	if err != nil {
		return nil, err
	}
	logResponse(resp, "queryPost")
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("status code is %d (%s)", resp.StatusCode(), parseErrorResponse(resp))
	}
	result := ListPostResponse{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func deletePost(host, userToken string, post Post) error {
	resp, err := Request().SetHeader("Authorization", "Basic "+userToken).SetBody(map[string]int{"version": post.Version}).Delete(host + "/api/post/" + strconv.Itoa(post.Id))
	if err != nil {
		return err
	}
	logResponse(resp, "deletePost")
	if resp.StatusCode() != 200 {
		return fmt.Errorf("status code is %d (%s)", resp.StatusCode(), parseErrorResponse(resp))
	}
	return nil
}
