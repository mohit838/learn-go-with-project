package authrpc

import (
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestUserStatsResponseProtoRoundTrip(t *testing.T) {
	t.Parallel()

	in := &UserStatsResponse{
		Total:    3,
		Active:   2,
		Inactive: 1,
		ByRole: []*RoleUserCount{
			{Role: "superadmin", Count: 1},
			{Role: "staff", Count: 2},
		},
		ByTenant: []*TenantUserCount{
			{TenantId: "tenant_01", TenantName: "Default", TenantSlug: "default", Count: 3},
		},
	}

	data, err := proto.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out UserStatsResponse
	if err := proto.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.Total != 3 || out.ByRole[0].Role != "superadmin" || out.ByTenant[0].TenantId != "tenant_01" {
		t.Fatalf("unexpected round trip result: %+v", out)
	}
}
