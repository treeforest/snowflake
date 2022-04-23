package snowflake

import (
	"errors"
	"sync"
	"time"
)

/*
* 1                                               42           52             64
* +-----------------------------------------------+------------+---------------+
* | timestamp(ms)                                 | workerid   | sequence      |
* +-----------------------------------------------+------------+---------------+
* | 0000000000 0000000000 0000000000 0000000000 0 | 0000000000 | 0000000000 00 |
* +-----------------------------------------------+------------+---------------+
*
* 1. 41位时间截(毫秒级)，注意这是时间截的差值（当前时间截 - 开始时间截)。可以使用约70年: (1L << 41) / (1000L * 60 * 60 * 24 * 365) = 69
* 2. 10位数据机器位，可以部署在1024个节点
* 3. 12位序列，毫秒内的计数，同一机器，同一时间截并发4096个序号
 */

const (
	twepoch        = int64(1483228800000)             //开始时间截 (2017-01-01)
	workeridBits   = uint(10)                         //机器id所占的位数
	sequenceBits   = uint(12)                         //序列所占的位数
	workeridMax    = int64(-1 ^ (-1 << workeridBits)) //支持的最大机器id数量
	sequenceMask   = int64(-1 ^ (-1 << sequenceBits)) //
	workeridShift  = sequenceBits                     //机器id左移位数
	timestampShift = sequenceBits + workeridBits      //时间戳左移位数
)

// A Snowflake struct holds the basic information needed for a snowflake generator worker
type snowflake struct {
	mu        sync.Mutex
	timestamp int64
	workerid  int64
	sequence  int64
}

// NewSnowflake NewNode returns a new snowflake worker that can be used to generate snowflake IDs
func NewSnowflake(workerid int64) (*snowflake, error) {
	if workerid < 0 || workerid > workeridMax {
		return nil, errors.New("workerid must be between 0 and 1023")
	}

	return &snowflake{
		timestamp: 0,
		workerid:  workerid,
		sequence:  0,
	}, nil
}

// Generate creates and returns a unique snowflake ID
func (s *snowflake) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixNano() / 1000000

	//如果是同一时间生成的，则进行毫秒内序列
	if s.timestamp == now {
		s.sequence = (s.sequence + 1) & sequenceMask
		//毫秒内序列溢出 即 序列 > 4095
		if s.sequence == 0 {
			//获得新的时间戳
			for now <= s.timestamp {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		//时间戳改变，毫秒内序列重置
		s.sequence = 0
	}

	//上次生成ID的时间截
	s.timestamp = now

	//移位并通过或运算拼到一起组成64位的ID
	r := int64((now-twepoch)<<timestampShift | (s.workerid << workeridShift) | (s.sequence))

	return r
}

var defaultInstance *snowflake

func init() {
	var err error
	defaultInstance, err = NewSnowflake(0)
	if err != nil {
		panic(err)
	}
}

func Generate() int64 {
	return defaultInstance.Generate()
}
