// Package bilibili bilibili卡片解析
package bilibili

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	bz "github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/pkg/errors"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	enableHex            = 0x10
	bilibiliparseReferer = "https://www.bilibili.com"
	ua                   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36" // 补充缺失的ua定义
)

// 保留原有变量定义...

func init() {
	// 保留原有初始化逻辑...
	// 仅修改开关逻辑（默认开启视频总结，按之前的方案）
	en.OnRegex(`^(开启|打开|启用|关闭|关掉|禁用)视频总结$`, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid <= 0 {
				gid = -ctx.Event.UserID
			}
			option := ctx.State["regex_matched"].([]string)[1]
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if !ok {
				ctx.SendChain(message.Text("找不到服务!"))
				return
			}
			var data int64
			switch option {
			case "开启", "打开", "启用":
				data = enableHex
			case "关闭", "关掉", "禁用":
				data = 0x2 // 手动关闭标记
			default:
				return
			}
			err := c.SetData(gid, data)
			if err != nil {
				ctx.SendChain(message.Text("出错啦: ", err))
				return
			}
			ctx.SendChain(message.Text("已", option, "视频总结"))
		})
	// 保留原有OnRegex注册...
}

// ------------------------------ 修改 handleVideo ------------------------------
func handleVideo(ctx *zero.Ctx) {
	id := ctx.State["regex_matched"].([]string)[1]
	if id == "" {
		id = ctx.State["regex_matched"].([]string)[2]
	}
	card, err := bz.GetVideoInfo(id)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}

	// 1. 获取AI总结
	var summaryMsg []message.Segment
	c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if ok {
		data := c.GetData(ctx.Event.GroupID)
		if data == 0 || data == enableHex { // 默认开启/手动开启
			sm, err := getVideoSummary(cfg, card)
			if err != nil {
				summaryMsg = append(summaryMsg, message.Text("ERROR: 视频总结生成失败 - ", err))
			} else {
				summaryMsg = sm
			}
		}
	}

	// 2. 渲染视频信息+总结为图片
	imgData, err := videoCard2msg(card, summaryMsg)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: 图片渲染失败 - ", err))
		// 降级为文字消息（保留原有逻辑）
		oldMsg, _ := oldVideoCard2msg(card) // 新增临时降级函数
		ctx.SendChain(oldMsg...)
		if len(summaryMsg) > 0 {
			ctx.SendChain(summaryMsg...)
		}
	} else {
		ctx.SendChain(message.ImageBytes(imgData))
	}

	// 3. 发送下载的视频
	downLoadMsg, err := getVideoDownload(cfg, card, cachePath)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(downLoadMsg...)
}

// ------------------------------ 修改其他handle函数 ------------------------------
func handleDynamic(ctx *zero.Ctx) {
	dynamicID := ctx.State["regex_matched"].([]string)[2]
	imgData, err := dynamicDetail(cfg, dynamicID)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		// 降级为文字消息
		dyc, _ := bz.GetDynamicDetail(cfg, dynamicID)
		oldMsg, _ := oldDynamicCard2msg(&dyc)
		ctx.SendChain(oldMsg...)
		return
	}
	ctx.SendChain(message.ImageBytes(imgData))
}

func handleArticle(ctx *zero.Ctx) {
	cvID := ctx.State["regex_matched"].([]string)[1]
	card, err := bz.GetArticleInfo(cvID)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	imgData, err := articleCard2msg(card, cvID)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: 图片渲染失败 - ", err))
		// 降级为文字消息
		oldMsg := oldArticleCard2msg(card, cvID)
		ctx.SendChain(oldMsg...)
		return
	}
	ctx.SendChain(message.ImageBytes(imgData))
}

func handleLive(ctx *zero.Ctx) {
	roomID := ctx.State["regex_matched"].([]string)[1]
	cookie, err := cfg.Load()
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	card, err := bz.GetLiveRoomInfo(roomID, cookie)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	imgData, err := liveCard2msg(card)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: 图片渲染失败 - ", err))
		// 降级为文字消息
		oldMsg := oldLiveCard2msg(card)
		ctx.SendChain(oldMsg...)
		return
	}
	ctx.SendChain(message.ImageBytes(imgData))
}

// ------------------------------ 新增降级用的旧版文字转换函数 ------------------------------
// oldVideoCard2msg 原文字转换函数（降级用）
func oldVideoCard2msg(card bz.Card) (msg []message.Segment, err error) {
	var mCard bz.MemberCard
	msg = make([]message.Segment, 0, 16)
	mCard, err = bz.GetMemberCard(card.Owner.Mid)
	msg = append(msg, message.Text("标题: ", card.Title, "\n"))
	if card.Rights.IsCooperation == 1 {
		for i := 0; i < len(card.Staff); i++ {
			msg = append(msg, message.Text(card.Staff[i].Title, ": ", card.Staff[i].Name, " 粉丝: ", bz.HumanNum(card.Staff[i].Follower), "\n"))
		}
	} else {
		if err != nil {
			msg = append(msg, message.Text("UP主: ", card.Owner.Name, "\n"))
		} else {
			msg = append(msg, message.Text("UP主: ", card.Owner.Name, " 粉丝: ", bz.HumanNum(mCard.Fans), "\n"))
		}
	}
	msg = append(msg, message.Image(card.Pic))
	msg = append(msg, message.Text("👀播放: ", bz.HumanNum(card.Stat.View), " 💬弹幕: ", bz.HumanNum(card.Stat.Danmaku),
		"\n👍点赞: ", bz.HumanNum(card.Stat.Like), " 💰投币: ", bz.HumanNum(card.Stat.Coin),
		"\n📁收藏: ", bz.HumanNum(card.Stat.Favorite), " 🔗分享: ", bz.HumanNum(card.Stat.Share),
		"\n📝简介: ", card.Desc, "\n", bz.VURL, card.BvID, "\n\n"))
	return
}

// 其他旧版函数（oldDynamicCard2msg、oldArticleCard2msg、oldLiveCard2msg）
// 直接复制原有 card2msg.go 中的对应函数，前缀改为 old，返回 []message.Segment
func oldDynamicCard2msg(dynamicCard *bz.DynamicCard) (msg []message.Segment, err error) {
	// 复制原有 dynamicCard2msg 函数逻辑
}
func oldArticleCard2msg(card bz.Card, defaultID string) []message.Segment {
	// 复制原有 articleCard2msg 函数逻辑
}
func oldLiveCard2msg(card bz.RoomCard) []message.Segment {
	// 复制原有 liveCard2msg 函数逻辑
}

// 保留原有 getVideoSummary 和 getVideoDownload 函数...
