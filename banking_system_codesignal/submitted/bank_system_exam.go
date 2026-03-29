package bankingsystem_exam

import (
	"fmt"
	"sort"
)

type Account struct {
	id             string
	balance        int
	outgoingAmount int
}

type Payment struct {
	id        string
	sender    string
	amount    int
	timestamp int
	status    bool
}

type BankingSystemImpl struct {
	Accounts             map[string]*Account
	paymentTimestamps    map[int][]*Payment
	paymentIdToPayment   map[string]*Payment
	paymentOrdinalNumber int
}

func NewBankingSystemImpl() *BankingSystemImpl {
	BankingSystem := BankingSystemImpl{
		Accounts:             make(map[string]*Account),
		paymentTimestamps:    make(map[int][]*Payment),
		paymentIdToPayment:   make(map[string]*Payment),
		paymentOrdinalNumber: 1,
	}
	return &BankingSystem
}

func (b *BankingSystemImpl) CreateAccount(timestamp int, accountId string) bool {
	b.allSmallerTransac(timestamp)
	_, exists := b.Accounts[accountId]
	if exists {
		return false
	}
	newAccount := &Account{
		id:             accountId,
		balance:        0,
		outgoingAmount: 0,
	}
	b.Accounts[accountId] = newAccount
	return true
}

func (b *BankingSystemImpl) Deposit(timestamp int, accountId string, amountToAdd int) *int {
	b.allSmallerTransac(timestamp)
	account, exists := b.Accounts[accountId]
	if !exists {
		return nil
	}
	account.balance += amountToAdd
	return &account.balance
}

func (b *BankingSystemImpl) allSmallerTransac(timestamp int) {
	for i := 0; i <= timestamp; i++ {
		if len(b.paymentTimestamps) > 0 {
			b.PerformTransaction(i)
		}
	}
}

func (b *BankingSystemImpl) PerformTransaction(timestamp int) {
	for _, payment := range b.paymentTimestamps[timestamp] {
		if payment.status == true {
			continue
		}
		accountId := payment.sender
		account, exists := b.Accounts[accountId]
		if !exists {
			continue
		}
		if account.balance < payment.amount {
			continue
		}
		account.balance -= payment.amount
		account.outgoingAmount += payment.amount
		payment.status = true
	}
}

func (b *BankingSystemImpl) Transfer(timestamp int, sourceAccountId string, targetAccountId string, amountToTransfer int) *int {
	if sourceAccountId == targetAccountId {
		return nil
	}
	source, exists := b.Accounts[sourceAccountId]
	if !exists {
		return nil
	}
	target, exists := b.Accounts[targetAccountId]
	if !exists {
		return nil
	}
	// perform timestamp transactions first
	b.allSmallerTransac(timestamp)
	if source.balance < amountToTransfer {
		return nil
	}
	source.balance -= amountToTransfer
	source.outgoingAmount += amountToTransfer
	target.balance += amountToTransfer
	return &source.balance
}

type Pair struct {
	accountId      string
	outgoingAmount int
}

func (b *BankingSystemImpl) TopSpenders(timestamp int, n int) []string {
	b.allSmallerTransac(timestamp)
	topSpenderPairs := []*Pair{}
	for k, v := range b.Accounts {
		topSpenderPairs = append(topSpenderPairs, &Pair{
			accountId:      k,
			outgoingAmount: v.outgoingAmount,
		})
	}
	sort.Slice(topSpenderPairs, func(i int, j int) bool {
		if topSpenderPairs[i].outgoingAmount == topSpenderPairs[j].outgoingAmount {
			return topSpenderPairs[i].accountId < topSpenderPairs[j].accountId
		}
		return topSpenderPairs[i].outgoingAmount > topSpenderPairs[j].outgoingAmount
	})
	topSenders := []string{}
	for i := 0; i <= min(n-1, len(topSpenderPairs)-1); i++ {
		curPair := fmt.Sprintf("%s(%d)", topSpenderPairs[i].accountId, topSpenderPairs[i].outgoingAmount)
		topSenders = append(topSenders, curPair)
	}
	return topSenders
}

func (b *BankingSystemImpl) SchedulePayment(timestamp int, accountId string, amountToPay int, delay int) *string {
	b.allSmallerTransac(timestamp)
	account, exists := b.Accounts[accountId]
	if !exists {
		return nil
	}
	if account.balance < amountToPay {
		return nil
	}
	newPayment := &Payment{
		id:        fmt.Sprintf("payment%d", b.paymentOrdinalNumber),
		sender:    accountId,
		amount:    amountToPay,
		timestamp: timestamp + delay,
		status:    false,
	}
	b.paymentOrdinalNumber++
	b.paymentTimestamps[timestamp] = append(b.paymentTimestamps[timestamp], newPayment)
	b.paymentIdToPayment[newPayment.id] = newPayment
	return &newPayment.id
}

func (b *BankingSystemImpl) CancelPayment(timestamp int, accountId string, paymentId string) bool {
	b.allSmallerTransac(timestamp)
	payment, exists := b.paymentIdToPayment[paymentId]
	if !exists {
		return false
	}
	if payment.sender != accountId {
		return false
	}
	delete(b.paymentIdToPayment, paymentId)
	paymentsAtTimestamp := b.paymentTimestamps[payment.timestamp]
	index := 0
	for ; index < len(paymentsAtTimestamp); index++ {
		if paymentsAtTimestamp[index].id == paymentId {
			paymentsAtTimestamp[index].status = true
			break
		}
	}
	return true
}
