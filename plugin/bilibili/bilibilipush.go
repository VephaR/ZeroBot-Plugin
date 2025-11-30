// Package bilibili b站推送
package bilibili

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	bz "github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 保留原有常量和变量定义...

// ------------------------------ 修改 sendDynamic ------------------------------
func sendDynamic(ctx *zero.Ctx) error {
	uids := bdb.getAllBuidByDynamic()
	for _, buid := range uids {
		time.Sleep(2 * time.Second)
		cardList, err := getUserDynamicCard(buid, cfg)
		if err != nil {
			return err
		}
		if len(cardList) == 0 {
			return nil
		}
		t, ok := lastTime[buid]
		if !ok {
			lastTime[buid] = cardList[0].Get("desc.timestamp").Int()
			return nil
		}
		for i := len(cardList) - 1; i >= 0; i-- {
			ct := cardList[i].Get("desc.timestamp").Int()
			if ct > t && ct > time.Now().Unix()-600 {
				lastTime[buid] = ct
				m, ok := control.Lookup("bilibilipush")
				if ok {
					groupList := bdb.getAllGroupByBuidAndDynamic(buid)
					dc, err := bz.LoadDynamicDetail(cardList[i].Raw)
					if err != nil {
						return errors.Errorf("动态%v解析失败:%v", cardList[i].Get("desc.dynamic_id_str"), err)
					}
					// 渲染动态为图片
					imgData, err := RenderDynamicCard(&dc)
					if err != nil {
						log.Errorln("[bilibili-push] 动态渲染失败:", err)
						// 降级为文字消息
						msg, _ := oldDynamicCard2msg(&dc)
						for _, gid := range groupList {
							if m.IsEnabledIn(gid) {
								time.Sleep(time.Millisecond * 100)
								switch {
								case gid > 0:
									ctx.SendGroupMessage(gid, msg)
								case gid < 0:
									ctx.SendPrivateMessage(-gid, msg)
								}
							}
						}
						continue
					}
					// 发送图片
					msg := message.ImageBytes(imgData)
					for _, gid := range groupList {
						if m.IsEnabledIn(gid) {
							time.Sleep(time.Millisecond * 100)
							switch {
							case gid > 0:
								ctx.SendGroupMessage(gid, msg)
							case gid < 0:
								ctx.SendPrivateMessage(-gid, msg)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

// ------------------------------ 修改 sendLive ------------------------------
func sendLive(ctx *zero.Ctx) error {
	uids := bdb.getAllBuidByLive()
	ll, err := getLiveList(uids...)
	if err != nil {
		return err
	}
	gjson.Get(ll, "data").ForEach(func(key, value gjson.Result) bool {
		newStatus := int(value.Get("live_status").Int())
		if newStatus == 2 {
			newStatus = 0
		}
		if _, ok := liveStatus[key.Int()]; !ok {
			liveStatus[key.Int()] = newStatus
			return true
		}
		oldStatus := liveStatus[key.Int()]
		if newStatus != oldStatus && newStatus == 1 {
			liveStatus[key.Int()] = newStatus
			m, ok := control.Lookup("bilibilipush")
			if ok {
				groupList := bdb.getAllGroupByBuidAndLive(key.Int())
				roomID := value.Get("short_id").Int()
				if roomID == 0 {
					roomID = value.Get("room_id").Int()
				}
				// 获取直播间详情并渲染图片
				cookie, _ := cfg.Load()
				liveCard, err := bz.GetLiveRoomInfo(strconv.FormatInt(roomID, 10), cookie)
				if err != nil {
					log.Errorln("[bilibili-push] 直播信息获取失败:", err)
					// 降级为文字消息
					oldMsg := []message.Segment{
						message.Text(value.Get("uname").String() + " 正在直播：\n"),
						message.Text(value.Get("title").String()),
						message.Image(value.Get("cover_from_user").String()),
						message.Text("直播链接：", bz.LiveURL+strconv.FormatInt(roomID, 10)),
					}
					for _, gid := range groupList {
						if m.IsEnabledIn(gid) {
							time.Sleep(time.Millisecond * 100)
							switch {
							case gid > 0:
								if res := bdb.getAtAll(gid); res == 1 {
									oldMsg = append([]message.Segment{message.AtAll()}, oldMsg...)
								}
								ctx.SendGroupMessage(gid, oldMsg)
							case gid < 0:
								ctx.SendPrivateMessage(-gid, oldMsg)
							}
						}
					}
					return true
				}
				// 渲染直播图片
				imgData, err := RenderLiveCard(liveCard)
				if err != nil {
					log.Errorln("[bilibili-push] 直播渲染失败:", err)
					// 降级为文字消息
					oldMsg := []message.Segment{
						message.Text(value.Get("uname").String() + " 正在直播：\n"),
						message.Text(value.Get("title").String()),
						message.Image(value.Get("cover_from_user").String()),
						message.Text("直播链接：", bz.LiveURL+strconv.FormatInt(roomID, 10)),
					}
					for _, gid := range groupList {
						if m.IsEnabledIn(gid) {
							time.Sleep(time.Millisecond * 100)
							switch {
							case gid > 0:
								if res := bdb.getAtAll(gid); res == 1 {
									oldMsg = append([]message.Segment{message.AtAll()}, oldMsg...)
								}
								ctx.SendGroupMessage(gid, oldMsg)
							case gid < 0:
								ctx.SendPrivateMessage(-gid, oldMsg)
							}
						}
					}
					return true
				}
				// 发送直播图片
				msg := message.ImageBytes(imgData)
				for _, gid := range groupList {
					if m.IsEnabledIn(gid) {
						time.Sleep(time.Millisecond * 100)
						switch {
						case gid > 0:
							if res := bdb.getAtAll(gid); res == 1 {
								msg = append([]message.Segment{message.AtAll()}, msg)
							}
							ctx.SendGroupMessage(gid, msg)
						case gid < 0:
							ctx.SendPrivateMessage(-gid, msg)
						}
					}
				}
			}
		} else if newStatus != oldStatus {
			liveStatus[key.Int()] = newStatus
		}
		return true
	})
	return nil
}

// 保留原有其他函数...
