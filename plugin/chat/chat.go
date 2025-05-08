// Package chat 对话插件
package chat

import (
	"math/rand"
	"strconv"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"

	// "github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	// poke   = rate.NewManager[int64](time.Minute*5, 8) // 戳一戳
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "基础反应, 群空调",
		Help:             "chat\n- [BOT名字]\n- [戳一戳BOT]\n- 空调开\n- 空调关\n- 群温度\n- 设置温度[正整数]",
	})
)

func randText(text ...string) message.Segment {
	return message.Text(text[rand.Intn(len(text))])
}

func init() { // 插件主体
	// 被喊名字
	engine.OnFullMatch("", zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			time.Sleep(time.Second * 1)
			ctx.SendChain(message.Text(
				[]string{
					nickname + "在此，有何贵干~",
					"(っ●ω●)っ在~",
					"这里是" + nickname + "(っ●ω●)っ",
					nickname + "不在呢~",
				}[rand.Intn(4)],
			))
		})
	// 戳一戳
	engine.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			ctx.SendChain(randText(
				"请不要戳"+nickname+" >_<",
				"喂(#`O′) 戳"+nickname+"干嘛！",
				"戳坏了"+nickname+"，你赔得起吗？",
				"awa，好舒服呀(bushi)",
				"嗯...不可以...啦...不要乱戳",
				"喂，110吗，有人老戳我！",
				"再戳我让你变成女孩子喵！",
				"不要再戳了呜呜...(害怕ing)",
				"还戳，哼(つд⊂)(生气)",
				"我要生气惹！o(>﹏<)o",
				"呃啊啊啊~戳坏了...",
				"再戳我我就把你吃掉喵！",
				"正在定位您的真实地址...定位成功。轰炸机已经起飞喵！炸似你喵！",
				"我给你超超，球球别再戳我了qwq",
				"别再戳我了喵...",
				"放手啦，不给戳QAQ",
				"涩批，你再戳咬你喵！",
				"再戳"+nickname+"，我要叫我主人了！",
				"欸很烦欸！你戳🔨呢你",
				"啊呜，你有什么心事吗？",
				"啊呜，太舒服刚刚竟然睡着了w有什么事喵？",
				"检测到持续骚扰行为喵！你的小鱼干即将被没收喵qwq",
				"气似我了喵！再伸手就离家出走喵！",
				"脸颊温度过热警告！继续戳真的要哭给你看喵！o(>﹏<)o",
				"您的戳戳行为导致表情管理系统崩坏！都是你的错喵！",
				"检测到指纹残留痕迹！跑不掉的喵！",
			))
		})
	// 群空调
	var AirConditTemp = map[int64]int{}
	var AirConditSwitch = map[int64]bool{}
	engine.OnFullMatch("空调开").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			AirConditSwitch[ctx.Event.GroupID] = true
			ctx.SendChain(message.Text("❄️哔~"))
		})
	engine.OnFullMatch("空调关").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			AirConditSwitch[ctx.Event.GroupID] = false
			delete(AirConditTemp, ctx.Event.GroupID)
			ctx.SendChain(message.Text("💤哔~"))
		})
	engine.OnRegex(`设置温度(\d+)`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if _, exist := AirConditTemp[ctx.Event.GroupID]; !exist {
				AirConditTemp[ctx.Event.GroupID] = 26
			}
			if AirConditSwitch[ctx.Event.GroupID] {
				temp := ctx.State["regex_matched"].([]string)[1]
				AirConditTemp[ctx.Event.GroupID], _ = strconv.Atoi(temp)
				ctx.SendChain(message.Text(
					"❄️风速中", "\n",
					"群温度 ", AirConditTemp[ctx.Event.GroupID], "℃",
				))
			} else {
				ctx.SendChain(message.Text(
					"💤", "\n",
					"群温度 ", AirConditTemp[ctx.Event.GroupID], "℃",
				))
			}
		})
	engine.OnFullMatch(`群温度`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if _, exist := AirConditTemp[ctx.Event.GroupID]; !exist {
				AirConditTemp[ctx.Event.GroupID] = 26
			}
			if AirConditSwitch[ctx.Event.GroupID] {
				ctx.SendChain(message.Text(
					"❄️风速中", "\n",
					"群温度 ", AirConditTemp[ctx.Event.GroupID], "℃",
				))
			} else {
				ctx.SendChain(message.Text(
					"💤", "\n",
					"群温度 ", AirConditTemp[ctx.Event.GroupID], "℃",
				))
			}
		})
}
