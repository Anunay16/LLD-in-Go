# Problem 1: Concurrent FizzBuzz

## Problem Statement
You are given an integer `n`. You need to output the FizzBuzz sequence from `1` to `n` in strict sequential order using **4 separate goroutines**:

1. **`fizz` Goroutine**: Prints `"Fizz "` when the current number is divisible by 3 and not by 5.
2. **`buzz` Goroutine**: Prints `"Buzz "` when the current number is divisible by 5 and not by 3.
3. **`fizzbuzz` Goroutine**: Prints `"FizzBuzz "` when the current number is divisible by both 3 and 5.
4. **`number` Goroutine**: Prints the number itself (`"<num> "`) when the number is not divisible by 3 or 5.

### Example Output for `n = 15`:
```
1 2 Fizz 4 Buzz Fizz 7 8 Fizz Buzz 11 Fizz 13 14 FizzBuzz
```

---

## Constraints & Rules
- Each of the 4 printing responsibilities must be handled by its dedicated goroutine.
- The output must appear in exact order `1, 2, Fizz, 4, ...` without race conditions.
- No deadlocks. All goroutines must exit cleanly when reaching `n`.
- Avoid busy waiting (e.g. infinite loops without channel blocking or synchronization).

---

## Hints & Strategies
- **Turn Signaling**: Similar to `print_abc_concurrent` or `odd_even_concurrent`, you can pass turns using channels.
- **Controller/Dispatcher approach**: One dispatcher or turn-channel can tell the appropriate goroutine when it is its turn to print, and wait for it to finish before proceeding to the next number.
- **Dedicated Channels**: You can give each of the 4 goroutines a channel to receive the current number to print, and a done channel to signal back when printing is complete.

---

## Running the Code
```bash
go run fizzbuzz_concurrent/fizzbuzz.go
```

