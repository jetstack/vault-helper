package okta

import (
	"os"
	"strings"
	"testing"
)

var client *Client

func init() {
	client = NewClient(os.Getenv("OKTA_ORG"))
	client.Url = "oktapreview.com"
}

func TestAPIFailure(t *testing.T) {
	client := NewClient("organization")
	_, err := client.Authenticate("username", "password")
	if !strings.Contains(err.Error(), "E0000007") {
		t.Error("Expected E0000007, got ", err.Error())
	}
}

func TestAcceptance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	_, err := client.Authenticate(os.Getenv("OKTA_USERNAME"), os.Getenv("OKTA_PASSWORD"))
	if err != nil {
		t.Error("Expected nil, got ", err.Error())
	}
}

func TestNoTokenUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	_, err := client.User(os.Getenv("OKTA_USERNAME"))
	if !strings.Contains(err.Error(), "E0000005") {
		t.Error("Expected E0000005, got ", err.Error())
	}
}

func TestUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	client.ApiToken = os.Getenv("OKTA_API_TOKEN")
	_, err := client.User(os.Getenv("OKTA_USERNAME"))
	if err != nil {
		t.Error("Expected nil, got ", err.Error())
	}
}

func TestUserGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	client.ApiToken = os.Getenv("OKTA_API_TOKEN")
	groups, err := client.Groups(os.Getenv("OKTA_USERNAME"))
	if err != nil {
		t.Error("Expected nil, got ", err.Error())
	} else if len(*groups) != 2 {
		t.Error("Expected length of 2, got ", len(*groups))
	}

}
