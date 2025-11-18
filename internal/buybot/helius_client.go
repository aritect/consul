package buybot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HeliusClient struct {
	rpcURL     string
	httpClient *http.Client
}

type RPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RPCResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type SignatureInfo struct {
	Signature string `json:"signature"`
	Slot      uint64 `json:"slot"`
	Err       interface{} `json:"err"`
	Memo      *string `json:"memo"`
	BlockTime *int64  `json:"blockTime"`
}

type TransactionResponse struct {
	Slot        uint64                 `json:"slot"`
	Transaction TransactionData        `json:"transaction"`
	Meta        *TransactionMeta       `json:"meta"`
	BlockTime   *int64                 `json:"blockTime"`
}

type TransactionData struct {
	Message    MessageData `json:"message"`
	Signatures []string    `json:"signatures"`
}

type MessageData struct {
	AccountKeys     []string      `json:"accountKeys"`
	Instructions    []Instruction `json:"instructions"`
	RecentBlockhash string        `json:"recentBlockhash"`
}

type Instruction struct {
	ProgramIdIndex int    `json:"programIdIndex"`
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
}

type TransactionMeta struct {
	Err               interface{}       `json:"err"`
	Fee               uint64            `json:"fee"`
	PreBalances       []uint64          `json:"preBalances"`
	PostBalances      []uint64          `json:"postBalances"`
	PreTokenBalances  []TokenBalance    `json:"preTokenBalances"`
	PostTokenBalances []TokenBalance    `json:"postTokenBalances"`
	LogMessages       []string          `json:"logMessages"`
}

type TokenBalance struct {
	AccountIndex  int         `json:"accountIndex"`
	Mint          string      `json:"mint"`
	Owner         string      `json:"owner"`
	ProgramId     string      `json:"programId"`
	UiTokenAmount TokenAmount `json:"uiTokenAmount"`
}

type TokenAmount struct {
	Amount         string  `json:"amount"`
	Decimals       int     `json:"decimals"`
	UiAmount       float64 `json:"uiAmount"`
	UiAmountString string  `json:"uiAmountString"`
}

func NewHeliusClient(rpcURL string) *HeliusClient {
	return &HeliusClient{
		rpcURL: rpcURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *HeliusClient) call(method string, params []interface{}) (json.RawMessage, error) {
	reqBody := RPCRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.rpcURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

func (c *HeliusClient) GetSignaturesForAddress(address string, limit int) ([]SignatureInfo, error) {
	params := []interface{}{
		address,
		map[string]interface{}{
			"limit": limit,
		},
	}

	result, err := c.call("getSignaturesForAddress", params)
	if err != nil {
		return nil, err
	}

	var signatures []SignatureInfo
	if err := json.Unmarshal(result, &signatures); err != nil {
		return nil, fmt.Errorf("failed to unmarshal signatures: %w", err)
	}

	return signatures, nil
}

func (c *HeliusClient) GetTransaction(signature string) (*TransactionResponse, error) {
	params := []interface{}{
		signature,
		map[string]interface{}{
			"encoding":                       "json",
			"maxSupportedTransactionVersion": 0,
		},
	}

	result, err := c.call("getTransaction", params)
	if err != nil {
		return nil, err
	}

	var tx TransactionResponse
	if err := json.Unmarshal(result, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &tx, nil
}
