// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package sqlstore

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/mattermost/gorp"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/store"
)

type groupMembers []model.GroupMember

func initSqlSupplierGroups(sqlStore SqlStore) {
	for _, db := range sqlStore.GetAllConns() {
		groups := db.AddTableWithName(model.Group{}, "Groups").SetKeys(false, "Id")
		groups.ColMap("Id").SetMaxSize(26)
		groups.ColMap("Name").SetMaxSize(model.GroupNameMaxLength).SetUnique(true)
		groups.ColMap("DisplayName").SetMaxSize(model.GroupDisplayNameMaxLength)
		groups.ColMap("Description").SetMaxSize(model.GroupDescriptionMaxLength)
		groups.ColMap("Type").SetMaxSize(model.GroupTypeMaxLength)
		groups.ColMap("TypeProps").SetMaxSize(model.GroupTypePropsMaxLength)

		groupMembers := db.AddTableWithName(model.GroupMember{}, "GroupMembers").SetKeys(false, "GroupId", "UserId")
		groupMembers.ColMap("GroupId").SetMaxSize(26)
		groupMembers.ColMap("UserId").SetMaxSize(26)

		groupTeams := db.AddTableWithName(model.GroupTeam{}, "GroupTeams").SetKeys(false, "GroupId", "TeamId")
		groupTeams.ColMap("GroupId").SetMaxSize(26)
		groupTeams.ColMap("TeamId").SetMaxSize(26)

		groupChannels := db.AddTableWithName(model.GroupChannel{}, "GroupChannels").SetKeys(false, "GroupId", "ChannelId")
		groupChannels.ColMap("GroupId").SetMaxSize(26)
		groupChannels.ColMap("ChannelId").SetMaxSize(26)
	}
}

func (s *SqlSupplier) GroupSave(ctx context.Context, group *model.Group, hints ...store.LayeredStoreHint) *store.LayeredStoreSupplierResult {
	result := store.NewSupplierResult()
	var err error

	if len(group.Id) == 0 {
		if err := group.IsValidForCreate(); err != nil {
			result.Err = err
			return result
		}

		var transaction *gorp.Transaction

		if transaction, err = s.GetMaster().Begin(); err != nil {
			result.Err = model.NewAppError("SqlGroupStore.GroupSave", "store.sql_group.save.open_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
			return result
		}

		result = s.createGroup(ctx, group, transaction, hints...)

		if result.Err != nil {
			transaction.Rollback()
		} else if err := transaction.Commit(); err != nil {
			result.Err = model.NewAppError("SqlGroupStore.GroupSave", "store.sql_group.save_group.commit_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	} else {
		if err := group.IsValidForUpdate(); err != nil {
			result.Err = err
			return result
		}

		group.UpdateAt = model.GetMillis()

		if rowsChanged, err := s.GetMaster().Update(group); err != nil {
			result.Err = model.NewAppError("SqlGroupStore.Save", "store.sql_group.save.update.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else if rowsChanged != 1 {
			result.Err = model.NewAppError("SqlGroupStore.Save", "store.sql_group.save.update.app_error", nil, "no record to update", http.StatusInternalServerError)
		}

		result.Data = group
	}

	return result
}

func (s *SqlSupplier) createGroup(ctx context.Context, group *model.Group, transaction *gorp.Transaction, hints ...store.LayeredStoreHint) *store.LayeredStoreSupplierResult {
	result := store.NewSupplierResult()

	if err := group.IsValidForCreate(); err != nil {
		result.Err = err
		return result
	}

	group.Id = model.NewId()
	group.CreateAt = model.GetMillis()
	group.UpdateAt = group.CreateAt

	if err := transaction.Insert(group); err != nil {
		result.Err = model.NewAppError("SqlGroupStore.Save", "store.sql_group.save.insert.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	result.Data = group

	return result
}

func (s *SqlSupplier) GroupGet(ctx context.Context, groupId string, hints ...store.LayeredStoreHint) *store.LayeredStoreSupplierResult {
	result := store.NewSupplierResult()

	var group model.Group

	if err := s.GetReplica().SelectOne(&group, "SELECT * from Groups WHERE Id = :Id", map[string]interface{}{"Id": groupId}); err != nil {
		if err == sql.ErrNoRows {
			result.Err = model.NewAppError("SqlGroupStore.Get", "store.sql_group.get.app_error", nil, "Id="+groupId+", "+err.Error(), http.StatusNotFound)
		} else {
			result.Err = model.NewAppError("SqlGroupStore.Get", "store.sql_group.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	result.Data = group

	return result
}

func (s *SqlSupplier) GroupGetAllPage(ctx context.Context, offset int, limit int, hints ...store.LayeredStoreHint) *store.LayeredStoreSupplierResult {
	result := store.NewSupplierResult()

	var groups []*model.Group

	if _, err := s.GetReplica().Select(&groups, "SELECT * from Groups WHERE DeleteAt = 0 ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset", map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
		result.Err = model.NewAppError("SqlGroupStore.Get", "store.sql_group.get.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	result.Data = groups

	return result
}

func (s *SqlSupplier) GroupDelete(ctx context.Context, groupId string, hints ...store.LayeredStoreHint) *store.LayeredStoreSupplierResult {
	result := store.NewSupplierResult()

	var group *model.Group
	if err := s.GetReplica().SelectOne(&group, "SELECT * from Groups WHERE Id = :Id", map[string]interface{}{"Id": groupId}); err != nil {
		if err == sql.ErrNoRows {
			result.Err = model.NewAppError("SqlGroupStore.Delete", "store.sql_group.get.app_error", nil, "Id="+groupId+", "+err.Error(), http.StatusNotFound)
		} else {
			result.Err = model.NewAppError("SqlGroupStore.Delete", "store.sql_group.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}

		return result
	}

	time := model.GetMillis()
	group.DeleteAt = time
	group.UpdateAt = time

	if rowsChanged, err := s.GetMaster().Update(group); err != nil {
		result.Err = model.NewAppError("SqlGroupStore.Delete", "store.sql_group.delete.update.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else if rowsChanged != 1 {
		result.Err = model.NewAppError("SqlGroupStore.Delete", "store.sql_group.delete.update.app_error", nil, "no record to update", http.StatusInternalServerError)
	} else {
		result.Data = group
	}

	return result
}

func (s *SqlSupplier) GroupAddMember(member *model.GroupMember) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if result.Err = member.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(member); err != nil {
			if IsUniqueConstraintError(err, []string{"GroupId", "groupmembers_pkey", "PRIMARY"}) {
				result.Err = model.NewAppError("SqlGroupStore.SaveMember", "store.sql_group.save_member.exists.app_error", nil, "group_id="+member.GroupId+", user_id="+member.UserId+", "+err.Error(), http.StatusBadRequest)
				return
			}
			result.Err = model.NewAppError("SqlGroupStore.SaveMember", "store.sql_group.save_member.save.app_error", nil, "group_id="+member.GroupId+", user_id="+member.UserId+", "+err.Error(), http.StatusInternalServerError)
			return
		}

		var retrievedMember *model.GroupMember
		if err := s.GetMaster().SelectOne(&retrievedMember, "SELECT * FROM GroupMembers WHERE GroupId = :GroupId AND UserId = :UserId", map[string]interface{}{"GroupId": member.GroupId, "UserId": member.UserId}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlGroupStore.SaveMember", "store.sql_group.get_member.missing.app_error", nil, "group_id="+member.GroupId+"user_id="+member.UserId+","+err.Error(), http.StatusNotFound)
				return
			}
			result.Err = model.NewAppError("SqlGroupStore.SaveMember", "store.sql_group.get_member.app_error", nil, "group_id="+member.GroupId+"user_id="+member.UserId+","+err.Error(), http.StatusInternalServerError)
			return
		}
		result.Data = retrievedMember
	})
}

// // TODO: set the DeleteAt field
// func (s *SqlSupplier) GroupRemoveMember(teamId string, userId string) store.StoreChannel {
// 	return store.Do(func(result *store.StoreResult) {
// 		_, err := s.GetMaster().Exec("DELETE FROM GroupMembers WHERE GroupId = :GroupId AND UserId = :UserId", map[string]interface{}{"GroupId": teamId, "UserId": userId})
// 		if err != nil {
// 			result.Err = model.NewAppError("SqlChannelStore.RemoveMember", "store.sql_group.remove_member.app_error", nil, "group_id="+teamId+", user_id="+userId+", "+err.Error(), http.StatusInternalServerError)
// 		}
// 	})
// }

// func (s SqlGroupStore) UpdateMember(member *model.GroupMember) store.StoreChannel {
// 	return store.Do(func(result *store.StoreResult) {
// 		member.PreUpdate()

// 		if result.Err = member.IsValid(); result.Err != nil {
// 			return
// 		}

// 		if _, err := s.GetMaster().Update(NewTeamMemberFromModel(member)); err != nil {
// 			result.Err = model.NewAppError("SqlGroupStore.UpdateMember", "store.sql_group.save_member.save.app_error", nil, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		var retrievedMember groupMembers
// 		if err := s.GetMaster().SelectOne(&retrievedMember, TEAM_MEMBERS_WITH_SCHEME_SELECT_QUERY+"WHERE GroupMembers.GroupId = :GroupId AND GroupMembers.UserId = :UserId", map[string]interface{}{"GroupId": member.GroupId, "UserId": member.UserId}); err != nil {
// 			if err == sql.ErrNoRows {
// 				result.Err = model.NewAppError("SqlGroupStore.UpdateMember", "store.sql_group.get_member.missing.app_error", nil, "group_id="+member.GroupId+"user_id="+member.UserId+","+err.Error(), http.StatusNotFound)
// 				return
// 			}
// 			result.Err = model.NewAppError("SqlGroupStore.UpdateMember", "store.sql_group.get_member.app_error", nil, "group_id="+member.GroupId+"user_id="+member.UserId+","+err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		result.Data = retrievedMember.ToModel()
// 	})
// }

// func (s *SqlSupplier) GroupGetMember(teamId string, userId string) store.StoreChannel {
// 	return store.Do(func(result *store.StoreResult) {
// 		var dbMember groupMembers
// 		err := s.GetReplica().SelectOne(&dbMember, TEAM_MEMBERS_WITH_SCHEME_SELECT_QUERY+"WHERE GroupMembers.GroupId = :GroupId AND GroupMembers.UserId = :UserId", map[string]interface{}{"GroupId": teamId, "UserId": userId})
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				result.Err = model.NewAppError("SqlGroupStore.GetMember", "store.sql_group.get_member.missing.app_error", nil, "teamId="+teamId+" userId="+userId+" "+err.Error(), http.StatusNotFound)
// 				return
// 			}
// 			result.Err = model.NewAppError("SqlGroupStore.GetMember", "store.sql_group.get_member.app_error", nil, "teamId="+teamId+" userId="+userId+" "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		result.Data = dbMember.ToModel()
// 	})
// }

// func (s *SqlSupplier) GroupGetMembers(teamId string, offset int, limit int) store.StoreChannel {
// 	return store.Do(func(result *store.StoreResult) {
// 		var dbMembers teamMemberWithSchemeRolesList
// 		_, err := s.GetReplica().Select(&dbMembers, TEAM_MEMBERS_WITH_SCHEME_SELECT_QUERY+"WHERE GroupMembers.GroupId = :GroupId AND GroupMembers.DeleteAt = 0 LIMIT :Limit OFFSET :Offset", map[string]interface{}{"GroupId": teamId, "Limit": limit, "Offset": offset})
// 		if err != nil {
// 			result.Err = model.NewAppError("SqlGroupStore.GetMembers", "store.sql_group.get_members.app_error", nil, "teamId="+teamId+" "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		result.Data = dbMembers.ToModel()
// 	})
// }
