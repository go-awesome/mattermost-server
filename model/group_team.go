// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

type GroupTeam struct {
	GroupId  string `json:"group_id"`
	TeamId   string `json:"team_id"`
	CanLeave bool   `json:"can_leave"`
	AutoAdd  bool   `json:"auto_add"`
	CreateAt int64  `json:"create_at"`
	DeleteAt int64  `json:"delete_at"`
}
