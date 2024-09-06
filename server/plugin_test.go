package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServeHTTP(t *testing.T) {
    assert := assert.New(t)
    plugin := Plugin{}

    t.Run("NotFound", func(t *testing.T) {
        w := httptest.NewRecorder()
        r := httptest.NewRequest(http.MethodGet, "/not-found", nil)
        plugin.ServeHTTP(nil, w, r)
        assert.Equal(http.StatusNotFound, w.Code)
    })

    t.Run("GetRestrictions", func(t *testing.T) {
        w := httptest.NewRecorder()
        r := httptest.NewRequest(http.MethodGet, "/restrictions", nil)
        r.Header.Set("Mattermost-User-ID", "user1")

        api := &plugintest.API{}
        api.On("GetUser", "user1").Return(&model.User{
            Id:    "user1",
            Roles: "system_user",
        }, nil)
        plugin.SetAPI(api)

        plugin.configuration = &configuration{
            DisableDMForExistingUser:    true,
            DisableGMForExistingUser:    true,
            AllowedRoles:                []string{"system_admin"},
        }

        plugin.ServeHTTP(nil, w, r)

        assert.Equal(http.StatusOK, w.Code)
        var restrictions struct {
            CanSendDMs                 bool `json:"canSendDMs"`
            CanReceiveDMs              bool `json:"canReceiveDMs"`
            CanParticipateInGroupChats bool `json:"canParticipateInGroupChats"`
        }
        err := json.NewDecoder(w.Body).Decode(&restrictions)
        assert.Nil(err)
        assert.False(restrictions.CanSendDMs)
        assert.False(restrictions.CanReceiveDMs)
        assert.False(restrictions.CanParticipateInGroupChats)
    })
}

func TestMessageWillBePosted(t *testing.T) {
    assert := assert.New(t)
    plugin := Plugin{}

    t.Run("DirectMessageAllowed", func(t *testing.T) {
        api := &plugintest.API{}
        api.On("GetChannel", "channel1").Return(&model.Channel{
            Id:   "channel1",
            Type: model.ChannelTypeDirect,
        }, nil)
        api.On("GetUsersInChannel", "channel1", "", 0, 2).Return([]*model.User{
            {Id: "user1", Roles: "system_admin"},
            {Id: "user2", Roles: "system_user"},
        }, nil)
        api.On("GetUser", "user1").Return(&model.User{
            Id:    "user1",
            Roles: "system_admin",
        }, nil)
        api.On("GetUser", "user2").Return(&model.User{
            Id:    "user2",
            Roles: "system_user",
        }, nil)
        plugin.SetAPI(api)

        plugin.configuration = &configuration{
            DisableDMForExistingUser: false,
            AllowedRoles:             []string{"system_admin", "system_user"},
        }

        post := &model.Post{
            UserId:    "user1",
            ChannelId: "channel1",
        }

        resultPost, resultErr := plugin.MessageWillBePosted(nil, post)
        assert.Equal(post, resultPost)
        assert.Empty(resultErr)
    })

    t.Run("DirectMessageBlocked", func(t *testing.T) {
        api := &plugintest.API{}
        api.On("GetChannel", "channel1").Return(&model.Channel{
            Id:   "channel1",
            Type: model.ChannelTypeDirect,
        }, nil)
        api.On("GetUsersInChannel", "channel1", "", 0, 2).Return([]*model.User{
            {Id: "user1", Roles: "system_user"},
            {Id: "user2", Roles: "system_user"},
        }, nil)
        api.On("GetUser", "user1").Return(&model.User{
            Id:    "user1",
            Roles: "system_user",
        }, nil)
        api.On("GetUser", "user2").Return(&model.User{
            Id:    "user2",
            Roles: "system_user",
        }, nil)
        plugin.SetAPI(api)

        plugin.configuration = &configuration{
            DisableDMForExistingUser: true,
            AllowedRoles:             []string{"system_admin"},
        }

        post := &model.Post{
            UserId:    "user1",
            ChannelId: "channel1",
        }

        resultPost, resultErr := plugin.MessageWillBePosted(nil, post)
        assert.Nil(resultPost)
        assert.Equal("You are not allowed to send direct messages.", resultErr)
    })
}

func TestUserHasJoined(t *testing.T) {
    assert := assert.New(t)
    plugin := Plugin{}

    t.Run("DisableDMAndGMOnJoin", func(t *testing.T) {
        api := &plugintest.API{}
        api.On("UpdateUser", mock.AnythingOfType("*model.User")).Return(func(u *model.User) (*model.User, *model.AppError) {
            assert.Equal("true", u.Props["disable_direct_message"])
            assert.Equal("true", u.Props["disable_group_message"])
            return u, nil
        })
        plugin.SetAPI(api)

        plugin.configuration = &configuration{
            DisableDMOnJoin: true,
            DisableGMOnJoin: true,
        }

        user := &model.User{
            Id:    "user1",
            Roles: "system_user",
            Props: make(model.StringMap),
        }

        plugin.UserHasJoined(nil, user)

        api.AssertCalled(t, "UpdateUser", mock.AnythingOfType("*model.User"))
    })

    t.Run("DoNotDisableDMAndGMOnJoin", func(t *testing.T) {
        api := &plugintest.API{}
        api.On("UpdateUser", mock.AnythingOfType("*model.User")).Return(func(u *model.User) (*model.User, *model.AppError) {
            assert.NotContains(u.Props, "disable_direct_message")
            assert.NotContains(u.Props, "disable_group_message")
            return u, nil
        })
        plugin.SetAPI(api)

        plugin.configuration = &configuration{
            DisableDMOnJoin: false,
            DisableGMOnJoin: false,
        }

        user := &model.User{
            Id:    "user2",
            Roles: "system_user",
            Props: make(model.StringMap),
        }

        plugin.UserHasJoined(nil, user)

        api.AssertCalled(t, "UpdateUser", mock.AnythingOfType("*model.User"))
    })
}
