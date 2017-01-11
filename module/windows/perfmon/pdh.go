package perfmon

import (
	"unsafe"

	"github.com/elastic/beats/libbeat/common"
)

type Handle struct {
	status      error
	query       uintptr
	counterType int
	counters    []Counter
}

type Counter struct {
	counterName  string
	counter      uintptr
	counterPath  string
	displayValue PdhCounterValue
}

func GetHandle(config []CounterConfig) (*Handle, int) {
	q := &Handle{}
	err := _PdhOpenQuery(0, 0, &q.query)
	if err != ERROR_SUCCESS {
		return nil, err
	}
	counters := make([]Counter, len(config))
	q.counters = counters
	for i, v := range config {
		counters[i] = Counter{counterPath: v.Query, counterName: v.Alias}
		err := _PdhAddCounter(q.query, counters[i].counterPath, 0, &counters[i].counter)
		if err != ERROR_SUCCESS {
			return q, err
		}
	}

	return q, 0
}

func (q *Handle) ReadData() ([]common.MapStr, int) {
	result := make([]common.MapStr, len(q.counters))
	err := _PdhCollectQueryData(q.query)

	if err != ERROR_SUCCESS {
		return nil, err
	}

	for i, v := range q.counters {
		err := _PdhGetFormattedCounterValue(v.counter, PdhFmtDouble, q.counterType, &v.displayValue)
		if err != ERROR_SUCCESS {
			return nil, err
		}

		doubleValue := (*float64)(unsafe.Pointer(&v.displayValue.LongValue))

		val := common.MapStr{
			"name":  v.counterName,
			"value": *doubleValue,
		}
		result[i] = val
	}
	return result, 0
}

//go:generate go run $GOROOT/src/syscall/mksyscall_windows.go -output pdh_windows.go pdh.go
// Windows API calls
//sys   _PdhOpenQuery(dataSource uintptr, userData uintptr, query *uintptr) (err int) = pdh.PdhOpenQuery
//sys   _PdhAddCounter(query uintptr, counterPath string, userData uintptr, counter *uintptr) (err int) = pdh.PdhAddCounterW
//sys   _PdhCollectQueryData(query uintptr) (err int) = pdh.PdhCollectQueryData
//sys   _PdhGetFormattedCounterValue(counter uintptr, format uint32, counterType int, value *PdhCounterValue) (err int) = pdh.PdhGetFormattedCounterValue
