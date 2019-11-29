package dana

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DanaSanguTestSuite struct {
	suite.Suite
}

func TestDanaSanguTestSuite(t *testing.T) {
	suite.Run(t, new(DanaSanguTestSuite))
}

func (dana *DanaSanguTestSuite) TestApplyTokenSuccess() {
	danaClient := NewClient()
	danaClient.BaseUrl = "https://api-sandbox.saas.dana.id"
	danaClient.Version = "2.0"
	danaClient.ClientId = "2019111272254582683145"
	danaClient.ClientSecret = "d83203415001474091383717b9e0572a"
	danaClient.LogLevel = 3
	danaClient.SignatureEnabled = true
	danaClient.PrivateKey = readKey("private-key.pem")
	danaClient.PublicKey = readKey("public-key.pem")

	coreGateway := CoreGateway{
		Client: danaClient,
	}

	reqBody := &RequestApplyAccessToken{
		GrantType:    "AUTHORIZATION_CODE",
		AuthCode:     "MpttSeDYJ8tLRJD5MdhQ9Hr7v2DZ8Uhs538f7300",
		RefreshToken: "",
	}

	resp, err := coreGateway.ApplyAccessToken(reqBody)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		fmt.Printf("response: %v\n", resp.Response.Body)
	}
}

func readKey(filePath string) []byte {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	return b
}
