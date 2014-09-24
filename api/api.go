package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

func Me(token string) (accountId, mAccountId, onlineId string, err error) {
	host := "https://vl.api.np.km.playstation.net"
	url := host + "/vl/api/v1/mobile/users/me/info"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "com.playstation.companionutil.USER_AGENT")
	req.Header.Set("X-NP-ACCESS-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var data map[string]interface{}
	if err = json.Unmarshal(body, &data); err != nil {
		return
	}

	accountId = data["accountId"].(string)
	mAccountId = data["mAccountId"].(string)
	onlineId = data["onlineId"].(string)

	return
}

func AddFriend(token, me, target, message string) error {
	host := "https://se-prof.np.community.playstation.net"
	params := url.Values{}
	params.Set("requestMessage", message)
	url := host + "/userProfile/v1/users/" + me + "/friendList/" + target + "?" + params.Encode()

	data := `{"requestMessage":"` + message + `"}`
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(data)))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "com.playstation.companionutil.USER_AGENT")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Content-Length", strconv.Itoa(len(data)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Requested-with", "com.scee.psxandroid")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var data map[string]interface{}
		if err = json.Unmarshal(body, &data); err != nil {
			return err
		}

		return errors.New(data["error"].(map[string]interface{})["message"].(string))
	}

	return nil
}
