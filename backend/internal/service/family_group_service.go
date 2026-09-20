package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

// FamilyGroupService 家庭组服务：创建、邀请、加入、成员管理。
type FamilyGroupService struct {
	groupRepo  *repository.FamilyGroupRepository
	memberRepo *repository.FamilyMemberRepository
	log        *slog.Logger
}

// NewFamilyGroupService 构造家庭组服务。
func NewFamilyGroupService(groupRepo *repository.FamilyGroupRepository, memberRepo *repository.FamilyMemberRepository, log *slog.Logger) *FamilyGroupService {
	return &FamilyGroupService{groupRepo: groupRepo, memberRepo: memberRepo, log: log}
}

// genInviteCode 生成 8 位邀请码。
func genInviteCode() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Create 创建家庭组（创建者自动成为 Admin 成员）。
func (s *FamilyGroupService) Create(ctx context.Context, ownerID uint, name string) (*model.FamilyGroup, error) {
	group := &model.FamilyGroup{Name: name, OwnerID: ownerID, InviteCode: genInviteCode()}
	err := s.groupRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.groupRepo.WithTx(tx).Create(group); err != nil {
			return fmt.Errorf("create family group: %w", err)
		}
		member := &model.FamilyMember{FamilyID: group.ID, UserID: ownerID, Role: constants.FamilyRoleAdmin}
		if err := s.memberRepo.WithTx(tx).Create(member); err != nil {
			return fmt.Errorf("create family member: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_FAMILY_GROUP_CREATED, err)
	}
	s.log.InfoContext(ctx, constants.LOG_FAMILY_GROUP_CREATED, "family_id", group.ID, "owner_id", ownerID)
	return group, nil
}

// GetByID 查询家庭组详情。
func (s *FamilyGroupService) GetByID(ctx context.Context, id uint) (*model.FamilyGroup, error) {
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError("家庭组（FamilyGroup）不存在", err)
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_FAMILY_GROUP_CREATED, fmt.Errorf("find family group: %w", err))
	}
	return group, nil
}

// ListMine 列出我加入的家庭组。
func (s *FamilyGroupService) ListMine(ctx context.Context, userID uint) ([]model.FamilyGroup, error) {
	members, err := s.memberRepo.ListByUser(userID)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_FAMILY_GROUP_CREATED, fmt.Errorf("list my family members: %w", err))
	}
	groups := make([]model.FamilyGroup, 0, len(members))
	for _, m := range members {
		groups = append(groups, m.Family)
	}
	return groups, nil
}

// ListMembers 列出家庭成员。
func (s *FamilyGroupService) ListMembers(ctx context.Context, familyID uint) ([]model.FamilyMember, error) {
	return s.memberRepo.ListByFamily(familyID)
}

// InviteMember 邀请成员（家庭 Admin 权限）。
func (s *FamilyGroupService) InviteMember(ctx context.Context, familyID, operatorID, userID uint) (*model.FamilyMember, error) {
	if _, err := s.requireAdmin(ctx, familyID, operatorID); err != nil {
		return nil, err
	}
	existing, err := s.memberRepo.FindByFamilyAndUser(familyID, userID)
	if err == nil && existing != nil {
		return nil, util.ConflictError("该用户已是家庭组（FamilyGroup）成员", errors.New("family member duplicated"))
	}
	if !errors.Is(err, util.ErrNotFound) {
		return nil, util.LogError(s.log, ctx, constants.LOG_FAMILY_MEMBER_INVITED, fmt.Errorf("check member: %w", err))
	}
	member := &model.FamilyMember{FamilyID: familyID, UserID: userID, Role: constants.FamilyRoleMember}
	if err := s.memberRepo.Create(member); err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_FAMILY_MEMBER_INVITED, fmt.Errorf("create member: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_FAMILY_MEMBER_INVITED, "family_id", familyID, "user_id", userID)
	return member, nil
}

// JoinByCode 通过邀请码加入。
func (s *FamilyGroupService) JoinByCode(ctx context.Context, userID uint, inviteCode string) (*model.FamilyGroup, error) {
	group, err := s.groupRepo.FindByInviteCode(inviteCode)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError("邀请码（FamilyGroup.invite_code）无效", err)
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_FAMILY_MEMBER_JOINED, fmt.Errorf("find by invite code: %w", err))
	}
	if _, err := s.memberRepo.FindByFamilyAndUser(group.ID, userID); err == nil {
		return nil, util.ConflictError("您已加入该家庭组（FamilyGroup）", errors.New("member exists"))
	}
	member := &model.FamilyMember{FamilyID: group.ID, UserID: userID, Role: constants.FamilyRoleMember}
	if err := s.memberRepo.Create(member); err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_FAMILY_MEMBER_JOINED, fmt.Errorf("create member: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_FAMILY_MEMBER_JOINED, "family_id", group.ID, "user_id", userID)
	return group, nil
}

// SetMemberRole 设置成员权限（家庭 Admin 权限）。
func (s *FamilyGroupService) SetMemberRole(ctx context.Context, familyID, operatorID, memberID uint, role string) error {
	if _, err := s.requireAdmin(ctx, familyID, operatorID); err != nil {
		return err
	}
	member, err := s.memberRepo.FindByID(memberID)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return util.NotFoundError("家庭成员（FamilyMember）不存在", err)
		}
		return util.LogError(s.log, ctx, constants.LOG_FAMILY_MEMBER_ROLE_CHANGED, fmt.Errorf("find member: %w", err))
	}
	if member.FamilyID != familyID {
		return util.ForbiddenError(constants.MsgNotFamilyMember, errors.New("member not in family"))
	}
	if err := s.memberRepo.UpdateRole(member, role); err != nil {
		return util.LogError(s.log, ctx, constants.LOG_FAMILY_MEMBER_ROLE_CHANGED, fmt.Errorf("update member role: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_FAMILY_MEMBER_ROLE_CHANGED, "family_id", familyID, "member_id", memberID, "role", role)
	return nil
}

// RemoveMember 移除成员。
func (s *FamilyGroupService) RemoveMember(ctx context.Context, familyID, operatorID, memberID uint) error {
	if _, err := s.requireAdmin(ctx, familyID, operatorID); err != nil {
		return err
	}
	member, err := s.memberRepo.FindByID(memberID)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return util.NotFoundError("家庭成员（FamilyMember）不存在", err)
		}
		return util.LogError(s.log, ctx, constants.LOG_FAMILY_MEMBER_REMOVED, fmt.Errorf("find member: %w", err))
	}
	if member.FamilyID != familyID {
		return util.ForbiddenError(constants.MsgNotFamilyMember, errors.New("member not in family"))
	}
	if err := s.memberRepo.Delete(member); err != nil {
		return util.LogError(s.log, ctx, constants.LOG_FAMILY_MEMBER_REMOVED, fmt.Errorf("delete member: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_FAMILY_MEMBER_REMOVED, "family_id", familyID, "member_id", memberID)
	return nil
}

// requireAdmin 校验操作人是家庭 Admin。
func (s *FamilyGroupService) requireAdmin(ctx context.Context, familyID, userID uint) (*model.FamilyMember, error) {
	member, err := s.memberRepo.FindByFamilyAndUser(familyID, userID)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.ForbiddenError(constants.MsgNotFamilyMember, errors.New("user not in family"))
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_FAMILY_MEMBER_ROLE_CHANGED, fmt.Errorf("find member: %w", err))
	}
	if member.Role != constants.FamilyRoleAdmin {
		return nil, util.ForbiddenError("仅家庭管理员（FamilyMember.role=admin）可执行此操作", errors.New("member role not admin"))
	}
	return member, nil
}

// IsMember 校验用户是否为家庭成员。
func (s *FamilyGroupService) IsMember(ctx context.Context, familyID, userID uint) error {
	if _, err := s.memberRepo.FindByFamilyAndUser(familyID, userID); err != nil {
		return util.ForbiddenError(constants.MsgNotFamilyMember, err)
	}
	return nil
}
