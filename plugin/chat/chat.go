// Package chat 对话插件
package chat

import (
	"math/rand"
	"strconv"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	poke   = rate.NewManager[int64](time.Minute*5, 8) // 戳一戳
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "基础反应, 群空调",
		Help:             "chat\n- [BOT名字]\n- [戳一戳BOT]\n- 空调开\n- 空调关\n- 群温度\n- 设置温度[正整数]",
	})
)

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
	// // 戳一戳
	// engine.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).
	// 	Handle(func(ctx *zero.Ctx) {
	// 		var nickname = zero.BotConfig.NickName[0]
	// 		switch {
	// 		case poke.Load(ctx.Event.GroupID).AcquireN(2):
	// 			// 5分钟共100块命令牌 一次消耗2块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("请不要戳", nickname, " >_<"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("喂(#`O′) 戳", nickname, "干嘛！"))
	// 		case poke.Load(ctx.Event.GroupID).AcquireN(2):
	// 			// 5分钟共100块命令牌 一次消耗2块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("戳坏了", nickname, "，你赔得起吗？"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("awa，好舒服呀(bushi)"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("嗯……不可以……啦……不要乱戳"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("喂，110吗，有人老戳我"))
	// 		case poke.Load(ctx.Event.GroupID).AcquireN(2):
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("再戳我让你变成女孩子喵！"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("不要再戳了呜呜……(害怕ing)"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("还戳，哼(つд⊂)(生气)"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("我要生气惹！o(>﹏<)o"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("呃啊啊啊~戳坏了……"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("再戳我我就把你吃掉喵！"))
	// 		case poke.Load(ctx.Event.GroupID).AcquireN(2):
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("正在定位您的真实地址……定位成功。轰炸机已经起飞喵！炸似你喵！"))
	// 		case poke.Load(ctx.Event.GroupID).AcquireN(2):
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("我给你超超，球球别再戳我了qwq"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("别再戳我了喵……"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("放手啦，不给戳QAQ"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("涩批，你再戳咬你喵！"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("再戳", nickname, "，我要叫我主人了"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("欸很烦欸！你戳🔨呢你"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("啊呜，你有什么心事吗？"))
	// 		case poke.Load(ctx.Event.GroupID).Acquire():
	// 			// 5分钟共100块命令牌 一次消耗1块命令牌
	// 			time.Sleep(time.Second * 1)
	// 			ctx.SendChain(message.Text("啊呜，太舒服刚刚竟然睡着了w 有什么事喵？"))
	// 		default:
	// 			// 频繁触发，不回复
	// 		}
	// 	})
	// 戳一戳
	engine.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			rand.Seed(time.Now().UnixNano())
			r := rand.Intn(20)
			if r == 0 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("请不要戳", nickname, " >_<"))
			} else if r == 1 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("喂(#`O′) 戳", nickname, "干嘛！"))
			} else if r == 2 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("戳坏了", nickname, "，你赔得起吗？"))
			} else if r == 3 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("awa，好舒服呀(bushi)"))
			} else if r == 4 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("嗯……不可以……啦……不要乱戳"))
			} else if r == 5 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("喂，110吗，有人老戳我"))
			} else if r == 6 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("再戳我让你变成女孩子喵！"))
			} else if r == 7 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("不要再戳了呜呜……(害怕ing)"))
			} else if r == 8 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("还戳，哼(つд⊂)(生气)"))
			} else if r == 9 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("我要生气惹！o(>﹏<)o"))
			} else if r == 10 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("呃啊啊啊~戳坏了……"))
			} else if r == 11 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("再戳我我就把你吃掉喵！"))
			} else if r == 12 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("正在定位您的真实地址……定位成功。轰炸机已经起飞喵！炸似你喵！"))
			} else if r == 13 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("我给你超超，球球别再戳我了qwq"))
			} else if r == 14 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("别再戳我了喵……"))
			} else if r == 15 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("放手啦，不给戳QAQ"))
			} else if r == 16 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("涩批，你再戳咬你喵！"))
			} else if r == 17 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("再戳", nickname, "，我要叫我主人了"))
			} else if r == 18 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("欸很烦欸！你戳🔨呢你"))
			} else if r == 19 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("啊呜，你有什么心事吗？"))
			} else if r == 20 {
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("啊呜，太舒服刚刚竟然睡着了w 有什么事喵？"))
			}
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
