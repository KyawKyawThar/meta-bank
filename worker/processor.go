package worker

import (
	"context"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/mail"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

// TaskProcessor define pickup the tasks from redis queue and process them
type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
	ShutDown()
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
	mailer mail.EmailSender
}

func NewRedisTaskProcessor(redisOpts asynq.RedisClientOpt, store db.Store, mailer mail.EmailSender) TaskProcessor {
	logger := NewLogger()
	redis.SetLogger(logger)

	server := asynq.NewServer(redisOpts,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			// RetryDelayFunc calculates retry delay with exponential backoff.
			//RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
			//	baseDelay := 10 * time.Second
			//	// Calculate the delay as baseDelay * 2^(n-1), capped by a maximum duration if needed
			//
			//	maxDelay := 10 * time.Minute        // Maximum delay (optional)
			//	delay := baseDelay * (1 << (n - 1)) // 2^(n-1) exponential backoff
			//
			//	if delay > maxDelay {
			//		return maxDelay
			//	}
			//	return delay
			//},
			//ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			//	log.Error().Err(err).Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("process task failed")
			//}),
			Logger: logger,
		})

	return &RedisTaskProcessor{
		server: server,
		store:  store,
		mailer: mailer,
	}
}

func (processor *RedisTaskProcessor) ShutDown() {
	processor.server.Shutdown()

}

func (processor *RedisTaskProcessor) Start() error {

	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)

	return processor.server.Start(mux)
}
