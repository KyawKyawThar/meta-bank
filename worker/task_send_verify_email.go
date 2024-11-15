package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) DistributorSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)

	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueue task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {

	log.Info().Bytes("payload", task.Payload()).Msg("ProcessTaskSendVerifyEmail function should called.")
	var payload PayloadSendVerifyEmail

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", asynq.SkipRetry)
	}
	//user, err := processor.store.GetUser(ctx, payload.Username)
	//if err != nil {
	//
	//	if errors.Is(err, pgx.ErrNoRows) {
	//		return fmt.Errorf("user does not exists %w", asynq.SkipRetry)
	//	}
	//
	//	return fmt.Errorf("failed to get user: %w", err)
	//}
	//
	//arg := db.CreateVerifyEmailParams{
	//	Username:   user.Username,
	//	Email:      user.Email,
	//	SecretCode: util.RandomString(32),
	//}

	verifyEmail, err := processor.store.GetVerifyEmail(ctx, payload.Username)
	fmt.Println("verifyEmail is:", verifyEmail, err)

	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	subject := "Welcome to Simple Bank"
	// TODO: replace this URL with an environment variable that points to a front-end page
	verifyUrl := fmt.Sprintf("http://localhost:8080/v1/verify_email?email_id=%d&secret_code=%s",
		verifyEmail.ID, verifyEmail.SecretCode)
	content := fmt.Sprintf(`Hello %s,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>
	`, verifyEmail.Username, verifyUrl)
	to := []string{verifyEmail.Email}

	fmt.Println("code is run in here.....")
	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)

	fmt.Println("sendEmail error", err)
	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Str("email", verifyEmail.Email).Msg("processed task")

	return nil
}
