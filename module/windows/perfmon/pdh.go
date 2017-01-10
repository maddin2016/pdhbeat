package perfmon

import (
	"unsafe"

	"errors"

	"strconv"

	"syscall"

	"github.com/elastic/beats/libbeat/common"
)

type Handle struct {
	status      int
	query       *syscall.Handle
	counterType int
	counters    []Counter
}

type Counter struct {
	counterName  string
	counter      syscall.Handle
	counterPath  string
	displayValue PdhCounterValue
}

func GetHandle(config []CounterConfig) (handle *Handle, err error) {
	q := &Handle{query: nil}
	q.status = _PdhOpenQuery(nil, 0, q.query)
	counters := make([]Counter, len(config))
	q.counters = counters
	for i, v := range config {
		counters[i] = Counter{counterPath: v.Query, counterName: v.Alias}
		q.status = _PdhAddCounter(*q.query, counters[i].counterPath, 0, &counters[i].counter)
		if q.status != ERROR_SUCCESS {
			err := errors.New("PdhAddCounter is failed for " + v.Alias)
			return q, err
		}
	}

	return q, nil
}

func (q *Handle) ReadData() (data []common.MapStr, err error) {
	result := make([]common.MapStr, len(q.counters))
	q.status = _PdhCollectQueryData(*q.query)

	if q.status != ERROR_SUCCESS {
		sP := (*int)(unsafe.Pointer(&q.status))
		s := strconv.Itoa(*sP)
		err := errors.New("PdhCollectQueryData failed with status " + s)
		return nil, err
	}

	for i, v := range q.counters {
		q.status = _PdhGetFormattedCounterValue(*v.counter, PdhFmtDouble, q.counterType, v.displayValue)
		if q.status != ERROR_SUCCESS {
			sP := (*int)(unsafe.Pointer(&q.status))
			s := strconv.Itoa(*sP)
			err := errors.New("PdhGetFormattedCounterValue failed with status " + s)
			return nil, err
		}

		doubleValue := (*float64)(unsafe.Pointer(&v.displayValue))

		val := common.MapStr{
			"name":  v.counterName,
			"value": *doubleValue,
		}
		result[i] = val
	}
	return result, nil
}

//go:generate go run $GOROOT/src/syscall/mksyscall_windows.go -output pdh_windows.go pdh.go
// Windows API calls
//sys   _PdhOpenQuery(dataSource *string, userData int, query *syscall.Handle) (err int) = pdh.PdhOpenQuery
//sys   _PdhAddCounter(query syscall.Handle, counterPath string, userData int, counter *syscall.Handle) (err int) = pdh.PdhAddCounter
//sys   _PdhCollectQueryData(query syscall.Handle) (err int) = pdh.PdhCollectQueryData
//sys   _PdhGetFormattedCounterValue(counter syscall.Handle, format int, counterType int, value PdhCounterValue) (err int) = pdh.GetFormattedCounterValue
