package bankingsystem

import (
	"reflect"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	b := NewBankingSystemImpl()

	// Create new account
	if !b.CreateAccount(1, "A") {
		t.Errorf("expected account creation to succeed")
	}

	// Duplicate account
	if b.CreateAccount(2, "A") {
		t.Errorf("expected duplicate account creation to fail")
	}
}

func TestDeposit(t *testing.T) {
	b := NewBankingSystemImpl()

	// Deposit to non-existing account
	if res := b.Deposit(1, "A", 100); res != nil {
		t.Errorf("expected nil for non-existing account")
	}

	b.CreateAccount(1, "A")

	// First deposit
	res := b.Deposit(2, "A", 100)
	if res == nil || *res != 100 {
		t.Errorf("expected balance 100, got %v", res)
	}

	// Second deposit
	res = b.Deposit(3, "A", 50)
	if res == nil || *res != 150 {
		t.Errorf("expected balance 150, got %v", res)
	}
}

func TestTransfer(t *testing.T) {
	b := NewBankingSystemImpl()

	b.CreateAccount(1, "A")
	b.CreateAccount(1, "B")

	b.Deposit(2, "A", 200)

	// Valid transfer
	res := b.Transfer(3, "A", "B", 100)
	if res == nil || *res != 100 {
		t.Errorf("expected source balance 100, got %v", res)
	}

	// Check balances
	if b.accounts["B"].balance != 100 {
		t.Errorf("expected target balance 100, got %d", b.accounts["B"].balance)
	}

	// Insufficient funds
	res = b.Transfer(4, "A", "B", 200)
	if res != nil {
		t.Errorf("expected nil due to insufficient funds")
	}

	// Same account transfer
	res = b.Transfer(5, "A", "A", 10)
	if res != nil {
		t.Errorf("expected nil for same account transfer")
	}

	// Non-existing account
	res = b.Transfer(6, "A", "C", 10)
	if res != nil {
		t.Errorf("expected nil for non-existing account")
	}
}

func TestTopSpenders(t *testing.T) {
	b := NewBankingSystemImpl()

	b.CreateAccount(1, "A")
	b.CreateAccount(1, "B")
	b.CreateAccount(1, "C")

	b.Deposit(2, "A", 500)
	b.Deposit(2, "B", 300)
	b.Deposit(2, "C", 200)

	// Transfers to build outgoing amounts
	b.Transfer(3, "A", "B", 200) // A: 200
	b.Transfer(4, "B", "C", 100) // B: 100
	b.Transfer(5, "A", "C", 100) // A: 300 total

	// Expected ranking:
	// A(300), B(100), C(0)
	expected := []string{"A(300)", "B(100)", "C(0)"}

	result := b.TopSpenders(6, 3)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// Test top 2 only
	expectedTop2 := []string{"A(300)", "B(100)"}
	result = b.TopSpenders(7, 2)

	if !reflect.DeepEqual(result, expectedTop2) {
		t.Errorf("expected %v, got %v", expectedTop2, result)
	}
}

func TestTopSpenders_Tie(t *testing.T) {
	b := NewBankingSystemImpl()

	b.CreateAccount(1, "A")
	b.CreateAccount(1, "B")

	b.Deposit(2, "A", 200)
	b.Deposit(2, "B", 200)

	// Both spend same amount
	b.Transfer(3, "A", "B", 100) // A: 100
	b.Transfer(4, "B", "A", 100) // B: 100

	// Tie → alphabetical order
	expected := []string{"A(100)", "B(100)"}

	result := b.TopSpenders(5, 2)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSchedulePayment(t *testing.T) {
	t.Run("Test_SchedulePayment_Order", func(t *testing.T) {
		b := NewBankingSystemImpl()

		b.CreateAccount(1, "A")
		b.Deposit(2, "A", 100)

		b.SchedulePayment(3, "A", 60, 2) // payment1 at 5
		b.SchedulePayment(3, "A", 50, 2) // payment2 at 5

		b.Deposit(5, "A", 0)

		// payment1 succeeds → balance 40
		// payment2 fails (insufficient funds)

		if b.accounts["A"].balance != 40 {
			t.Fatalf("expected 40, got %d", b.accounts["A"].balance)
		}
	})

	t.Run("Test_SchedulePayment_BasicExecution", func(t *testing.T) {
		b := NewBankingSystemImpl()

		b.CreateAccount(1, "A")
		b.Deposit(2, "A", 100)

		pid := b.SchedulePayment(3, "A", 50, 2) // executes at 5
		if pid == nil {
			t.Fatalf("expected payment id")
		}

		// Trigger execution
		b.Deposit(5, "A", 0)

		if b.accounts["A"].balance != 50 {
			t.Fatalf("expected balance 50, got %d", b.accounts["A"].balance)
		}
	})

	t.Run("Test_SchedulePayment_InsufficientFunds", func(t *testing.T) {
		b := NewBankingSystemImpl()

		b.CreateAccount(1, "A")
		b.Deposit(2, "A", 30)

		b.SchedulePayment(3, "A", 50, 2) // executes at 5

		b.Deposit(5, "A", 0)

		if b.accounts["A"].balance != 30 {
			t.Fatalf("payment should be skipped")
		}
	})

	t.Run("Test_CancelPayment_Success", func(t *testing.T) {
		b := NewBankingSystemImpl()

		b.CreateAccount(1, "A")
		b.Deposit(2, "A", 100)

		pid := b.SchedulePayment(3, "A", 50, 2)

		ok := b.CancelPayment(4, "A", *pid)
		if !ok {
			t.Fatalf("expected cancel success")
		}

		b.Deposit(5, "A", 0)

		if b.accounts["A"].balance != 100 {
			t.Fatalf("payment should not execute")
		}
	})

	t.Run("Test_CancelPayment_AfterExecution", func(t *testing.T) {
		b := NewBankingSystemImpl()

		b.CreateAccount(1, "A")
		b.Deposit(2, "A", 100)

		pid := b.SchedulePayment(3, "A", 50, 2) // executes at 5

		// At timestamp 5 → payment executes first
		ok := b.CancelPayment(5, "A", *pid)

		if ok {
			t.Fatalf("cancel should fail after execution")
		}

		if b.accounts["A"].balance != 50 {
			t.Fatalf("payment should already be deducted")
		}
	})

	t.Run("Test_CancelPayment_WrongAccount", func(t *testing.T) {
		b := NewBankingSystemImpl()

		b.CreateAccount(1, "A")
		b.CreateAccount(1, "B")

		b.Deposit(2, "A", 100)

		pid := b.SchedulePayment(3, "A", 50, 2)

		ok := b.CancelPayment(4, "B", *pid)
		if ok {
			t.Fatalf("should fail due to wrong account")
		}
	})

	t.Run("Test_CancelPayment_NotFound", func(t *testing.T) {
		b := NewBankingSystemImpl()

		b.CreateAccount(1, "A")

		ok := b.CancelPayment(2, "A", "payment999")
		if ok {
			t.Fatalf("should fail for non-existent payment")
		}
	})

	t.Run("Test_TopSpenders_WithScheduledPayments", func(t *testing.T) {
		b := NewBankingSystemImpl()

		b.CreateAccount(1, "A")
		b.CreateAccount(1, "B")

		b.Deposit(2, "A", 100)
		b.Deposit(2, "B", 100)

		b.SchedulePayment(3, "A", 70, 2) // executes at 5
		b.SchedulePayment(3, "B", 50, 2)

		b.Deposit(5, "A", 0)

		res := b.TopSpenders(6, 2)

		expected := []string{"A(70)", "B(50)"}

		for i := range res {
			if res[i] != expected[i] {
				t.Fatalf("unexpected result: %v", res)
			}
		}
	})

	t.Run("Test_Order_ScheduledBeforeTransfer", func(t *testing.T) {
		b := NewBankingSystemImpl()

		b.CreateAccount(1, "A")
		b.CreateAccount(1, "B")

		b.Deposit(2, "A", 100)

		b.SchedulePayment(3, "A", 80, 2) // executes at 5

		// At timestamp 5:
		// payment executes first → balance = 20
		res := b.Transfer(5, "A", "B", 50)

		if res != nil {
			t.Fatalf("transfer should fail due to insufficient funds")
		}
	})
}

func Test_Complex_Interleaving(t *testing.T) {
	b := NewBankingSystemImpl()

	b.CreateAccount(1, "A")
	b.CreateAccount(1, "B")

	b.Deposit(2, "A", 200)

	b.SchedulePayment(3, "A", 100, 2) // executes at 5
	b.Transfer(4, "A", "B", 50)       // A = 150

	b.Deposit(5, "A", 0) // triggers payment

	// payment executes → A: 150 - 100 = 50

	if b.accounts["A"].balance != 50 {
		t.Fatalf("expected 50, got %d", b.accounts["A"].balance)
	}

	if b.accounts["B"].balance != 50 {
		t.Fatalf("expected B = 50")
	}
}

func Test_MultiAccount_HeavyFlow(t *testing.T) {
	b := NewBankingSystemImpl()

	b.CreateAccount(1, "A")
	b.CreateAccount(1, "B")
	b.CreateAccount(1, "C")

	b.Deposit(2, "A", 300)
	b.Deposit(2, "B", 200)
	b.Deposit(2, "C", 100)

	b.SchedulePayment(3, "A", 100, 2) // 5
	b.SchedulePayment(3, "B", 50, 2)  // 5
	b.SchedulePayment(3, "C", 100, 2) // 5

	b.Transfer(4, "A", "B", 50) // A:250

	b.Deposit(5, "A", 0)

	// Expected:
	// A: 250 - 100 = 150 (outgoing: 150 total)
	// B: 200 + 50 - 50 = 200 (outgoing: 50)
	// C: 100 - 100 = 0 (outgoing: 100)

	res := b.TopSpenders(6, 3)

	expected := []string{"A(150)", "C(100)", "B(50)"}

	for i := range expected {
		if res[i] != expected[i] {
			t.Fatalf("expected %v got %v", expected, res)
		}
	}
}

func Test_SameTimestamp_Ordering(t *testing.T) {
	b := NewBankingSystemImpl()

	b.CreateAccount(1, "A")
	b.CreateAccount(1, "B")

	b.Deposit(2, "A", 100)

	b.SchedulePayment(3, "A", 80, 2) // executes at 5

	// At timestamp 5:
	// payment executes FIRST → A = 20
	res := b.Transfer(5, "A", "B", 30)

	if res != nil {
		t.Fatalf("transfer should fail due to insufficient funds")
	}
}

func Test_Cancel_Reschedule(t *testing.T) {
	b := NewBankingSystemImpl()

	b.CreateAccount(1, "A")
	b.Deposit(2, "A", 200)

	p1 := b.SchedulePayment(3, "A", 100, 2) // 5

	b.CancelPayment(4, "A", *p1)

	_ = b.SchedulePayment(4, "A", 50, 1) // also 5

	b.Deposit(5, "A", 0)

	if b.accounts["A"].balance != 150 {
		t.Fatalf("expected 150 got %d", b.accounts["A"].balance)
	}
}

func Test_PartialExecutionChain(t *testing.T) {
	b := NewBankingSystemImpl()

	b.CreateAccount(1, "A")
	b.Deposit(2, "A", 120)

	b.SchedulePayment(3, "A", 50, 2) // payment1
	b.SchedulePayment(3, "A", 50, 2) // payment2
	b.SchedulePayment(3, "A", 50, 2) // payment3

	b.Deposit(5, "A", 0)

	// payment1 → 70
	// payment2 → 20
	// payment3 → FAIL

	if b.accounts["A"].balance != 20 {
		t.Fatalf("expected 20 got %d", b.accounts["A"].balance)
	}
}

func Test_Cancel_JustBeforeExecution(t *testing.T) {
	b := NewBankingSystemImpl()

	b.CreateAccount(1, "A")
	b.Deposit(2, "A", 100)

	pid := b.SchedulePayment(3, "A", 50, 2) // executes at 5

	ok := b.CancelPayment(4, "A", *pid)
	if !ok {
		t.Fatalf("cancel should succeed")
	}

	b.Deposit(5, "A", 0)

	if b.accounts["A"].balance != 100 {
		t.Fatalf("payment should not execute")
	}
}

func Test_Cancel_AtSameTimestamp(t *testing.T) {
	b := NewBankingSystemImpl()

	b.CreateAccount(1, "A")
	b.Deposit(2, "A", 100)

	pid := b.SchedulePayment(3, "A", 50, 2) // executes at 5

	ok := b.CancelPayment(5, "A", *pid)

	// payment executes BEFORE cancel → cancel fails
	if ok {
		t.Fatalf("cancel should fail")
	}

	if b.accounts["A"].balance != 50 {
		t.Fatalf("payment should have executed")
	}
}

func Test_HeavyScenario(t *testing.T) {
	b := NewBankingSystemImpl()

	b.CreateAccount(1, "A")
	b.CreateAccount(1, "B")

	b.Deposit(2, "A", 500)

	b.SchedulePayment(3, "A", 200, 3) // 6
	b.SchedulePayment(3, "A", 100, 2) // 5

	b.Transfer(4, "A", "B", 100) // A=400

	b.Deposit(5, "A", 0)
	// payment(100) executes → A=300

	b.Transfer(5, "A", "B", 200) // A=100

	b.Deposit(6, "A", 0)
	// payment(200) executes → A= -100? NO → should fail (insufficient)

	if b.accounts["A"].balance != 100 {
		t.Fatalf("expected 100 got %d", b.accounts["A"].balance)
	}

	res := b.TopSpenders(7, 2)

	expected := []string{"A(400)", "B(0)"}

	for i := range expected {
		if res[i] != expected[i] {
			t.Fatalf("expected %v got %v", expected, res)
		}
	}
}
