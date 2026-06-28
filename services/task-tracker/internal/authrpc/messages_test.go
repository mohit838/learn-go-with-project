package authrpc

import (
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestCheckUserResponseProtoRoundTrip(t *testing.T) {
	t.Parallel()

	in := &CheckUserResponse{
		UserId:     "user_01",
		TenantId:   "tenant_01",
		TenantSlug: "default",
		Role:       "superadmin",
		IsActive:   true,
	}

	data, err := proto.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out CheckUserResponse
	if err := proto.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.UserId != "user_01" || out.TenantId != "tenant_01" || !out.IsActive {
		t.Fatalf("unexpected round trip result: %+v", out)
	}
}
