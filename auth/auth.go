package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.net/publicsuffix"
)

const (
	REMOTE_HOST      = "https://auth.api.sonyentertainmentnetwork.com"
	USER_AGENT       = "Mozilla/5.0 (Linux; Android 4.4.4; Nexus 5 Build/KTU84P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2062.117 Mobile Safari/537.36"
	OAUTH_APP_ID     = "b0d0d7ad-bb99-4ab1-b25e-afa0c76577b0"
	OAUTH_APP_SECRET = "Zo4y8eGIa3oazIEp"
	OAUTH_REDIRECT   = "com.scee.psxandroid.scecompcall://redirect"
	OAUTH_SCOPE      = "psn:sceapp"
	DUID             = "00000007000201283335323133363036383337303233343a4c4745202020202020203a68616d6d657268656164"
)

type AccessToken struct {
	Access_token  string
	Token_type    string
	Refresh_token string
	Expires_in    uint32
	Scope         string
}

func touch(client *http.Client) error {
	path := "/2.0/oauth/authorize"

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("service_entity", "urn:service-entity:psn")
	params.Set("returnAuthCode", "true")
	params.Set("cltm", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("redirect_uri", OAUTH_REDIRECT)
	params.Set("client_id", OAUTH_APP_ID)
	params.Set("scope", OAUTH_SCOPE)

	req, err := http.NewRequest("GET", REMOTE_HOST+path+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", USER_AGENT)

	_, err = client.Do(req)
	return err
}

func authorization(client *http.Client, username, password string) (string, error) {
	path := "/login.do"

	data := url.Values{}
	data.Set("j_username", username)
	data.Set("j_password", password)

	req, err := http.NewRequest("POST", REMOTE_HOST+path, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", USER_AGENT)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(req)
	if err != nil && !strings.HasSuffix(err.Error(), "__expected__") {
		return "", err
	}

	if code, ok := resp.Header["X-Np-Grant-Code"]; ok {
		return code[0], nil
	}

	return "", errors.New("cannot get grant code")
}

func auth(client *http.Client, authorizationCode string) (*AccessToken, error) {
	path := "/2.0/oauth/token"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", OAUTH_APP_ID)
	data.Set("client_secret", OAUTH_APP_SECRET)
	data.Set("code", authorizationCode)
	data.Set("redirect_uri", OAUTH_REDIRECT)
	data.Set("state", "x")
	data.Set("scope", OAUTH_SCOPE)
	data.Set("duid", DUID)

	req, err := http.NewRequest("POST", REMOTE_HOST+path, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "com.playstation.companionutil.USER_AGENT")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	token := &AccessToken{}
	if err = json.Unmarshal(body, token); err != nil {
		return nil, err
	}

	return token, nil
}

func redirectPolicyFunc(req *http.Request, via []*http.Request) error {
	if strings.HasPrefix(req.URL.Path, "/mobile-success.jsp") {
		return errors.New("__expected__")
	}
	return nil
}

func Authenticate(username, password string) (*AccessToken, error) {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar:           jar,
		CheckRedirect: redirectPolicyFunc,
	}

	if err = touch(client); err != nil {
		return nil, err
	}

	authCode, err := authorization(client, username, password)
	if err != nil {
		return nil, err
	}

	accessToken, err := auth(client, authCode)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func Refresh(token string) (*AccessToken, error) {
	path := "/2.0/oauth/token"

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", OAUTH_APP_ID)
	data.Set("client_secret", OAUTH_APP_SECRET)
	data.Set("refresh_token", token)
	data.Set("scope", OAUTH_SCOPE)
	data.Set("duid", DUID)

	req, err := http.NewRequest("POST", REMOTE_HOST+path, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "com.playstation.companionutil.USER_AGENT")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	newtoken := &AccessToken{}

	if err = json.Unmarshal(body, newtoken); err != nil {
		return nil, err
	}

	return newtoken, nil
}
