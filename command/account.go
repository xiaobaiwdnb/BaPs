package command

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/gucooing/BaPs/common/code"
	"github.com/gucooing/BaPs/db"
	"github.com/gucooing/BaPs/sdk"
	"github.com/gucooing/cdq"
)

var (
	login         = "login"
	ban           = "ban"
	unban         = "unban"
	getAccount    = "getAccount"
	showsetCode   = "showsetCode" // 完全替换 setCode
	fuckgetCode   = "fuckgetCode" // 完全替换 getCode
	delCode       = "delCode"
)

const (
	accountLoginErr    = -1
	accountUnknown     = -2
	accountDbErr       = -3
	accountMarshalErr  = -4
	accountTypeUnknown = -5
	accountCodeErr     = -6
	accountSetCodeErr  = -7
)

func (c *Command) ApplicationCommandAccount() {
	account := &cdq.Command{
		Name:        "account",
		AliasList:   []string{"ac"},
		Description: "操作玩家SDK账户",
		Permissions: cdq.User,
		Options: []*cdq.CommandOption{
			{
				Name:        "account",
				Description: "账户昵称",
				Required:    true,
				Alias:       "a",
			},
			{
				Name:        "type",
				Description: "操作类型",
				Required:    true,
				Alias:       "t",
				ExpectedS:   []string{login, ban, unban, getAccount, showsetCode, fuckgetCode, delCode}, // 仅使用新名称
			},
			{
				Name:        "banMsg",
				Description: "封禁原因",
				Required:    false,
				Default:     "默认封禁",
				Alias:       "b",
			},
			{
				Name:        "code",
				Alias:       "cd",
				Description: "添加指定的邮箱验证码",
			},
		},
		Handlers: cdq.AddHandlers(c.account),
	}

	c.C.ApplicationCommand(account)
}

func (c *Command) account(ctx *cdq.Context) {
	account := ctx.GetFlags().String("account")
	types := ctx.GetFlags().String("type")
	switch types {
	case login:
		ya, err := sdk.GetORAddYostarAccount(account)
		if err != nil || ya.YostarAccount != account {
			ctx.Return(accountLoginErr, fmt.Sprintf("账户注册失败 Account:%s", account))
			return
		}
		ctx.Return(cdq.ApiCodeOk, fmt.Sprintf("账户注册成功 Account:%s", account))
	case ban:
		yul := sdk.GetYostarUserLoginByAccount(account)
		if yul == nil {
			ctx.Return(accountUnknown, fmt.Sprintf("账户不存在 Account:%s", account))
			return
		}
		yul.Ban = true
		yul.BanMsg = ctx.GetFlags().String("banMsg")
		if db.GetDBGame().UpdateYoStarUserLogin(yul) != nil {
			ctx.Return(accountDbErr, fmt.Sprintf("数据库操作失败 Account:%s", account))
			return
		}
		ctx.Return(cdq.ApiCodeOk, fmt.Sprintf("ban Account:%s up!", account))
	case unban:
		yul := sdk.GetYostarUserLoginByAccount(account)
		if yul == nil {
			ctx.Return(accountUnknown, fmt.Sprintf("账户不存在 Account:%s", account))
			return
		}
		yul.Ban = false
		yul.BanMsg = ""
		if db.GetDBGame().UpdateYoStarUserLogin(yul) != nil {
			ctx.Return(accountDbErr, fmt.Sprintf("数据库操作失败 Account:%s", account))
			return
		}
		ctx.Return(cdq.ApiCodeOk, fmt.Sprintf("ban Account:%s up!", account))
	case getAccount:
		yul := sdk.GetYostarUserLoginByAccount(account)
		if yul == nil {
			ctx.Return(accountUnknown, fmt.Sprintf("账户不存在 Account:%s", account))
			return
		}
		str, err := sonic.MarshalString(&AccountInfo{
			Account:         account,
			AccountServerId: yul.AccountServerId,
			YostarUid:       yul.YostarUid,
			Ban:             yul.Ban,
			BanMsg:          yul.BanMsg,
		})
		if err != nil {
			ctx.Return(accountMarshalErr, fmt.Sprintf("账号信息序列化失败 Account:%s", account))
			return
		}
		ctx.Return(cdq.ApiCodeOk, str)
	case fuckgetCode: // 完全替换为 fuckgetCode
		if codeInfo := code.GetCodeInfo(account); codeInfo != nil {
			ctx.Return(cdq.ApiCodeOk, fmt.Sprintf("Account:%s Code:%v", account, codeInfo.Code))
		} else {
			ctx.Return(accountCodeErr, fmt.Sprintf("Account:%s 验证码已过期或失效", account))
		}
	case showsetCode: // 完全替换为 showsetCode
		cd := ctx.GetFlags().Int32("code")
		if err := code.SetCode(account, cd); err == nil {
			ctx.Return(cdq.ApiCodeOk, fmt.Sprintf("Account:%s Code:%v 设置成功", account, cd))
		} else {
			ctx.Return(accountSetCodeErr, fmt.Sprintf("Account:%s Code:%v 设置Code失败原因:%s", account, cd, err.Error()))
		}
	case delCode:
		code.DelCode(account)
		ctx.Return(cdq.ApiCodeOk, fmt.Sprintf("Account:%s 删除Code成功", account))
	default:
		ctx.Return(accountTypeUnknown, fmt.Sprintf("未知的操作类型 Type:%s", types))
	}
}

type AccountInfo struct {
	Account         string `json:"account"`
	AccountServerId int64  `json:"account_server_id"`
	YostarUid       int64  `json:"yostar_uid"`
	Ban             bool   `json:"ban"`
	BanMsg          string `json:"banMsg"`
}
