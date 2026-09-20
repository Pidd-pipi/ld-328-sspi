package service

import (
	"context"
	"testing"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/repository"
)

func TestFamilyGroupService_Create(t *testing.T) {
	db := newTestDB(t)
	groupRepo := repository.NewFamilyGroupRepository(db)
	memberRepo := repository.NewFamilyMemberRepository(db)
	svc := NewFamilyGroupService(groupRepo, memberRepo, testLogger())
	ctx := context.Background()

	group, err := svc.Create(ctx, 1, "测试家庭")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if group.ID == 0 || group.InviteCode == "" {
		t.Fatalf("group not persisted: %+v", group)
	}

	member, err := memberRepo.FindByFamilyAndUser(group.ID, 1)
	if err != nil {
		t.Fatalf("FindByFamilyAndUser() error = %v", err)
	}
	if member.Role != constants.FamilyRoleAdmin {
		t.Fatalf("owner role = %s, want %s", member.Role, constants.FamilyRoleAdmin)
	}
}

func TestFamilyGroupService_JoinByCode_Duplicate(t *testing.T) {
	db := newTestDB(t)
	groupRepo := repository.NewFamilyGroupRepository(db)
	memberRepo := repository.NewFamilyMemberRepository(db)
	svc := NewFamilyGroupService(groupRepo, memberRepo, testLogger())
	ctx := context.Background()

	group, err := svc.Create(ctx, 1, "测试家庭")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.JoinByCode(ctx, 2, group.InviteCode); err != nil {
		t.Fatalf("first JoinByCode() error = %v", err)
	}
	if _, err := svc.JoinByCode(ctx, 2, group.InviteCode); err == nil {
		t.Fatal("expected conflict for duplicate join")
	}
}
