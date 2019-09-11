package model

import (
	"context"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/evergreen-ci/cedar/util"
	"github.com/evergreen-ci/pail"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serialized   = "SerializedIterator"
	batched      = "BatchedIterator"
	parallelized = "ParallelizedIterator"
	merging      = "MergingIterator"
)

func TestLogIterator(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmpDir, err := ioutil.TempDir(".", "log-iterator-test")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	bucket, err := pail.NewLocalBucket(pail.LocalOptions{Path: tmpDir})
	require.NoError(t, err)
	chunks, lines, err := createLog(ctx, bucket, 100, 30)
	require.NoError(t, err)

	badBucket, err := pail.NewLocalBucket(pail.LocalOptions{Path: tmpDir, Prefix: "DNE"})
	require.NoError(t, err)

	completeTimeRange := util.TimeRange{EndAt: chunks[len(chunks)-1].End}
	partialTimeRange := util.TimeRange{
		StartAt: chunks[1].Start,
		EndAt:   chunks[len(chunks)-1].End.Add(-time.Minute),
	}
	offset := chunks[0].NumLines
	partialLinesLen := len(lines) - chunks[0].NumLines - 1

	for _, test := range []struct {
		name      string
		iterators map[string]LogIterator
		test      func(*testing.T, LogIterator)
	}{
		{
			name: "EmptyIterator",
			iterators: map[string]LogIterator{
				serialized:   NewSerializedLogIterator(bucket, []LogChunkInfo{}, completeTimeRange),
				batched:      NewBatchedLogIterator(bucket, []LogChunkInfo{}, 2, completeTimeRange),
				parallelized: NewParallelizedLogIterator(bucket, []LogChunkInfo{}, completeTimeRange),
				merging:      NewMergingIterator(ctx, NewSerializedLogIterator(bucket, []LogChunkInfo{}, completeTimeRange)),
			},
			test: func(t *testing.T, it LogIterator) {
				assert.False(t, it.Next(ctx))
				assert.NoError(t, it.Err())
				assert.Zero(t, it.Item())
				assert.NoError(t, it.Close())
			},
		},
		{
			name: "ExhaustedIterator",
			iterators: map[string]LogIterator{
				serialized:   NewSerializedLogIterator(bucket, chunks, completeTimeRange),
				batched:      NewBatchedLogIterator(bucket, chunks, 2, completeTimeRange),
				parallelized: NewParallelizedLogIterator(bucket, chunks, completeTimeRange),
				merging:      NewMergingIterator(ctx, NewBatchedLogIterator(bucket, chunks, 2, completeTimeRange)),
			},
			test: func(t *testing.T, it LogIterator) {
				count := 0
				for it.Next(ctx) {
					require.True(t, count < len(lines))
					assert.Equal(t, lines[count], it.Item())
					count++
				}
				assert.Equal(t, len(lines), count)
				assert.NoError(t, it.Err())
				assert.NoError(t, it.Close())
			},
		},
		{
			name: "ErroredIterator",
			iterators: map[string]LogIterator{
				serialized:   NewSerializedLogIterator(badBucket, chunks, completeTimeRange),
				batched:      NewBatchedLogIterator(badBucket, chunks, 2, completeTimeRange),
				parallelized: NewParallelizedLogIterator(badBucket, chunks, completeTimeRange),
				merging:      NewMergingIterator(ctx, NewBatchedLogIterator(badBucket, chunks, 2, completeTimeRange)),
			},
			test: func(t *testing.T, it LogIterator) {
				count := 0
				for it.Next(ctx) {
					count++
				}
				assert.Equal(t, 0, count)
				assert.Error(t, it.Err())
				assert.NoError(t, it.Close())
			},
		},
		{
			name: "LimitedTimeRange",
			iterators: map[string]LogIterator{
				serialized:   NewSerializedLogIterator(bucket, chunks, partialTimeRange),
				batched:      NewBatchedLogIterator(bucket, chunks, 2, partialTimeRange),
				parallelized: NewParallelizedLogIterator(bucket, chunks, partialTimeRange),
				merging:      NewMergingIterator(ctx, NewParallelizedLogIterator(bucket, chunks, partialTimeRange)),
			},
			test: func(t *testing.T, it LogIterator) {
				count := 0
				for it.Next(ctx) {
					require.True(t, offset+count < len(lines))
					assert.Equal(t, lines[offset+count], it.Item())
					count++
				}
				assert.Equal(t, partialLinesLen, count)
				assert.NoError(t, it.Err())
				assert.NoError(t, it.Close())
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			for impl, it := range test.iterators {
				t.Run(impl, func(t *testing.T) {
					test.test(t, it)
				})
			}
		})
	}
}

func TestMergeLogIterator(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tmpDir, err := ioutil.TempDir(".", "merge-logs-test")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	bucket, err := pail.NewLocalBucket(pail.LocalOptions{Path: tmpDir})
	require.NoError(t, err)

	t.Run("SingleLog", func(t *testing.T) {
		chunks, lines, err := createLog(ctx, bucket, 100, 10)
		require.NoError(t, err)
		timeRange := util.TimeRange{
			StartAt: chunks[0].Start,
			EndAt:   chunks[len(chunks)-1].End,
		}
		it := NewMergingIterator(ctx, NewBatchedLogIterator(bucket, chunks, 100, timeRange))

		count := 0
		for it.Next(ctx) {
			logLine := it.Item()
			require.True(t, count < len(lines))
			assert.Equal(t, lines[count], logLine)
			count++
		}
		assert.Equal(t, len(lines), count)
		assert.NoError(t, it.Err())
		assert.NoError(t, it.Close())
	})
	t.Run("MultipleLogs", func(t *testing.T) {
		numLogs := 10
		its := make([]LogIterator, numLogs)
		lineMap := map[string]bool{}
		for i := 0; i < numLogs; i++ {
			chunks, lines, err := createLog(ctx, bucket, 100, 10)
			require.NoError(t, err)

			timeRange := util.TimeRange{
				StartAt: chunks[0].Start,
				EndAt:   chunks[len(chunks)-1].End,
			}
			its[i] = NewBatchedLogIterator(bucket, chunks, 100, timeRange)
			for _, line := range lines {
				lineMap[line.Data] = false
			}
		}
		it := NewMergingIterator(ctx, its...)

		count := 0
		lastTime := time.Time{}
		for it.Next(ctx) {
			logLine := it.Item()

			assert.True(t, lastTime.Before(logLine.Timestamp) || lastTime.Equal(logLine.Timestamp))
			lastTime = logLine.Timestamp
			seen, ok := lineMap[logLine.Data]
			require.True(t, ok)
			assert.False(t, seen)
			lineMap[logLine.Data] = true
			count++
		}
		assert.Equal(t, len(lineMap), count)
		assert.NoError(t, it.Err())
		assert.NoError(t, it.Close())
	})
}

func TestLogIteratorReader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tmpDir, err := ioutil.TempDir(".", "merge-logs-test")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	bucket, err := pail.NewLocalBucket(pail.LocalOptions{Path: tmpDir})
	require.NoError(t, err)

	chunks, lines, err := createLog(ctx, bucket, 100, 10)
	require.NoError(t, err)
	expectedSize := 0
	for _, line := range lines {
		expectedSize += len(line.Data)
	}
	timeRange := util.TimeRange{
		StartAt: chunks[0].Start,
		EndAt:   chunks[len(chunks)-1].End,
	}

	t.Run("LeftOver", func(t *testing.T) {
		r := NewLogIteratorReader(ctx, NewBatchedLogIterator(bucket, chunks, 2, timeRange))
		nTotal := 0
		readData := []byte{}
		p := make([]byte, 22)
		for {
			n, err := r.Read(p)
			nTotal += n
			readData = append(readData, p[:n]...)
			assert.True(t, n >= 0)
			assert.True(t, n <= len(p))
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
		}
		assert.Equal(t, expectedSize, nTotal)

		current := 0
		readLines := strings.Split(string(readData), "\n")
		assert.Equal(t, "", readLines[len(readLines)-1])
		for _, line := range readLines[:len(readLines)-1] {
			require.True(t, current < len(lines))
			assert.Equal(t, lines[current].Data, line+"\n")
			current++
		}
		assert.Equal(t, len(lines), current)

		n, err := r.Read(p)
		assert.Zero(t, n)
		assert.Error(t, err)
	})
	t.Run("EmptyBuffer", func(t *testing.T) {
		r := NewLogIteratorReader(ctx, NewBatchedLogIterator(bucket, chunks, 2, timeRange))
		p := make([]byte, 0)
		for {
			n, err := r.Read(p)
			assert.Zero(t, n)
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
		}

		n, err := r.Read(p)
		assert.Zero(t, n)
		assert.Error(t, err)
	})
	t.Run("ContextError", func(t *testing.T) {
		errCtx, errCancel := context.WithCancel(context.Background())
		errCancel()

		r := NewLogIteratorReader(errCtx, NewBatchedLogIterator(bucket, chunks, 2, timeRange))
		p := make([]byte, 101)
		n, err := r.Read(p)
		assert.Zero(t, n)
		assert.Error(t, err)
	})
}

func createLog(ctx context.Context, bucket pail.Bucket, size, chunkSize int) ([]LogChunkInfo, []LogLine, error) {
	lines := make([]LogLine, size)
	numChunks := size / chunkSize
	if numChunks == 0 || size%chunkSize > 0 {
		numChunks += 1
	}
	chunks := make([]LogChunkInfo, numChunks)
	ts := time.Now().Round(time.Millisecond).UTC()

	for i := 0; i < numChunks; i++ {
		rawLines := ""
		chunks[i] = LogChunkInfo{
			Key:   newRandCharSetString(16),
			Start: ts,
		}

		j := 0
		for j < chunkSize && j+i*chunkSize < size {
			line := newRandCharSetString(100)
			lines[j+i*chunkSize] = LogLine{
				Timestamp: ts,
				Data:      line + "\n",
			}
			rawLines += prependTimestamp(ts, line)
			ts = ts.Add(time.Minute)
			j++
		}

		if err := bucket.Put(ctx, chunks[i].Key, strings.NewReader(rawLines)); err != nil {
			return []LogChunkInfo{}, []LogLine{}, errors.Wrap(err, "failed to add chunk to bucket")
		}

		chunks[i].NumLines = j
		chunks[i].End = ts.Add(-time.Minute)
		ts = ts.Add(time.Hour)
	}

	return chunks, lines, nil
}

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func newRandCharSetString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
