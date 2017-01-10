package perfmon

import (
	"syscall"
	"unsafe"

	"github.com/elastic/beats/libbeat/common"
)

type Handle struct {
	status      error
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

func GetHandle(config []CounterConfig) (*Handle, error) {
	q := &Handle{query: nil}
	err := _PdhOpenQuery(nil, 1, q.query)
	if err != nil {
		return nil, err
	}
	counters := make([]Counter, len(config))
	q.counters = counters
	for i, v := range config {
		counters[i] = Counter{counterPath: v.Query, counterName: v.Alias}
		err := _PdhAddCounter(*q.query, counters[i].counterPath, 0, &counters[i].counter)
		if err != nil {
			return q, err
		}
	}

	return q, nil
}

func (q *Handle) ReadData() ([]common.MapStr, error) {
	result := make([]common.MapStr, len(q.counters))
	err := _PdhCollectQueryData(*q.query)

	if err != nil {
		return nil, err
	}

	for i, v := range q.counters {
		q.status = _PdhGetFormattedCounterValue(v.counter, PdhFmtDouble, q.counterType, &v.displayValue)
		if q.status != nil {
			//err := errors.New("PdhGetFormattedCounterValue failed with status " + strconv.Itoa(q.status))
			return nil, err
		}

		doubleValue := (*float64)(unsafe.Pointer(&v.displayValue.Pad_cgo_0))

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
//sys   _PdhOpenQuery(dataSource *string, userData int, query *syscall.Handle) (err error) = pdh.PdhOpenQuery
//sys   _PdhAddCounter(query syscall.Handle, counterPath string, userData int, counter *syscall.Handle) (err error) = pdh.PdhAddCounterW
//sys   _PdhCollectQueryData(query syscall.Handle) (err error) = pdh.PdhCollectQueryData
//sys   _PdhGetFormattedCounterValue(counter syscall.Handle, format int, counterType int, value *PdhCounterValue) (err error) = pdh.GetFormattedCounterValue
