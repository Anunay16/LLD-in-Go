package bankingsystem

import (
	"fmt"
	"sort"
)

type Account struct {
	Id             string
	Balance        int
	OutgoingAmount int
}

type Payment struct {
	Id        string
	senderId  string
	amount    int
	executeAt int
}

type BankingSystemImpl struct {
	Accounts             map[string]*Account
	scheduledPayments    []*Payment
	paymentOrdinalNumber int
}

func NewBankingSystemImpl() *BankingSystemImpl {
	BankingSystem := BankingSystemImpl{
		Accounts:             make(map[string]*Account),
		scheduledPayments:    []*Payment{},
		paymentOrdinalNumber: 1,
	}
	return &BankingSystem
}

func (b *BankingSystemImpl) CreateAccount(timestamp int, accountId string) bool {
	b.processPayments(timestamp)

	_, exists := b.Accounts[accountId]
	if exists {
		return false
	}
	newAccount := &Account{
		Id:             accountId,
		Balance:        0,
		OutgoingAmount: 0,
	}
	b.Accounts[accountId] = newAccount
	return true
}

func (b *BankingSystemImpl) Deposit(timestamp int, accountId string, amountToAdd int) *int {
	b.processPayments(timestamp)

	account, exists := b.Accounts[accountId]
	if !exists {
		return nil
	}
	account.Balance += amountToAdd
	return &account.Balance
}

func (b *BankingSystemImpl) Transfer(timestamp int, sourceAccountId string, targetAccountId string, amountToTransfer int) *int {
	b.processPayments(timestamp)

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
	if source.Balance < amountToTransfer {
		return nil
	}
	source.Balance -= amountToTransfer
	source.OutgoingAmount += amountToTransfer
	target.Balance += amountToTransfer
	return &source.Balance
}

type AccountSummay struct {
	accountId      string
	OutgoingAmount int
}

func (b *BankingSystemImpl) TopSpenders(timestamp int, n int) []string {
	b.processPayments(timestamp)

	accountSummaries := []*AccountSummay{}
	for k, v := range b.Accounts {
		accountSummaries = append(accountSummaries, &AccountSummay{
			accountId:      k,
			OutgoingAmount: v.OutgoingAmount,
		})
	}
	sort.Slice(accountSummaries, func(i, j int) bool {
		if accountSummaries[i].OutgoingAmount == accountSummaries[j].OutgoingAmount {
			return accountSummaries[i].accountId < accountSummaries[j].accountId
		}
		return accountSummaries[i].OutgoingAmount > accountSummaries[j].OutgoingAmount
	})
	topSenders := []string{}
	if len(accountSummaries) == 0 {
		return topSenders
	}
	for i := 0; i <= min(n-1, len(accountSummaries)-1); i++ {
		curAccountSummay := fmt.Sprintf("%s(%d)", accountSummaries[i].accountId, accountSummaries[i].OutgoingAmount)
		topSenders = append(topSenders, curAccountSummay)
	}
	return topSenders
}

func (b *BankingSystemImpl) SchedulePayment(timestamp int, accountId string, amountToPay int, delay int) *string {
	b.processPayments(timestamp)

	_, exists := b.Accounts[accountId]
	if !exists {
		return nil
	}

	newPayment := &Payment{
		Id:        fmt.Sprintf("payment%d", b.paymentOrdinalNumber),
		senderId:  accountId,
		amount:    amountToPay,
		executeAt: timestamp + delay,
	}
	b.paymentOrdinalNumber++
	b.scheduledPayments = append(b.scheduledPayments, newPayment)
	return &newPayment.Id
}

func (b *BankingSystemImpl) CancelPayment(timestamp int, accountId string, paymentId string) bool {
	b.processPayments(timestamp)

	for i, payment := range b.scheduledPayments {
		if payment.Id == paymentId {
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
			account, exists := b.Accounts[payment.senderId]
			if exists && account.Balance >= payment.amount {
				account.Balance -= payment.amount
				account.OutgoingAmount += payment.amount
			}
		} else {
			newQueue = append(newQueue, payment)
		}
	}
	b.scheduledPayments = newQueue
}
