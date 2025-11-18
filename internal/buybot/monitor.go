package buybot

import (
	"consul-telegram-bot/internal/logger"
	"fmt"
	"sync"
	"time"
)

const (
	PollInterval = 10 * time.Second
)

type BuyTransaction struct {
	Signature string
	Buyer     string
	Amount    float64
	SolAmount float64
	BlockTime int64
	TxURL     string
}

type Monitor struct {
	client           *HeliusClient
	logger           *logger.Logger
	tokenAddress     string
	lastSignature    string
	mu               sync.RWMutex
	onBuyTransaction func(*BuyTransaction)
	processedSigs    map[string]bool
	processedSigsMu  sync.RWMutex
	lastSentTime     time.Time
	lastSentAmount   float64
	throttleMu       sync.RWMutex
}

func NewMonitor(client *HeliusClient, logger *logger.Logger, tokenAddress string) *Monitor {
	return &Monitor{
		client:        client,
		logger:        logger,
		tokenAddress:  tokenAddress,
		processedSigs: make(map[string]bool),
	}
}

func (m *Monitor) SetBuyHandler(handler func(*BuyTransaction)) {
	m.onBuyTransaction = handler
}

func (m *Monitor) Start() {
	m.logger.Info("starting Aritect buy monitor for token: %s", m.tokenAddress)

	signatures, err := m.client.GetSignaturesForAddress(m.tokenAddress, 1)
	if err != nil {
		m.logger.Error("failed to get initial signatures: %s", err)
	} else if len(signatures) > 0 {
		m.mu.Lock()
		m.lastSignature = signatures[0].Signature
		m.mu.Unlock()
		m.logger.Info("set initial signature: %s", m.lastSignature)
	}

	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.checkNewTransactions()
	}
}

func (m *Monitor) checkNewTransactions() {
	signatures, err := m.client.GetSignaturesForAddress(m.tokenAddress, 10)
	if err != nil {
		m.logger.Error("failed to get signatures: %s", err)
		return
	}

	if len(signatures) == 0 {
		return
	}

	m.mu.RLock()
	lastSig := m.lastSignature
	m.mu.RUnlock()

	var newSignatures []SignatureInfo
	for _, sig := range signatures {
		if sig.Signature == lastSig {
			break
		}
		newSignatures = append(newSignatures, sig)
	}

	if len(newSignatures) > 0 {
		m.mu.Lock()
		m.lastSignature = signatures[0].Signature
		m.mu.Unlock()

		m.logger.Info("found %d new transactions", len(newSignatures))

		var buyTransactions []*BuyTransaction

		for i := len(newSignatures) - 1; i >= 0; i-- {
			sig := newSignatures[i]

			m.processedSigsMu.RLock()
			processed := m.processedSigs[sig.Signature]
			m.processedSigsMu.RUnlock()

			if processed {
				continue
			}

			if sig.Err != nil {
				m.processedSigsMu.Lock()
				m.processedSigs[sig.Signature] = true
				m.processedSigsMu.Unlock()
				continue
			}

			buyTx := m.processTransaction(sig)
			if buyTx != nil {
				buyTransactions = append(buyTransactions, buyTx)
			}

			m.processedSigsMu.Lock()
			m.processedSigs[sig.Signature] = true
			m.processedSigsMu.Unlock()
		}

		if len(buyTransactions) > 0 {
			largestBuy := m.findLargestBuy(buyTransactions)
			if largestBuy != nil && m.onBuyTransaction != nil {
				if m.shouldSendBuy(largestBuy) {
					m.logger.Info("sending largest buy from %d transactions: %.2f ARITECT", len(buyTransactions), largestBuy.Amount)
					m.onBuyTransaction(largestBuy)

					m.throttleMu.Lock()
					m.lastSentTime = time.Now()
					m.lastSentAmount = largestBuy.Amount
					m.throttleMu.Unlock()
				} else {
					m.logger.Info("skipping buy notification (throttled): %.2f ARITECT", largestBuy.Amount)
				}
			}
		}
	}

	m.processedSigsMu.Lock()
	if len(m.processedSigs) > 1000 {
		m.processedSigs = make(map[string]bool)
	}
	m.processedSigsMu.Unlock()
}

func (m *Monitor) processTransaction(sig SignatureInfo) *BuyTransaction {
	tx, err := m.client.GetTransaction(sig.Signature)
	if err != nil {
		m.logger.Error("failed to get transaction %s: %s", sig.Signature, err)
		return nil
	}

	if tx == nil || tx.Meta == nil {
		return nil
	}

	return m.analyzeBuyTransaction(tx, sig.Signature)
}

func (m *Monitor) findLargestBuy(buys []*BuyTransaction) *BuyTransaction {
	if len(buys) == 0 {
		return nil
	}

	largest := buys[0]
	for _, buy := range buys[1:] {
		if buy.Amount > largest.Amount {
			largest = buy
		}
	}

	return largest
}

func (m *Monitor) shouldSendBuy(buyTx *BuyTransaction) bool {
	m.throttleMu.RLock()
	defer m.throttleMu.RUnlock()

	if m.lastSentTime.IsZero() {
		return true
	}

	timeSinceLastSent := time.Since(m.lastSentTime)

	if timeSinceLastSent < time.Minute {
		if buyTx.Amount > m.lastSentAmount {
			m.logger.Info("new buy is larger (%.2f > %.2f), resetting timer", buyTx.Amount, m.lastSentAmount)
			return true
		}
		return false
	}

	return true
}

func (m *Monitor) analyzeBuyTransaction(tx *TransactionResponse, signature string) *BuyTransaction {
	if tx.Meta == nil || tx.Meta.Err != nil {
		return nil
	}

	var buyer string
	var tokenAmount float64
	var solAmount float64

	for _, postBalance := range tx.Meta.PostTokenBalances {
		if postBalance.Mint != m.tokenAddress {
			continue
		}

		var preAmount float64
		for _, preBalance := range tx.Meta.PreTokenBalances {
			if preBalance.AccountIndex == postBalance.AccountIndex {
				preAmount = preBalance.UiTokenAmount.UiAmount
				break
			}
		}

		postAmount := postBalance.UiTokenAmount.UiAmount

		if postAmount > preAmount {
			tokenAmount = postAmount - preAmount
			buyer = postBalance.Owner
			break
		}
	}

	if tokenAmount == 0 || buyer == "" {
		return nil
	}

	if len(tx.Meta.PreBalances) > 0 && len(tx.Meta.PostBalances) > 0 {
		preSol := float64(tx.Meta.PreBalances[0]) / 1e9
		postSol := float64(tx.Meta.PostBalances[0]) / 1e9
		solAmount = preSol - postSol

		fee := float64(tx.Meta.Fee) / 1e9
		solAmount -= fee

		if solAmount < 0 {
			solAmount = 0
		}
	}

	blockTime := int64(0)
	if tx.BlockTime != nil {
		blockTime = *tx.BlockTime
	}

	return &BuyTransaction{
		Signature: signature,
		Buyer:     buyer,
		Amount:    tokenAmount,
		SolAmount: solAmount,
		BlockTime: blockTime,
		TxURL:     fmt.Sprintf("https://solscan.io/tx/%s", signature),
	}
}
