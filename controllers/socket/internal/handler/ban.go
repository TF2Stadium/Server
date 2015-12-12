package handler

import (
	"time"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

func newBan(steamid string, bantype models.PlayerBanType, until int64, reason string) *helpers.TPError {
	player, tperr := models.GetPlayerBySteamId(steamid)
	if tperr != nil {
		return tperr
	}

	time := time.Unix(time.Now().Unix()+until, 0)
	err := player.BanUntil(time, bantype, reason)

	if err != nil {
		tperr = helpers.NewTPErrorFromError(err)
	}

	return nil
}

func unban(steamid string, bantype models.PlayerBanType) *helpers.TPError {
	player, tperr := models.GetPlayerBySteamId(steamid)
	if tperr != nil {
		return tperr
	}

	err := player.Unban(bantype)
	if err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	return nil
}

func InitializeBans(server *wsevent.Server) {
	bans := []struct {
		eventName string
		action    authority.AuthAction
		banType   models.PlayerBanType
	}{
		{"banJoin", helpers.ActionBanJoin, models.PlayerBanJoin},
		{"banCreate", helpers.ActionBanCreate, models.PlayerBanCreate},
		{"banChat", helpers.ActionBanChat, models.PlayerBanChat},
	}

	for _, ban := range bans {
		server.On(ban.eventName, func(_ *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
			reqerr := chelpers.FilterRequest(so, ban.action, true)
			if reqerr != nil {
				return reqerr
			}

			var args struct {
				SteamID *string `json:"steamid"`
				Until   *int64  `json:"until"`
				Reason  *string `json:"reason"`
			}

			if err := chelpers.GetParams(data, &args); err != nil {
				return helpers.NewTPErrorFromError(err)
			}

			tperr := newBan(*args.SteamID, ban.banType, *args.Until, *args.Reason)
			if tperr != nil {
				return tperr
			}

			return chelpers.EmptySuccessJS
		})

		server.On("Un"+ban.eventName, func(_ *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
			reqerr := chelpers.FilterRequest(so, ban.action, true)
			if reqerr != nil {
				return reqerr
			}

			var args struct {
				SteamID *string `json:"steamid"`
			}

			if err := chelpers.GetParams(data, &args); err != nil {
				return helpers.NewTPErrorFromError(err)
			}

			err := unban(*args.SteamID, ban.banType)
			if err != nil {
				return err
			}
			return chelpers.EmptySuccessJS
		})
	}
}
