package userauthwiring

import (
	notificationcontract "github.com/dujiao-next/internal/modules/notification/contract"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

type agentRegistrationNotifier struct {
	notifications notificationcontract.NotificationEnqueuer
}

func (n agentRegistrationNotifier) NotifyNewAgentRegistered(email string, userID uint) error {
	if n.notifications == nil {
		return nil
	}
	return n.notifications.Enqueue(notificationcontract.EnqueueInput{
		EventType: "new_agent_registered",
		BizType:   "agent_registration",
		BizID:     userID,
		Data:      jsonmap.JSON{"email": email},
	})
}
