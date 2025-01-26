package common

import (
	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"errors"
	"fmt"
	"strconv"

	"github.com/sadeepa24/connected_bot/botapi"
	C "github.com/sadeepa24/connected_bot/constbot"
	"github.com/sadeepa24/connected_bot/controller"
	tgbotapi "github.com/sadeepa24/connected_bot/tgbotapi"
	"github.com/sadeepa24/connected_bot/update"
	"go.uber.org/zap"
)

type Sendreciver func(msg any) (*tgbotapi.Message, error)
type Callbackreciver func(msg any, btns *botapi.Buttons) (*tgbotapi.CallbackQuery, error)
type Alertsender func(msg string)

type Tgcalls struct {
	Callbackreciver Callbackreciver
	Sendreciver     Sendreciver
	Alertsender     Alertsender
}

type OptionExcutors struct {
	//Common
	Tgcalls
	Upx             *update.Updatectx
	Btns            *botapi.Buttons
	Usersession     *controller.CtrlSession
	MessageSession  *botapi.Msgsession
	Ctrl            *controller.Controller
	Logger 			*zap.Logger

	//For Exec Rule addr
}

type Initer interface {
	Init() 
}

func ReciveString(call Tgcalls) (string, error) {
	var(
		replymeassage *tgbotapi.Message
		err error
		confName string
	) 

	for {

		if replymeassage, err = call.Sendreciver(nil); err != nil {
			return "", err
		}
		if replymeassage.IsCommand() {
			call.Alertsender("Send Valid String Not Commands")
			continue
		}
		confName = replymeassage.Text
		if replymeassage.Text == "" {
			confName = "noname"
		}

		break

	}

	return confName, nil
	
}

func ReciveInt(call Tgcalls, max, min int) (int, error) {
	var (
		retry int
		replymeassage *tgbotapi.Message
		err error
		out int
	)
	for {
		
		if retry > 5 {
			call.Alertsender(C.GetMsg(C.Msgretryfail))
			return 0, errors.New("retry attemps failed")
		}
		if replymeassage, err = call.Sendreciver(nil); err != nil {
			return 0, err
		}
		if replymeassage == nil {
			continue
		}
		if out, err = strconv.Atoi(replymeassage.Text); err != nil {
			call.Alertsender(C.GetMsg(C.MsgValidInt))
			continue
		}
		 if out > max || out <= min {
			call.Alertsender(fmt.Sprintf("int should be between %d, and %d", min, max))
			continue
		 }

		break

	}
	
	return out, nil
}

//recived as GB,
//max, min should be in byte format,
//return as GB,
func ReciveBandwidth(call Tgcalls, max, min C.Bwidth) (C.Bwidth, error) {
	bwith := C.Bwidth(0)
	for {
		bth, err := ReciveInt(call, 100000, 0)
		if err != nil {
			return 0, err
		}
		bwith = C.Bwidth(bth)
		if bwith.GbtoByte() > max || bwith.GbtoByte() <= min {
			call.Alertsender(fmt.Sprintf("Bandwidth should be between %d, and %d", min, max))
			continue
		}
		break

	}

	return bwith, nil

}
