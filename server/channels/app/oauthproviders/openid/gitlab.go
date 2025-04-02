// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package oauthopenid

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
	"github.com/mattermost/mattermost/server/v8/einterfaces"
)

type OpenIdProvider struct {
}

type OpenIdUser struct {
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	DisplayName  string   `json:"display_name"`
	Emails       []string `json:"emails"`
	DefaultEmail string   `json:"default_email"`
	DefaultPhone struct {
		Id     int    `json:"id"`
		Number string `json:"number"`
	} `json:"default_phone"`
	RealName        string `json:"real_name"`
	IsAvatarEmpty   bool   `json:"is_avatar_empty"`
	Birthday        string `json:"birthday"`
	DefaultAvatarId string `json:"default_avatar_id"`
	Login           string `json:"login"`
	OldSocialLogin  string `json:"old_social_login"`
	Sex             string `json:"sex"`
	Id              string `json:"id"`
	ClientId        string `json:"client_id"`
	Psuid           string `json:"psuid"`
}

func init() {
	provider := &OpenIdProvider{}
	einterfaces.RegisterOAuthProvider(model.ServiceOpenid, provider)
}

func userFromYandexUser(logger mlog.LoggerIFace, glu *OpenIdUser) *model.User {
	user := &model.User{}
	splittedMail := strings.Split(glu.Emails[0], "@")[0]
	user.Nickname = glu.RealName
	user.Username = model.CleanUsername(logger, splittedMail)
	user.FirstName = glu.FirstName
	user.LastName = glu.LastName
	user.Email = glu.Emails[0]
	user.Email = strings.ToLower(user.Email)
	userId := glu.getAuthData()
	user.AuthData = &userId
	user.AuthService = model.ServiceOpenid
	user.Props = make(model.StringMap)
	user.Props["avatar_id"] = glu.DefaultAvatarId
	return user
}

func userFromJSON(data io.Reader) (*OpenIdUser, error) {
	decoder := json.NewDecoder(data)
	var glu OpenIdUser
	err := decoder.Decode(&glu)
	if err != nil {
		return nil, err
	}
	return &glu, nil
}

func (glu *OpenIdUser) IsValid() error {
	if glu.Id == "" {
		return errors.New("user id can't be 0")
	}

	if len(glu.Emails) == 0 {
		return errors.New("user e-mail should not be empty")
	}

	return nil
}

func (glu *OpenIdUser) getAuthData() string {
	return glu.Id
}

func (gp *OpenIdProvider) GetUserFromJSON(c request.CTX, data io.Reader, tokenUser *model.User) (*model.User, error) {
	glu, err := userFromJSON(data)
	if err != nil {
		return nil, err
	}
	if err = glu.IsValid(); err != nil {
		return nil, err
	}

	return userFromYandexUser(c.Logger(), glu), nil
}

func (gp *OpenIdProvider) GetSSOSettings(_ request.CTX, config *model.Config, service string) (*model.SSOSettings, error) {
	return &config.OpenIdSettings, nil
}

func (gp *OpenIdProvider) GetUserFromIdToken(_ request.CTX, idToken string) (*model.User, error) {
	return nil, nil
}

func (gp *OpenIdProvider) IsSameUser(_ request.CTX, dbUser, oauthUser *model.User) bool {
	return dbUser.AuthData == oauthUser.AuthData
}
