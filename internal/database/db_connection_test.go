package database

import (
	"testing"

	"github.com/sohWenMing/aggregator/internal/config"
	testutils "github.com/sohWenMing/aggregator/test_utils"
)

func TestCreateDBConnection(t *testing.T) {
	state, err := CreateDBConnection()
	testutils.AssertNoErr(err, t)
	compareConfig, err := config.Read()
	testutils.AssertNoErr(err, t)
	testutils.AssertStrings(state.Cfg.CurrentUserName, compareConfig.CurrentUserName, t)
}
