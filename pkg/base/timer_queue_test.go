package base

import (
    "time"
    "testing"
    "github.com/stretchr/testify/assert"
)

type TestTimeout struct {
    timeout time.Time
    id      uint64
}

func (t TestTimeout) Timeout() time.Time {
    return t.timeout
}

func (t TestTimeout) TimerID() uint64 {
    return t.id
}

func (t TestTimeout) HandleTimeout(id uint64) error {
    return nil
}

func TestTimeoutBasic(t *testing.T) {
    que := NewTimeoutQueue()
    assert.NotNil(t, que, "create timer queue")

    var t1, t2 TestTimeout
    cur := time.Now()
    expire := cur.Add(time.Second * 10)
    t1.timeout = cur
    t1.id = 1
    t2.timeout = expire
    t2.id = 2

    // items adding
    que.Push(&t2)
    que.Push(&t1)
    assert.Equal(t, 2, que.Len(), "check timer queue length")

    val := que.Peek()
    t.Logf("t1=%p, val = %p", &t1, val)
    assert.Equal(t, &t1, val, "check timer queue item peek")
    assert.True(t, val.Timeout().Equal(cur), "check timer queue item peek")

    val = que.Pop()
    assert.True(t, val.Timeout().Equal(cur), "check timer queue item peek")
    assert.Equal(t, 1, que.Len(), "check timer queue length")

    val = que.Peek()
    t.Logf("t2=%p, val = %p", &t2, val)
    assert.Equal(t, &t2, val, "check timer queue item peek")
    assert.True(t, val.Timeout().Equal(expire), "check timer queue item peek")

    // item sorting
    que.Push(&t1)
    val = que.Peek()
    assert.Equal(t, &t1, val, "check timer queue item sorting")
    assert.True(t, val.Timeout().Equal(cur), "check timer queue item sorting")

    // remove item
    que.Remove(1)
    assert.Equal(t, 1, que.Len(), "check timer queue item removing")
    val = que.Peek()
    assert.Equal(t, &t2, val, "check timer queue item removing")
    assert.True(t, val.Timeout().Equal(expire), "check timer queue item removing")
}
