package wallet_test

import (
	"os"
	"testing"

	"github.com/buni/wallet/internal/pkg/testing/dt"
)

func TestMain(m *testing.M) {
	dt.SetupNATS()
	res := dt.SetupPostgres()

	code := m.Run()

	dt.Cleanup(res)
	os.Exit(code)
}
