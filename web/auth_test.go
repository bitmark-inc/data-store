package web

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	bitmarksdk "github.com/bitmark-inc/bitmark-sdk-go"
	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/macaroon.v2"
)

func TestRegister(t *testing.T) {
	sdk.Init(&sdk.Config{
		Network:    bitmarksdk.Testnet,
		APIToken:   viper.GetString("bitmarksdk.token"),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	})

	clientAccount, _ := account.FromSeed("9J87EKVYuxzdCuo7QA7fcLL8kKkiBXtpN")
	serverAccount, _ := account.FromSeed("9J87Ga31xgbhPMqmRucMavUkv3zToPdBr")

	// set up server
	s := NewServer(nil, serverAccount.(*account.AccountV2), "localhost", []byte("ROOT KEY"))
	r := gin.Default()
	r.POST("/register", s.Register)
	testServer := httptest.NewServer(r)
	defer testServer.Close()

	// send registration request
	ts := fmt.Sprintf("%d", time.Now().UTC().Unix())
	sig := hex.EncodeToString(clientAccount.Sign([]byte(ts)))
	pubkey := hex.EncodeToString(clientAccount.(*account.AccountV2).EncrKey.PublicKeyBytes())
	reqBody, _ := json.Marshal(map[string]string{
		"requester":             clientAccount.AccountNumber(),
		"timestamp":             ts,
		"signature":             sig,
		"encryption_public_key": pubkey,
	})
	resp, err := http.Post(fmt.Sprintf("%s/register", testServer.URL), "application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// decrypt macaroons
	var respBody struct {
		R string `json:"r"`
		W string `json:"w"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.NoError(t, checkMacaroon(clientAccount.(*account.AccountV2), respBody.R, serverAccount.(*account.AccountV2).EncrKey.PublicKeyBytes()))
	assert.NoError(t, checkMacaroon(clientAccount.(*account.AccountV2), respBody.W, serverAccount.(*account.AccountV2).EncrKey.PublicKeyBytes()))
}

func checkMacaroon(clientAccount *account.AccountV2, encryptedMacaroon string, serverPublicKeys []byte) error {
	ciphertext, err := hex.DecodeString(encryptedMacaroon)
	if err != nil {
		return err
	}

	macaroonBytes, err := clientAccount.EncrKey.Decrypt(ciphertext, serverPublicKeys)
	if err != nil {
		return err
	}

	var m macaroon.Macaroon
	if err := m.UnmarshalBinary(macaroonBytes); err != nil {
		return err
	}

	return nil
}
