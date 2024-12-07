package config

import (
	"testing"

	testHelpers "github.com/sohWenMing/aggregator/test_helpers"
)

func TestRead(t *testing.T) {

	want := "postgres://example"
	got, err := Read()
	testHelpers.AssertNoError(err, t)
	if got.Db_url != want {
		t.Errorf("\ngot: %s\nwant %s", got.Db_url, want)

	}
}

func TestSetUser(t *testing.T) {
	testUserName := "test user"
	initialConfig, err := Read()
	testHelpers.AssertNoError(err, t)

	copyConfig := initialConfig
	setUserErr := copyConfig.SetUser(testUserName)
	testHelpers.AssertNoError(setUserErr, t)

	configAfterWrite, err := Read()
	testHelpers.AssertNoError(err, t)
	testHelpers.AssertStrings(configAfterWrite.Db_url, copyConfig.Db_url, t)
	testHelpers.AssertStrings(configAfterWrite.Current_user_name, testUserName, t)

	initialConfig.SetUser("")
}
