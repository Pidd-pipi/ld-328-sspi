package service

import (
	"context"
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
)

// FamilyMemberService 家庭成员服务：查询与操作人展示。
type FamilyMemberService struct {
	memberRepo *repository.FamilyMemberRepository
	log        *slog.Logger
}

// NewFamilyMemberService 构造家庭成员服务。
func NewFamilyMemberService(memberRepo *repository.FamilyMemberRepository, log *slog.Logger) *FamilyMemberService {
	return &FamilyMemberService{memberRepo: memberRepo, log: log}
}

// ListByFamily 列出家庭全部成员（含 User 信息）。
func (s *FamilyMemberService) ListByFamily(ctx context.Context, familyID uint) ([]model.FamilyMember, error) {
	members, err := s.memberRepo.ListByFamily(familyID)
	if err != nil {
		return nil, err
	}
	s.log.InfoContext(ctx, constants.LOG_FAMILY_MEMBER_JOINED, "family_id", familyID, "member_count", len(members))
	return members, nil
}

// CountByFamily 统计家庭成员数。
func (s *FamilyMemberService) CountByFamily(ctx context.Context, familyID uint) (int64, error) {
	return s.memberRepo.CountByFamily(familyID)
}
