package bitcoin_load_spike

import "testing"

func TestFilePrefix(t *testing.T) {
	expectedPrefix := "prefix"

	cl := CumulativeLogger{
		[]*cumulativePlot{},
		expectedPrefix,
	}

	if cl.FilePrefix() != expectedPrefix {
		t.Error("Expected file prefix", expectedPrefix, ", got", cl.FilePrefix())
	}
}

func TestFileExtension(t *testing.T) {
	expectedExtension := "cl-dat"

	cl := CumulativeLogger{
		[]*cumulativePlot{},
		"",
	}

	if cl.FileExtension() != expectedExtension {
		t.Error("Expected file extension", expectedExtension, ", got", cl.FileExtension())
	}
}

var logTests = []struct {
	blockTimestamp float64
	t              txn
	expectedBucket int64
	shouldRecover  bool
}{
	{
		0.0,
		txn{0.0, 0.0},
		0,
		false,
	},
	{
		10.0,
		txn{0.0, 0.0},
		2000,
		false,
	},
	{
		10000.0,
		txn{0.0, 0.0},
		5000,
		false,
	},
	{
		100000000000000000000, // Some very high number
		txn{0.0, 0.0},
		0, // Not used, test for panicking instead
		true,
	},
}

func TestLog(t *testing.T) {
	for _, test := range logTests {
		// Test for panicking if index would be out of bounds
		if test.shouldRecover {
			defer func() {
				if r := recover(); r != nil {
					if r != "Not enough buckets to record txn confirmation time." {
						t.Error("Panic message different from expected, got", r)
					}
				} else {
					t.Error("Expected log to panic if bucket index would cause array out of bounds error")
				}
			}()
		}

		cl := CumulativeLogger{
			[]*cumulativePlot{newCumulativePlot()},
			"",
		}
		cl.Log(test.blockTimestamp, test.t)

		if cl.plots[0].buckets[test.expectedBucket] != 1 {
			// find actual bucket
			var actualBucket int
			for i, count := range cl.plots[0].buckets {
				if count == 1 {
					actualBucket = i
					break
				}
			}
			t.Error("Expected bucket", test.expectedBucket, "to be incremented for block timestamp", test.blockTimestamp, "and txn timestamp", test.t.time, ", got", actualBucket)
		}

		if cl.plots[0].smallestBucket != test.expectedBucket {
			t.Error("Expected smallestBucket", test.expectedBucket, "but found", cl.plots[0].smallestBucket)
		}

		if cl.plots[0].largestBucket != test.expectedBucket {
			t.Error("Expected largestBucket", test.expectedBucket, "but found", cl.plots[0].largestBucket)
		}
	}
}

func TestOutput(t *testing.T) {
	expectedOutput := ("" + /* Empty string to satisfy gofmt/ */
		"3999 | 997.700064 | 0.400000 | 0.400000\n" +
		"4000 | 1000.000000 | 0.600000 | 1.000000\n")

	cl := CumulativeLogger{
		[]*cumulativePlot{newCumulativePlot()},
		"",
	}
	for i := float64(0); i < 5; i++ {
		cl.Log(1000.0, txn{i, 0})
	}

	output := cl.Outputs()[0]

	if output != expectedOutput {
		t.Error("Expected output '", expectedOutput, "', got '", output, "'")
	}
}

func TestCumulativePlotBucketRange(t *testing.T) {
	cp := newCumulativePlot()

	cp.incrementBucket(4)
	cp.incrementBucket(7)
	cp.incrementBucket(7)
	cp.incrementBucket(9)

	output := cp.output()

	expected := ("" + /* Empty string to satisfy gofmt/ */
		"4 | 0.100925 | 0.250000 | 0.250000\n" +
		"5 | 0.101158 | 0.000000 | 0.250000\n" +
		"6 | 0.101391 | 0.000000 | 0.250000\n" +
		"7 | 0.101625 | 0.500000 | 0.750000\n" +
		"8 | 0.101859 | 0.000000 | 0.750000\n" +
		"9 | 0.102094 | 0.250000 | 1.000000\n")

	if output != expected {
		t.Error("Unexpected output:\n", output, "Was expecting:\n", expected)
	}
}
