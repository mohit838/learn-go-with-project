package notificationrpc

import (
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestNotificationResponseProtoRoundTrip(t *testing.T) {
	in := &NotificationResponse{
		Id:        "notif_1",
		UserId:    "user_1",
		TenantId:  "tenant_1",
		Recipient: "dev@example.com",
		Message:   "hello",
		Status:    "queued",
	}

	data, err := proto.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out NotificationResponse
	if err := proto.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.Id != in.Id || out.Recipient != in.Recipient || out.Status != in.Status {
		t.Fatalf("unexpected round trip result: %+v", out)
	}
}
