package dana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	uuid "github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

const (
	//ORDER_PATH       = "https://api-sandbox.saas.dana.id/alipayplus/acquiring/order/createOrder.htm"
	ORDER_PATH       = "alipayplus/acquiring/order/createOrder.htm"
	QUERY_PATH       = "alipayplus/acquiring/order/query.htm"
	DANA_TIME_LAYOUT = "2006-01-02T15:04:05.000-07:00"
	CURRENCY_IDR     = "IDR"
)

// CoreGateway struct
type CoreGateway struct {
	Client Client
}

// Call : base method to call Core API
func (gateway *CoreGateway) Call(method, path string, header map[string]string, body io.Reader, v interface{}) error {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	path = gateway.Client.BaseUrl + path

	return gateway.Client.Call(method, path, header, body, v)
}

func (gateway *CoreGateway) Order(reqBody *OrderRequestData) (res ResponseBody, err error) {
	reqBody.Order.OrderAmount.Value = fmt.Sprintf("%v00", reqBody.Order.OrderAmount.Value)

	res, err = gateway.requestToDana(reqBody)
	if err != nil {
		return
	}

	var orderResponseData OrderResponseData
	err = mapstructure.Decode(res.Response.Body, &orderResponseData)
	if err != nil {
		return
	}

	res.Response.Body = orderResponseData

	//response RSA verification
	err = verifySignature(res.Response, res.Signature, gateway.Client.PublicKey)
	if err != nil {
		err = fmt.Errorf("could not verify request: %v", err)
	}

	return
}

func (gateway *CoreGateway) OrderDetail(reqBody *OrderDetailRequestData) (res ResponseBody, err error) {
	res, err = gateway.requestToDana(reqBody)
	if err != nil {
		return
	}

	var orderDetailData OrderDetailData
	err = mapstructure.Decode(res.Response.Body, &orderDetailData)
	if err != nil {
		return
	}

	res.Response.Body = orderDetailData

	err = verifySignature(res.Response, res.Signature, gateway.Client.PublicKey)
	if err != nil {
		err = fmt.Errorf("could not verify request: %v", err)
	}

	return
}

func (gateway *CoreGateway) GenerateSignature(req interface{}) (signature string, err error) {
	signature, err = generateSignature(req, gateway.Client.PrivateKey)
	if err != nil {
		err = fmt.Errorf("failed to generate signature: %v", err)
		return
	}

	return
}

func (gateway *CoreGateway) VerifySignature(res interface{}, signature string) (err error) {
	err = verifySignature(res, signature, gateway.Client.PublicKey)
	if err != nil {
		err = fmt.Errorf("could not verify request: %v", err)
	}
	return
}

func (gateway *CoreGateway) requestToDana(reqBody interface{}) (res ResponseBody, err error) {
	now := time.Now()

	head := RequestHeader{}
	head.Version = gateway.Client.Version
	head.Function = gateway.Client.Function
	head.ClientID = gateway.Client.ClientId
	head.ReqTime = now.Format(DANA_TIME_LAYOUT)
	head.ClientSecret = gateway.Client.ClientSecret

	var id uuid.UUID
	id, err = uuid.NewUUID()
	if err != nil {
		return res, err
	}

	head.ReqMsgID = id.String()

	req := Request{
		Head: head,
		Body: reqBody,
	}

	sig, err := generateSignature(req, gateway.Client.PrivateKey)
	if err != nil {
		err = fmt.Errorf("failed to generate signature: %v", err)
		return
	}
	orderDetailReq := RequestBody{
		Request:   req,
		Signature: sig,
	}

	reqJson, err := json.Marshal(orderDetailReq)
	if err != nil {
		return
	}

	log.Println("Dana request: ", string(reqJson))
	requestReader := bytes.NewBuffer(reqJson)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	err = gateway.Call("POST", ORDER_PATH, headers, requestReader, &res)
	if err != nil {
		return
	}

	return
}
