package notificationrpc

import "fmt"

const ServiceName = "notification.v1.NotificationService"

type SubmitNotificationRequest struct {
	UserId     string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	TenantId   string `protobuf:"bytes,2,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	TenantSlug string `protobuf:"bytes,3,opt,name=tenant_slug,json=tenantSlug,proto3" json:"tenant_slug,omitempty"`
	Role       string `protobuf:"bytes,4,opt,name=role,proto3" json:"role,omitempty"`
	Recipient  string `protobuf:"bytes,5,opt,name=recipient,proto3" json:"recipient,omitempty"`
	Message    string `protobuf:"bytes,6,opt,name=message,proto3" json:"message,omitempty"`
}

func (m *SubmitNotificationRequest) Reset()         { *m = SubmitNotificationRequest{} }
func (m *SubmitNotificationRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*SubmitNotificationRequest) ProtoMessage()    {}

type GetNotificationRequest struct {
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (m *GetNotificationRequest) Reset()         { *m = GetNotificationRequest{} }
func (m *GetNotificationRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*GetNotificationRequest) ProtoMessage()    {}

type StatsRequest struct{}

func (m *StatsRequest) Reset()         { *m = StatsRequest{} }
func (m *StatsRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*StatsRequest) ProtoMessage()    {}

type NotificationResponse struct {
	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	UserId      string `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	TenantId    string `protobuf:"bytes,3,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	Recipient   string `protobuf:"bytes,4,opt,name=recipient,proto3" json:"recipient,omitempty"`
	Message     string `protobuf:"bytes,5,opt,name=message,proto3" json:"message,omitempty"`
	Status      string `protobuf:"bytes,6,opt,name=status,proto3" json:"status,omitempty"`
	WorkerId    int32  `protobuf:"varint,7,opt,name=worker_id,json=workerId,proto3" json:"worker_id,omitempty"`
	Error       string `protobuf:"bytes,8,opt,name=error,proto3" json:"error,omitempty"`
	SubmittedAt string `protobuf:"bytes,9,opt,name=submitted_at,json=submittedAt,proto3" json:"submitted_at,omitempty"`
	UpdatedAt   string `protobuf:"bytes,10,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

func (m *NotificationResponse) Reset()         { *m = NotificationResponse{} }
func (m *NotificationResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*NotificationResponse) ProtoMessage()    {}

type StatsResponse struct {
	Workers     int32 `protobuf:"varint,1,opt,name=workers,proto3" json:"workers,omitempty"`
	QueueSize   int32 `protobuf:"varint,2,opt,name=queue_size,json=queueSize,proto3" json:"queue_size,omitempty"`
	Queued      int32 `protobuf:"varint,3,opt,name=queued,proto3" json:"queued,omitempty"`
	Sending     int32 `protobuf:"varint,4,opt,name=sending,proto3" json:"sending,omitempty"`
	Delivered   int32 `protobuf:"varint,5,opt,name=delivered,proto3" json:"delivered,omitempty"`
	Failed      int32 `protobuf:"varint,6,opt,name=failed,proto3" json:"failed,omitempty"`
	JobsWaiting int32 `protobuf:"varint,7,opt,name=jobs_waiting,json=jobsWaiting,proto3" json:"jobs_waiting,omitempty"`
}

func (m *StatsResponse) Reset()         { *m = StatsResponse{} }
func (m *StatsResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*StatsResponse) ProtoMessage()    {}
