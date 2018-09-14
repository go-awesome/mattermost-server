// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package storetest

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/store"
)

func TestGroupStore(t *testing.T, ss store.Store) {
	t.Run("Save", func(t *testing.T) { testGroupStoreSave(t, ss) })
	t.Run("Get", func(t *testing.T) { testGroupStoreGet(t, ss) })
	// t.Run("Delete", func(t *testing.T) { testGroupStoreDelete(t, ss) })
}

func testGroupStoreSave(t *testing.T, ss store.Store) {
	// Save a new group.
	g1 := &model.Group{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Type:        model.GroupTypeLdap,
		Description: model.NewId(),
		TypeProps:   model.NewId(),
	}

	// Happy path
	res1 := <-ss.Group().Save(g1)
	assert.Nil(t, res1.Err)
	d1 := res1.Data.(*model.Group)
	assert.Len(t, d1.Id, 26)
	assert.Equal(t, g1.Name, d1.Name)
	assert.Equal(t, g1.DisplayName, d1.DisplayName)
	assert.Equal(t, g1.Description, d1.Description)
	assert.Equal(t, g1.TypeProps, d1.TypeProps)

	// Requires name and display name
	g2 := &model.Group{
		Name:        "",
		DisplayName: model.NewId(),
		Type:        model.GroupTypeLdap,
	}
	res2 := <-ss.Group().Save(g2)
	assert.Nil(t, res2.Data)
	assert.NotNil(t, res2.Err)

	g2.Name = model.NewId()
	g2.DisplayName = ""
	res3 := <-ss.Group().Save(g2)
	assert.Nil(t, res3.Data)
	assert.NotNil(t, res3.Err)

	// Can't invent an ID and save it.
	g3 := &model.Group{
		Id:          model.NewId(),
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Type:        model.GroupTypeLdap,
	}
	res4 := <-ss.Group().Save(g3)
	assert.Nil(t, res4.Data)
	assert.NotNil(t, res4.Err)

	// Won't accept a duplicate name.
	g4 := &model.Group{
		Name:        g1.Name,
		DisplayName: model.NewId(),
	}
	res5 := <-ss.Group().Save(g4)
	assert.Nil(t, res5.Data)
	assert.NotNil(t, res5.Err)

	// Fields cannot be greater than max values
	g5 := &model.Group{
		Name:        strings.Repeat("x", model.GroupNameMaxLength),
		DisplayName: strings.Repeat("x", model.GroupDisplayNameMaxLength),
		Description: strings.Repeat("x", model.GroupDescriptionMaxLength),
		TypeProps:   strings.Repeat("x", model.GroupTypePropsMaxLength),
		Type:        model.GroupTypeLdap,
	}
	assert.Nil(t, g5.IsValidForCreate())

	g5.Name = g5.Name + "x"
	assert.NotNil(t, g5.IsValidForCreate())
	g5.Name = model.NewId()
	assert.Nil(t, g5.IsValidForCreate())

	g5.DisplayName = g5.DisplayName + "x"
	assert.NotNil(t, g5.IsValidForCreate())
	g5.DisplayName = model.NewId()
	assert.Nil(t, g5.IsValidForCreate())

	g5.Description = g5.Description + "x"
	assert.NotNil(t, g5.IsValidForCreate())
	g5.Description = model.NewId()
	assert.Nil(t, g5.IsValidForCreate())

	g5.TypeProps = g5.TypeProps + "x"
	assert.NotNil(t, g5.IsValidForCreate())
	g5.TypeProps = model.NewId()
	assert.Nil(t, g5.IsValidForCreate())

	// Must use a valid type
	g6 := &model.Group{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Description: model.NewId(),
		TypeProps:   model.NewId(),
		Type:        "fake",
	}
	assert.NotNil(t, g6.IsValidForCreate())
}

func testGroupStoreGet(t *testing.T, ss store.Store) {
	// Create a group
	g1 := &model.Group{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Description: model.NewId(),
		Type:        model.GroupTypeLdap,
		TypeProps:   model.NewId(),
	}
	res1 := <-ss.Group().Save(g1)
	assert.Nil(t, res1.Err)
	d1 := res1.Data.(*model.Group)
	assert.Len(t, d1.Id, 26)

	// Get the group
	res2 := <-ss.Group().Get(d1.Id)
	assert.Nil(t, res2.Err)
	d2 := res2.Data.(*model.Group)
	assert.Equal(t, d1.Id, d2.Id)
	assert.Equal(t, d1.Name, d2.Name)
	assert.Equal(t, d1.DisplayName, d2.DisplayName)
	assert.Equal(t, d1.Description, d2.Description)
	assert.Equal(t, d1.TypeProps, d2.TypeProps)
	assert.Equal(t, d1.CreateAt, d2.CreateAt)
	assert.Equal(t, d1.UpdateAt, d2.UpdateAt)
	assert.Equal(t, d1.DeleteAt, d2.DeleteAt)

	// Get an invalid group
	res3 := <-ss.Group().Get(model.NewId())
	assert.NotNil(t, res3.Err)
}

// func testGroupStoreDelete(t *testing.T, ss store.Store) {
// 	// Save a group to test with.
// 	g1 := &model.Group{
// 		Name:        model.NewId(),
// 		DisplayName: model.NewId(),
// 		Description: model.NewId(),
// 		Permissions: []string{
// 			"invite_user",
// 			"create_public_channel",
// 			"add_user_to_team",
// 		},
// 		SchemeManaged: false,
// 	}

// 	res1 := <-ss.Group().Save(g1)
// 	assert.Nil(t, res1.Err)
// 	d1 := res1.Data.(*model.Group)
// 	assert.Len(t, d1.Id, 26)

// 	// Check the group is there.
// 	res2 := <-ss.Group().Get(d1.Id)
// 	assert.Nil(t, res2.Err)

// 	// Delete the group.
// 	res3 := <-ss.Group().Delete(d1.Id)
// 	assert.Nil(t, res3.Err)

// 	// Check the group is deleted there.
// 	res4 := <-ss.Group().Get(d1.Id)
// 	assert.Nil(t, res4.Err)
// 	d2 := res4.Data.(*model.Group)
// 	assert.NotZero(t, d2.DeleteAt)

// 	res5 := <-ss.Group().GetByName(d1.Name)
// 	assert.Nil(t, res5.Err)
// 	d3 := res5.Data.(*model.Group)
// 	assert.NotZero(t, d3.DeleteAt)

// 	// Try and delete a group that does not exist.
// 	res6 := <-ss.Group().Delete(model.NewId())
// 	assert.NotNil(t, res6.Err)
// }
