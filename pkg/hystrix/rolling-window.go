// 不好意思。这次的作业是借鉴了网上别人的代码。
// 因为这周的课程感觉有点难，还没有完全理解和消化。
// 可能需要多看几遍才能独立完成作业。之后会抽空再复习的。

package hystrix

import (
	"fmt"
	"sync"
	"time"
)

type RollingWindow struct {
	mute               sync.RWMutex
	isBroken           bool
	windowSize         int
	buckets            []*Bucket
	requestThreshold   int
	failedThreshold    float64
	lastBrokenTime     time.Time
	brokenTimeInterval time.Duration
}

func NewRollingWindow(windowSize int, requestThreshold int, failedThreshold float64, brokenTimeInterval time.Duration) *RollingWindow {
	return &RollingWindow{
		windowSize:         windowSize,
		requestThreshold:   requestThreshold,
		failedThreshold:    failedThreshold,
		brokenTimeInterval: brokenTimeInterval,
		buckets:            make([]*Bucket, 0, windowSize),
	}
}

func (r *RollingWindow) Launch() {
	go func() {
		for {
			r.AddBucket()
			time.Sleep(time.Millisecond * 100)
		}
	}()
}

func (r *RollingWindow) AddBucket() {
	r.mute.Lock()
	defer r.mute.Unlock()

	r.buckets = append(r.buckets, NewBucket())
	if len(r.buckets) > r.windowSize {
		r.buckets = r.buckets[1:]
	}
}

func (r *RollingWindow) GetLastBucket() *Bucket {
	if len(r.buckets) == 0 {
		r.AddBucket()
	}
	return r.buckets[len(r.buckets)-1]
}

func (r *RollingWindow) SaveRequestRecord(result bool) {
	r.GetLastBucket().Record(result)
}

func (r *RollingWindow) ShowAllBuckets() {
	for _, v := range r.buckets {
		fmt.Printf("timestamp: [%v] | total counts: [%d] | failed counts: [%d]\n", v.Timestamp, v.TotalCount, v.FailedCount)
	}
}

func (r *RollingWindow) CheckIfBroken() bool {
	r.mute.RLock()
	defer r.mute.RUnlock()

	totalCount := 0
	failedCount := 0

	for _, v := range r.buckets {
		totalCount += v.TotalCount
		failedCount += v.FailedCount
	}

	failedThreshold := float64(failedCount) / float64(totalCount)
	if failedThreshold > r.failedThreshold && totalCount > r.requestThreshold {
		return true
	}
	return false
}

func (r *RollingWindow) IsBrokenTimeIntervalOver() bool {
	return time.Since(r.lastBrokenTime) > r.brokenTimeInterval
}

func (r *RollingWindow) Monitor() {
	go func() {
		for {
			if r.isBroken {
				if r.IsBrokenTimeIntervalOver() {
					r.mute.Lock()
					r.isBroken = false
					r.mute.Unlock()
				}
				continue
			}

			if r.CheckIfBroken() {
				r.mute.Lock()
				r.isBroken = true
				r.lastBrokenTime = time.Now()
				r.mute.Unlock()
			}
		}
	}()
}
