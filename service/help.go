package service

import (
	"context"
	"strconv"
	"time"

	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
	"github.com/sadeepa24/connected_bot/update/bottype"
)

const (
	sthomehelp   = 0
	sthelpclosed = 1
	sthelpabout  = 3
	stgototpage  = 4
)

type HelpState struct {
	State          int
	Messagesession *botapi.Msgsession
	btns           *botapi.Buttons
	wiz            *Usersrv
	upx            *update.Updatectx
	ctx            context.Context
	Page           int
	MaxPages       int
	PageName       string

	helperinfo bottype.HelpCommandInfo
}

func (h *HelpState) home() error {
	h.btns.Reset([]int16{2, 2, 1, 1})
	h.btns.AddBtcommon(C.Btncommand)
	h.btns.AddBtcommon(C.BtnBtinfo)
	h.btns.AddBtcommon(C.BtnBuilderHelp)
	//h.btns.AddBtcommon(C.BtnFaq)
	h.btns.AddBtcommon(C.BtnTutorial)
	h.btns.AddBtcommon(C.BtnAbout)

	h.btns.AddClose(false)

	h.Messagesession.Edit(struct {
		Name     string
		Username string
		TgId     int64
	}{
		Name:     h.upx.User.Name,
		Username: h.upx.User.Tguser.UserName,
		TgId:     h.upx.User.TgID,
	}, h.btns, C.TmpHelpHome)

	var (
		callback *tgbotapi.CallbackQuery
		err      error
	)

	if callback, err = h.wiz.callback.GetcallbackContext(h.ctx, h.btns.ID()); err != nil {
		return err
	}
	switch callback.Data {
	case C.BtnClose:
		h.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgHeloClosed), false)
		h.State = sthelpclosed
		return nil

	case C.BtnFaq:
		h.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgCallbackFaq), true)
		return nil

	case C.Btncommand, C.BtnBtinfo, C.BtnBuilderHelp, C.BtnTutorial:
		if !h.upx.User.Isverified() {
			h.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.Msghelpnoverify), true)
			return nil
		}

		switch callback.Data {
		case C.Btncommand:
			h.PageName = C.TmpHelpCmPage
			h.MaxPages = int(h.helperinfo.CommandPageCount)
		case C.BtnBtinfo:
			h.PageName = C.TmpHelpInfoPage
			h.MaxPages = int(h.helperinfo.InfoPageCount)
		case C.BtnBuilderHelp:
			h.PageName = C.TmplHelpBuilderHelp
			h.MaxPages = int(h.helperinfo.BuilderHelp)
		case C.BtnTutorial:
			h.PageName = C.TmplHelpTuto
			h.MaxPages = int(h.helperinfo.TutorialPageCount)
		}

		h.Page = 1
		h.State = stgototpage

	case C.BtnAbout:
		h.State = sthelpabout

	}
	return nil
}

func (h *HelpState) about() error {
	h.btns.Reset([]int16{2})
	h.btns.AddCloseBack()
	h.Messagesession.Edit(struct {
		*botapi.CommonUser
	}{
		&botapi.CommonUser{
			Name:     h.upx.User.Name,
			Username: h.upx.Chat.UserName,
			TgId:     h.upx.User.TgID,
		},
	}, h.btns, C.TmpAbout)

	var callback *tgbotapi.CallbackQuery
	var err error
	if callback, err = h.wiz.callback.GetcallbackContext(h.ctx, h.btns.ID()); err != nil {
		return err
	}
	switch callback.Data {
	case C.BtnClose:
		h.Messagesession.RemoveBtns()
		h.State = sthelpclosed
	case C.BtnBack:
		h.State = sthomehelp
	}
	return nil
}

func (h *HelpState) gotopage() error {
	h.btns.Reset([]int16{2})
	h.btns.AddBack(false)
	if h.Page != h.MaxPages {
		h.btns.AddBtcommon(C.BtnNext)
	}
	h.btns.AddClose(false)

	h.Messagesession.Edit(struct {
		*botapi.CommonUser
	}{
		&botapi.CommonUser{
			Name:     h.upx.User.Name,
			Username: h.upx.Chat.UserName,
			TgId:     h.upx.User.TgID,
		},
	}, h.btns, h.PageName+strconv.Itoa(h.Page))

	var (
		callback *tgbotapi.CallbackQuery
		err      error
	)

	if callback, err = h.wiz.callback.GetcallbackContext(h.ctx, h.btns.ID()); err != nil {
		return err
	}

	switch callback.Data {
	case C.BtnBack:
		if h.Page == 1 {
			h.State = sthomehelp
			return nil
		}
		h.Page--
	case C.BtnNext:
		h.Page++
	case C.BtnClose:
		//Messagesession.DeleteAllMsg()
		h.Messagesession.RemoveBtns()
		h.Messagesession.Callbackanswere(callback.ID, C.GetMsg(C.MsgHeloClosed), false)
		h.State = sthelpclosed
		return nil
	}
	return nil
}

func (u *Usersrv) commandHelpV2(upx *update.Updatectx) error {
	upx.Ctx, upx.Cancle = context.WithTimeout(u.ctx, 5*time.Minute)
	Messagesession := botapi.NewMsgsession(upx.Ctx, u.botapicaller, upx.User.TgID, upx.User.TgID, upx.User.Lang)
	btns := botapi.NewButtons([]int16{1})

	state := HelpState{
		Messagesession: Messagesession,
		btns:           btns,
		wiz:            u,
		ctx:            upx.Ctx,
		upx:            upx,
		helperinfo:     u.ctrl.GetHelepCmdInfo(),
	}
	var err error

	help:
	for {
		switch state.State {
		case sthomehelp:
			err = state.home()
		case stgototpage:
			err = state.gotopage()
		case sthelpabout:
			err = state.about()
		case sthelpclosed:
			return nil
		default:
			break help
		}
		if err != nil || upx.Ctx.Err() != nil {
			return nil
		}
	}
	return nil
}
