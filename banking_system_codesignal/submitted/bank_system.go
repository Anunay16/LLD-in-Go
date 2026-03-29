package bankingsystem

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
	senderId  string
	amount    int
	executeAt int
}

type BankingSystemImpl struct {
	accounts             map[string]*Account
	scheduledPayments    []*Payment
	paymentOrdinalNumber int
}

func NewBankingSystemImpl() *BankingSystemImpl {
	BankingSystem := BankingSystemImpl{
		accounts:             make(map[string]*Account),
		scheduledPayments:    []*Payment{},
		paymentOrdinalNumber: 1,
	}
	return &BankingSystem
}

func (b *BankingSystemImpl) CreateAccount(timestamp int, accountId string) bool {
	b.processPayments(timestamp)

	_, exists := b.accounts[accountId]
	if exists {
		return false
	}
	newAccount := &Account{
		id:             accountId,
		balance:        0,
		outgoingAmount: 0,
	}
	b.accounts[accountId] = newAccount
	return true
}

func (b *BankingSystemImpl) Deposit(timestamp int, accountId string, amountToAdd int) *int {
	b.processPayments(timestamp)

	account, exists := b.accounts[accountId]
	if !exists {
		return nil
	}
	account.balance += amountToAdd
	return &account.balance
}

func (b *BankingSystemImpl) Transfer(timestamp int, sourceAccountId string, targetAccountId string, amountToTransfer int) *int {
	b.processPayments(timestamp)

	if sourceAccountId == targetAccountId {
		return nil
	}
	source, exists := b.accounts[sourceAccountId]
	if !exists {
		return nil
	}
	target, exists := b.accounts[targetAccountId]
	if !exists {
		return nil
	}
	if source.balance < amountToTransfer {
		return nil
	}
	source.balance -= amountToTransfer
	source.outgoingAmount += amountToTransfer
	target.balance += amountToTransfer
	return &source.balance
}

type AccountSummay struct {
	accountId      string
	outgoingAmount int
}

func (b *BankingSystemImpl) TopSpenders(timestamp int, n int) []string {
	b.processPayments(timestamp)

	accountSummaries := []*AccountSummay{}
	for k, v := range b.accounts {
		accountSummaries = append(accountSummaries, &AccountSummay{
			accountId:      k,
			outgoingAmount: v.outgoingAmount,
		})
	}
	sort.Slice(accountSummaries, func(i, j int) bool {
		if accountSummaries[i].outgoingAmount == accountSummaries[j].outgoingAmount {
			return accountSummaries[i].accountId < accountSummaries[j].accountId
		}
		return accountSummaries[i].outgoingAmount > accountSummaries[j].outgoingAmount
	})
	topSenders := []string{}
	if len(accountSummaries) == 0 {
		return topSenders
	}
	for i := 0; i <= min(n-1, len(accountSummaries)-1); i++ {
		curAccountSummay := fmt.Sprintf("%s(%d)", accountSummaries[i].accountId, accountSummaries[i].outgoingAmount)
		topSenders = append(topSenders, curAccountSummay)
	}
	return topSenders
}

func (b *BankingSystemImpl) SchedulePayment(timestamp int, accountId string, amountToPay int, delay int) *string {
	b.processPayments(timestamp)

	_, exists := b.accounts[accountId]
	if !exists {
		return nil
	}

	newPayment := &Payment{
		id:        fmt.Sprintf("payment%d", b.paymentOrdinalNumber),
		senderId:  accountId,
		amount:    amountToPay,
		executeAt: timestamp + delay,
	}
	b.paymentOrdinalNumber++
	b.scheduledPayments = append(b.scheduledPayments, newPayment)
	return &newPayment.id
}

func (b *BankingSystemImpl) CancelPayment(timestamp int, accountId string, paymentId string) bool {
	b.processPayments(timestamp)

	for i, payment := range b.scheduledPayments {
		if payment.id == paymentId {
			if payment.senderId != accountId {
				return false
			}
			b.scheduledPayments = append(b.scheduledPayments[:i], b.scheduledPayments[i+1:]...)
			return true
		}
	}
	return false
}

func (b *BankingSystemImpl) processPayments(timestamp int) {
	if len(b.scheduledPayments) < 1 {
		return
	}
	newQueue := []*Payment{}
	for _, payment := range b.scheduledPayments {
		if payment.executeAt <= timestamp {
			account, exists := b.accounts[payment.senderId]
			if exists && account.balance >= payment.amount {
				account.balance -= payment.amount
				account.outgoingAmount += payment.amount
			}
		} else {
			newQueue = append(newQueue, payment)
		}
	}
	b.scheduledPayments = newQueue
}
