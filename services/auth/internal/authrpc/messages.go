package authrpc

import "fmt"

const ServiceName = "auth.v1.AuthService"

type CheckUserRequest struct {
	UserId   string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	TenantId string `protobuf:"bytes,2,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
}

func (m *CheckUserRequest) Reset()         { *m = CheckUserRequest{} }
func (m *CheckUserRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*CheckUserRequest) ProtoMessage()    {}

type CheckUserResponse struct {
	UserId     string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	TenantId   string `protobuf:"bytes,2,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	TenantSlug string `protobuf:"bytes,3,opt,name=tenant_slug,json=tenantSlug,proto3" json:"tenant_slug,omitempty"`
	Role       string `protobuf:"bytes,4,opt,name=role,proto3" json:"role,omitempty"`
	IsActive   bool   `protobuf:"varint,5,opt,name=is_active,json=isActive,proto3" json:"is_active,omitempty"`
}

func (m *CheckUserResponse) Reset()         { *m = CheckUserResponse{} }
func (m *CheckUserResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*CheckUserResponse) ProtoMessage()    {}

type UserStatsRequest struct{}

func (m *UserStatsRequest) Reset()         { *m = UserStatsRequest{} }
func (m *UserStatsRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (*UserStatsRequest) ProtoMessage()    {}

type UserStatsResponse struct {
	Total    int64              `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	Active   int64              `protobuf:"varint,2,opt,name=active,proto3" json:"active,omitempty"`
	Inactive int64              `protobuf:"varint,3,opt,name=inactive,proto3" json:"inactive,omitempty"`
	ByRole   []*RoleUserCount   `protobuf:"bytes,4,rep,name=by_role,json=byRole,proto3" json:"by_role,omitempty"`
	ByTenant []*TenantUserCount `protobuf:"bytes,5,rep,name=by_tenant,json=byTenant,proto3" json:"by_tenant,omitempty"`
}

func (m *UserStatsResponse) Reset()         { *m = UserStatsResponse{} }
func (m *UserStatsResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (*UserStatsResponse) ProtoMessage()    {}

type RoleUserCount struct {
	Role  string `protobuf:"bytes,1,opt,name=role,proto3" json:"role,omitempty"`
	Count int64  `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

func (m *RoleUserCount) Reset()         { *m = RoleUserCount{} }
func (m *RoleUserCount) String() string { return fmt.Sprintf("%+v", *m) }
func (*RoleUserCount) ProtoMessage()    {}

type TenantUserCount struct {
	TenantId   string `protobuf:"bytes,1,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	TenantName string `protobuf:"bytes,2,opt,name=tenant_name,json=tenantName,proto3" json:"tenant_name,omitempty"`
	TenantSlug string `protobuf:"bytes,3,opt,name=tenant_slug,json=tenantSlug,proto3" json:"tenant_slug,omitempty"`
	Count      int64  `protobuf:"varint,4,opt,name=count,proto3" json:"count,omitempty"`
}

func (m *TenantUserCount) Reset()         { *m = TenantUserCount{} }
func (m *TenantUserCount) String() string { return fmt.Sprintf("%+v", *m) }
func (*TenantUserCount) ProtoMessage()    {}
