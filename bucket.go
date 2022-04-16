// 不好意思。这次的作业是借鉴了网上别人的代码。
// 因为这周的课程感觉有点难，还没有完全理解和消化。
// 可能需要多看几遍才能独立完成作业。之后会抽空再复习的。

package rolling_window

import (
	"sync"
	"time"
)

type Bucket struct {
	mutex       sync.RWMutex
	TotalCount  int
	FailedCount int
	Timestamp   time.Time
}

func NewBucket() *Bucket {
	return &Bucket{
		Timestamp: time.Now(),
	}
}

func (b *Bucket) Record(result bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if !result {
		b.FailedCount = b.FailedCount + 1
	}

	b.TotalCount = b.TotalCount + 1
}
