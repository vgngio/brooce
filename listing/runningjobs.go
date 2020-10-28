package listing

import (
	"brooce/config"
	"brooce/heartbeat"
	myredis "brooce/redis"
	"brooce/task"

	"github.com/go-redis/redis/v8"
)

var redisClient, ctx = myredis.Get()
var redisHeader = config.Config.ClusterName

// SCAN takes about 0.5s per million total items in redis
// we skip it by guessing the possible working list names
// from worker heartbeat data
// this is much faster, but the prune functions still need
// the true list to find any zombie working lists
func RunningJobs(fast bool) (jobs []*task.Task, err error) {
	jobs = []*task.Task{}

	var keys []string

	if fast {
		var workers []*heartbeat.HeartbeatType

		workers, err = RunningWorkers()
		if err != nil {
			return
		}

		for _, worker := range workers {
			for _, thread := range worker.Threads {
				keys = append(keys, thread.WorkingList())
			}
		}

	} else {
		keys, err = myredis.ScanKeys(redisHeader + ":queue:*:working:*")
		if err != nil {
			return
		}
	}

	if len(keys) == 0 {
		return
	}

	values := make([]*redis.StringCmd, len(keys))
	_, err = redisClient.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, key := range keys {
			values[i] = pipe.LIndex(ctx, key, 0)
		}
		return nil
	})
	// it's possible for an item to vanish between the KEYS and LINDEX steps -- this is not fatal!
	if err == redis.Nil {
		err = nil
	}
	if err != nil {
		return
	}

	for i, value := range values {
		if value.Err() != nil {
			// possible to get a redis.Nil error here if a job vanished between the KEYS and LINDEX steps
			continue
		}
		job, err := task.NewFromJson(value.Val(), task.QueueNameFromRedisKey(keys[i]))
		if err != nil {
			continue
		}
		job.RedisKey = keys[i]
		jobs = append(jobs, job)
	}

	task.PopulateHasLog(jobs)
	return
}
