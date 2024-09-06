package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin
	configurationLock sync.RWMutex
	configuration     *configuration
}

func (p *Plugin) UserHasJoined(c *plugin.Context, user *model.User) {
    config := p.getConfiguration()

    if user.Props == nil {
        user.Props = make(model.StringMap)
    }

    if config.DisableDMOnJoin {
        user.Props["disable_direct_message"] = "true"
    }

    if config.DisableGMOnJoin {
        user.Props["disable_group_message"] = "true"
    }

    _, err := p.API.UpdateUser(user)
    if err != nil {
        p.API.LogError("Failed to update user", "error", err.Error())
    }
}

func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {

    channel, appErr := p.API.GetChannel(post.ChannelId)
    if appErr != nil {
        return nil, fmt.Sprintf("Failed to get channel: %v", appErr)
    }

    sender, appErr := p.API.GetUser(post.UserId)
    if appErr != nil {
        return nil, fmt.Sprintf("Failed to get sender: %v", appErr)
    }

    if p.isUserExcluded(sender.Username) {
        return post, ""
    }

    if channel.Type == model.ChannelTypeDirect {
        if !p.canUserSendDM(sender) {
            return nil, "You are not allowed to send direct messages."
        }

        otherUserId := p.getOtherUserInDM(channel, sender.Id)
        recipient, appErr := p.API.GetUser(otherUserId)
        if appErr != nil {
            return nil, fmt.Sprintf("Failed to get recipient: %v", appErr)
        }

        if !p.canUserReceiveDM(recipient) {
            return nil, "The recipient is not allowed to receive direct messages."
        }
    } else if channel.Type == model.ChannelTypeGroup {
        if !p.canUserSendGM(sender) {
            return nil, "You are not allowed to participate in group messages."
        }
    }

    return post, ""
}

func (p *Plugin) isUserExcluded(username string) bool {
	config := p.getConfiguration()
	for _, excludedUser := range config.ExcludedUsers {
		if username == excludedUser {
			return true
		}
	}
	return false
}

func (p *Plugin) canUserSendDM(user *model.User) bool {
	config := p.getConfiguration()
	if config.DisableDMForExistingUser {
		return p.isUserAllowed(user)
	}
	return user.Props["disable_direct_message"] != "true" || p.isUserAllowed(user)
}

func (p *Plugin) canUserReceiveDM(user *model.User) bool {
	return p.canUserSendDM(user)
}

func (p *Plugin) canUserSendGM(user *model.User) bool {
	config := p.getConfiguration()
	if config.DisableGMForExistingUser {
		return p.isUserAllowed(user)
	}
	return user.Props["disable_group_message"] != "true" || p.isUserAllowed(user)
}

func (p *Plugin) isUserAllowed(user *model.User) bool {
	config := p.getConfiguration()
	for _, role := range config.AllowedRoles {
		if user.IsInRole(role) {
			return true
		}
	}
	return false
}

func (p *Plugin) getOtherUserInDM(channel *model.Channel, userId string) string {
    if channel.Type != model.ChannelTypeDirect {
        return ""
    }

    users, err := p.API.GetUsersInChannel(channel.Id, "", 0, 2)
    if err != nil {
        p.API.LogError("Failed to get users in channel", "error", err.Error())
        return ""
    }

    for _, user := range users {
        if user.Id != userId {
            return user.Id
        }
    }

    return ""
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/restrictions":
		p.handleGetRestrictions(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) handleGetRestrictions(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	user, err := p.API.GetUser(userID)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	restrictions := struct {
		CanSendDMs                 bool `json:"canSendDMs"`
		CanReceiveDMs              bool `json:"canReceiveDMs"`
		CanParticipateInGroupChats bool `json:"canParticipateInGroupChats"`
	}{
		CanSendDMs:                 p.canUserSendDM(user),
		CanReceiveDMs:              p.canUserReceiveDM(user),
		CanParticipateInGroupChats: p.canUserSendGM(user),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restrictions)
}
