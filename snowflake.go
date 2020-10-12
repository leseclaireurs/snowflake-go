package snowflake

import (
	"sync"
	"time"
)

const (
	epoch              uint64 = 1602388800000                                          // 开始时间戳 (2020-10-11 12:00:00)
	workerIdBits       uint8  = 5                                                      // 机器ID所占位数
	dataCenterIdBits   uint8  = 5                                                      // 数据标识ID所占位数
	sequenceBits       uint8  = 12                                                     // 序列在ID中占的位数
	maxWorkerId               = -1 ^ (-1 << workerIdBits)                              // 支持的最大机器id，结果是31 (这个移位算法可以很快的计算出几位二进制数所能表示的最大十进制数)
	maxDatacenterId           = -1 ^ (-1 << dataCenterIdBits)                          // 支持的最大数据标识id，结果是31
	sequenceMask              = -1 ^ (-1 << sequenceBits)                              // 生成序列的掩码，这里为4095 (0b111111111111=0xfff=4095)
	workerIdShift             = uint64(sequenceBits)                                   // 机器ID向左移12位
	datacenterIdShift         = uint64(sequenceBits + workerIdBits)                    // 数据标识id向左移17位(12+5)
	timestampLeftShift        = uint64(sequenceBits + workerIdBits + dataCenterIdBits) // 时间戳向左移22位(5+5+12)
)

type SnowFlake struct {
	workerID      uint64
	datacenterID  uint64
	lastTimestamp uint64
	sequence      uint64
	lock          sync.Mutex
}

func NewSnowFlake(workerID, datacenterID uint64) *SnowFlake {
	if (workerID > maxWorkerId || workerID < 0) || (datacenterID > maxDatacenterId || datacenterID < 0) {
		panic(1)
	}
	return &SnowFlake{
		workerID:     workerID,
		datacenterID: datacenterID,
	}
}

/**
 * 获得下一个ID (该方法是线程安全的)
 * @return SnowflakeId
 */
func (r *SnowFlake) NextID() uint64 {
	r.lock.Lock()
	defer r.lock.Unlock()
	timestamp := r.genMillTimestamp()
	// 如果当前时间小于上一次ID生成的时间戳，说明系统时钟回退过这个时候应当抛出异常
	if timestamp < r.lastTimestamp {
		panic(1)
	}
	// 如果是同一时间生成的，则进行毫秒内序列
	if timestamp == r.lastTimestamp {
		r.sequence = (r.sequence + 1) & sequenceMask
		// 毫秒内序列溢出
		if r.sequence == 0 {
			// 阻塞到下一个毫秒,获得新的时间戳
			timestamp = r.tilNextMillis(r.lastTimestamp)
		}
	} else {
		// 时间戳改变，毫秒内序列重置
		r.sequence = 0
	}
	// 上次生成ID的时间戳
	r.lastTimestamp = timestamp
	id := ((timestamp - epoch) << timestampLeftShift) | (r.datacenterID << datacenterIdShift) | (r.workerID << workerIdShift) | r.sequence
	return id
}

/**
 * 阻塞到下一个毫秒，直到获得新的时间戳
 * @param lastTimestamp 上次生成ID的时间戳
 * @return 当前时间戳
 */
func (r *SnowFlake) tilNextMillis(lastTimestamp uint64) (timestamp uint64) {
	timestamp = r.genMillTimestamp()
	for timestamp <= lastTimestamp {
		timestamp = r.genMillTimestamp()
	}
	return
}

/**
 * 返回以毫秒为单位的当前时间
 * @return 当前时间(毫秒)
 */
func (r *SnowFlake) genMillTimestamp() uint64 {
	return uint64(time.Now().UnixNano() / 1e6)
}
