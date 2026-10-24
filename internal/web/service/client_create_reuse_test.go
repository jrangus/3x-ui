package service

import (
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func settingsHoldUUID(t *testing.T, inboundSvc *InboundService, inboundId int, uuid string) bool {
	t.Helper()
	ib, err := inboundSvc.GetInbound(inboundId)
	if err != nil {
		t.Fatalf("GetInbound %d: %v", inboundId, err)
	}
	return strings.Contains(ib.Settings, uuid)
}

func TestCreateRepeatKeepsExistingUUID(t *testing.T) {
	setupBulkDB(t)
	svc := &ClientService{}
	inboundSvc := &InboundService{}

	ibA := mkInbound(t, 21001, model.VLESS, `{"clients":[]}`)
	ibB := mkInbound(t, 21002, model.VLESS, `{"clients":[]}`)

	const originalUUID = "aaaaaaaa-1111-2222-3333-444444444444"
	if _, err := svc.Create(inboundSvc, &ClientCreatePayload{
		Client:     model.Client{Email: "repeat@x", ID: originalUUID, SubID: "sub-repeat", Enable: true},
		InboundIds: []int{ibA.Id},
	}); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if rec := lookupClientRecord(t, "repeat@x"); rec.UUID != originalUUID {
		t.Fatalf("record UUID after first Create = %q, want %q", rec.UUID, originalUUID)
	}

	if _, err := svc.Create(inboundSvc, &ClientCreatePayload{
		Client:     model.Client{Email: "repeat@x", SubID: "sub-repeat", Enable: true},
		InboundIds: []int{ibB.Id},
	}); err != nil {
		t.Fatalf("repeat Create: %v", err)
	}

	if rec := lookupClientRecord(t, "repeat@x"); rec.UUID != originalUUID {
		t.Fatalf("record UUID after repeat Create = %q, want %q", rec.UUID, originalUUID)
	}
	if !settingsHoldUUID(t, inboundSvc, ibA.Id, originalUUID) {
		t.Fatalf("inbound A settings lost the original UUID")
	}
	if !settingsHoldUUID(t, inboundSvc, ibB.Id, originalUUID) {
		t.Fatalf("inbound B settings did not reuse the original UUID")
	}
}

func TestCreateAllowsSharedSubIDOnlyForInternalCaller(t *testing.T) {
	setupBulkDB(t)
	svc := &ClientService{}
	inboundSvc := &InboundService{}
	inbound := mkInbound(t, 21003, model.VLESS, `{"clients":[]}`)

	first := ClientCreatePayload{
		Client:     model.Client{Email: "pool-us@exit.local", SubID: "pool-shared", Enable: true},
		InboundIds: []int{inbound.Id},
	}
	if _, err := svc.Create(inboundSvc, &first); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	second := ClientCreatePayload{
		Client:           model.Client{Email: "pool-jp@exit.local", SubID: "pool-shared", Enable: true},
		InboundIds:       []int{inbound.Id},
		AllowSharedSubID: true,
	}
	if _, err := svc.Create(inboundSvc, &second); err != nil {
		t.Fatalf("shared-subId Create: %v", err)
	}

	third := ClientCreatePayload{
		Client:     model.Client{Email: "ordinary@x", SubID: "pool-shared", Enable: true},
		InboundIds: []int{inbound.Id},
	}
	if _, err := svc.Create(inboundSvc, &third); err == nil {
		t.Fatal("ordinary Create unexpectedly reused a shared subId")
	}
}
