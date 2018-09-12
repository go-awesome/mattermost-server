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

func initSqlSupplierGroups(sqlStore SqlStore) {
	kibibyte := 1024
	for _, db := range sqlStore.GetAllConns() {
		groups := db.AddTableWithName(model.Group{}, "Groups").SetKeys(false, "Id")
		groups.ColMap("Id").SetMaxSize(26)
		groups.ColMap("Name").SetMaxSize(64).SetUnique(true)
		groups.ColMap("DisplayName").SetMaxSize(128)
		groups.ColMap("Description").SetMaxSize(kibibyte)
		groups.ColMap("Type").SetMaxSize(64)
		groups.ColMap("TypeProps").SetMaxSize(64 * kibibyte)

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
	validationErr := model.NewAppError("SqlGroupStore.GroupSave", "store.sql_group.save.invalid_group.app_error", nil, "", http.StatusBadRequest)

	if len(group.Id) == 0 {
		if !group.IsValidForCreate() {
			result.Err = validationErr
			return result
		}

		var transaction *gorp.Transaction
		var err error

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
		if !group.IsValidForUpdate() {
			result.Err = validationErr
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

	if !group.IsValidForCreate() {
		result.Err = model.NewAppError("SqlGroupStore.GroupSave", "store.sql_group.save.invalid_group.app_error", nil, "", http.StatusBadRequest)
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

func (s *SqlSupplier) GroupGetByName(ctx context.Context, name string, hints ...store.LayeredStoreHint) *store.LayeredStoreSupplierResult {
	result := store.NewSupplierResult()

	var group model.Group

	if err := s.GetReplica().SelectOne(&group, "SELECT * from Groups WHERE Name = :Name", map[string]interface{}{"Name": name}); err != nil {
		if err == sql.ErrNoRows {
			result.Err = model.NewAppError("SqlGroupStore.GetByName", "store.sql_group.get_by_name.app_error", nil, "name="+name+",err="+err.Error(), http.StatusNotFound)
		} else {
			result.Err = model.NewAppError("SqlGroupStore.GetByName", "store.sql_group.get_by_name.app_error", nil, "name="+name+",err="+err.Error(), http.StatusInternalServerError)
		}
	}

	result.Data = group

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
