package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strconv"
	"time"

	"github.com/howeyc/gopass"

	"github.com/x1a0/gosen/api"
	"github.com/x1a0/gosen/auth"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println(Usage("gosen"))
		os.Exit(0)
	}

	if debug, err := strconv.ParseBool(os.Getenv("DEBUG")); err == nil && debug {
		proxyUrl, _ := url.Parse("http://127.0.0.1:8888")
		http.DefaultTransport = &http.Transport{
			Proxy:           http.ProxyURL(proxyUrl),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	switch cmd := os.Args[1]; cmd {
	case "login":
		token, err := login()
		check(err)

		accountId, mAccountId, onlineId, err := api.Me(token.Access_token)
		check(err)

		account := &Account{
			accountId,
			mAccountId,
			onlineId,
			token.Access_token,
			token.Refresh_token,
			time.Now().Unix() + int64(token.Expires_in)}

		check(saveAccount(account))

	case "friend":
		if len(os.Args) < 3 {
			fmt.Println(Usage("friend"))
		}

		account, err := loadAccount()
		check(err)

		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Enter Message (optional):")
		message, err := reader.ReadString('\n')
		check(err)
		message = TrimInput(message)

		fmt.Print("Sending friend request to: ")
		fmt.Println(os.Args[2:])
		if message == "" {
			fmt.Println("Without message")
		} else {
			fmt.Println("With message:")
			fmt.Println(message)
		}
		fmt.Print("OK? (y/n):")

		if Confirm() {
			for _, target := range os.Args[2:] {
				if err := api.AddFriend(account.AccessToken, account.OnlineId, target, message); err != nil {
					log.Println(fmt.Sprintf("Cannot send friend request to '%s': %s", target, err))
				} else {
					log.Println(fmt.Sprintf("A friend request has been sent to '%s'", target))
				}
			}
		}

	case "refresh":
		check(refresh(false))

	default:
		fmt.Println(Usage("gosen"))
		os.Exit(1)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

type Account struct {
	AccountId    string
	MAccountId   string
	OnlineId     string
	AccessToken  string
	RefreshToken string
	Expired      int64
}

func login() (*auth.AccessToken, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Account: ")
	account, _ := reader.ReadString('\n')
	fmt.Printf("Password: ")
	pass := gopass.GetPasswd()

	token, err := auth.Authenticate(
		TrimInput(account),
		TrimInput(string(pass)))

	if err != nil {
		return nil, err
	}

	return token, nil
}

func saveAccount(account *Account) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}

	data, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(usr.HomeDir+"/.gosen", data, 0600)
}

func loadAccount() (*Account, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(usr.HomeDir + "/.gosen")
	if err != nil {
		return nil, err
	}

	account := &Account{}
	if err = json.Unmarshal(data, account); err != nil {
		return nil, err
	}

	return account, nil
}

func refresh(force bool) error {
	account, err := loadAccount()
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	if !force && account.Expired > now+(60*2) {
		return errors.New("Current access token has not expired")
	}

	newToken, err := auth.Refresh(account.RefreshToken)
	if err != nil {
		return err
	}

	account.AccessToken = newToken.Access_token
	account.RefreshToken = newToken.Refresh_token
	account.Expired = now + int64(newToken.Expires_in)

	return saveAccount(account)
}
