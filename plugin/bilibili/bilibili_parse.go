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
	unableHex            = 0x7fffffff_fffffffd
	bilibiliparseReferer = "https://www.bilibili.com"
)

var (
	limit            = ctxext.NewLimiterManager(time.Second*10, 1)
	searchVideo      = `bilibili.com\\?/video\\?/(?:av(\d+)|([bB][vV][0-9a-zA-Z]+))`
	searchDynamic    = `(t.bilibili.com|m.bilibili.com\\?/dynamic)\\?/(\d+)`
	searchArticle    = `bilibili.com\\?/read\\?/(?:cv|mobile\\?/)(\d+)`
	searchLiveRoom   = `live.bilibili.com\\?/(\d+)`
	searchVideoRe    = regexp.MustCompile(searchVideo)
	searchDynamicRe  = regexp.MustCompile(searchDynamic)
	searchArticleRe  = regexp.MustCompile(searchArticle)
	searchLiveRoomRe = regexp.MustCompile(searchLiveRoom)
	cachePath        string
)

// 插件主体
func init() {
	en := control.Register("bilibiliparse", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "b站链接解析",
		Help:             "例:- t.bilibili.com/642277677329285174\n- bilibili.com/read/cv17134450\n- bilibili.com/video/BV13B4y1x7pS\n- live.bilibili.com/22603245 ",
	})
	cachePath = en.DataFolder() + "cache/"
	_ = os.RemoveAll(cachePath)
	_ = os.MkdirAll(cachePath, 0755)
	en.OnRegex(`((b23|acg).tv|bili2233.cn)\\?/[0-9a-zA-Z]+`).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			u := ctx.State["regex_matched"].([]string)[0]
			u = strings.ReplaceAll(u, "\\", "")
			realurl, err := bz.GetRealURL("https://" + u)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			switch {
			case searchVideoRe.MatchString(realurl):
				ctx.State["regex_matched"] = searchVideoRe.FindStringSubmatch(realurl)
				handleVideo(ctx)
			case searchDynamicRe.MatchString(realurl):
				ctx.State["regex_matched"] = searchDynamicRe.FindStringSubmatch(realurl)
				handleDynamic(ctx)
			case searchArticleRe.MatchString(realurl):
				ctx.State["regex_matched"] = searchArticleRe.FindStringSubmatch(realurl)
				handleArticle(ctx)
			case searchLiveRoomRe.MatchString(realurl):
				ctx.State["regex_matched"] = searchLiveRoomRe.FindStringSubmatch(realurl)
				handleLive(ctx)
			}
		})
	en.OnRegex(`^(开启|打开|启用|关闭|关掉|禁用)视频总结$`, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid <= 0 {
				// 个人用户设为负数
				gid = -ctx.Event.UserID
			}
			option := ctx.State["regex_matched"].([]string)[1]
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if !ok {
				ctx.SendChain(message.Text("找不到服务!"))
				return
			}
			data := c.GetData(ctx.Event.GroupID)
			switch option {
			case "开启", "打开", "启用":
				data |= enableHex
			case "关闭", "关掉", "禁用":
				data &= unableHex
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
	en.OnRegex(searchVideo).SetBlock(true).Limit(limit.LimitByGroup).Handle(handleVideo)
	en.OnRegex(searchDynamic).SetBlock(true).Limit(limit.LimitByGroup).Handle(handleDynamic)
	en.OnRegex(searchArticle).SetBlock(true).Limit(limit.LimitByGroup).Handle(handleArticle)
	en.OnRegex(searchLiveRoom).SetBlock(true).Limit(limit.LimitByGroup).Handle(handleLive)
}

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
	msg, err := videoCard2msg(card)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if ok && c.GetData(ctx.Event.GroupID)&enableHex == enableHex {
		summaryMsg, err := getVideoSummary(cfg, card)
		if err != nil {
			msg = append(msg, message.Text("ERROR: ", err))
		} else {
			msg = append(msg, summaryMsg...)
		}
	}
	ctx.SendChain(msg...)
	downLoadMsg, err := getVideoDownload(cfg, card, cachePath)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(downLoadMsg...)
}

func handleDynamic(ctx *zero.Ctx) {
	msg, err := dynamicDetail(cfg, ctx.State["regex_matched"].([]string)[2])
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(msg...)
}

func handleArticle(ctx *zero.Ctx) {
	card, err := bz.GetArticleInfo(ctx.State["regex_matched"].([]string)[1])
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(articleCard2msg(card, ctx.State["regex_matched"].([]string)[1])...)
}

func handleLive(ctx *zero.Ctx) {
	cookie, err := cfg.Load()
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	card, err := bz.GetLiveRoomInfo(ctx.State["regex_matched"].([]string)[1], cookie)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(liveCard2msg(card)...)
}

// getVideoSummary AI视频总结
func getVideoSummary(cookiecfg *bz.CookieConfig, card bz.Card) (msg []message.Segment, err error) {
	var (
		data         []byte
		videoSummary bz.VideoSummary
	)
	data, err = web.RequestDataWithHeaders(web.NewDefaultClient(), bz.SignURL(fmt.Sprintf(bz.VideoSummaryURL, card.BvID, card.CID, card.Owner.Mid)), "GET", func(req *http.Request) error {
		if cookiecfg != nil {
			cookie := ""
			cookie, err = cookiecfg.Load()
			if err != nil {
				return err
			}
			req.Header.Add("cookie", cookie)
		}
		req.Header.Set("User-Agent", ua)
		return nil
	}, nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &videoSummary)
	msg = make([]message.Segment, 0, 16)
	msg = append(msg, message.Text("已为你生成视频总结\n\n"))
	msg = append(msg, message.Text(videoSummary.Data.ModelResult.Summary, "\n\n"))
	for _, v := range videoSummary.Data.ModelResult.Outline {
		msg = append(msg, message.Text("● ", v.Title, "\n"))
		for _, p := range v.PartOutline {
			msg = append(msg, message.Text(fmt.Sprintf("%d:%d %s\n", p.Timestamp/60, p.Timestamp%60, p.Content)))
		}
		msg = append(msg, message.Text("\n"))
	}
	return
}

func getVideoDownload(cookiecfg *bz.CookieConfig, card bz.Card, cachePath string) (msg []message.Segment, err error) {
	var (
		data          []byte
		videoDownload bz.VideoDownload
		stderr        bytes.Buffer
	)
	today := time.Now().Format("20060102")
	videoFile := fmt.Sprintf("%s%s%s.mp4", cachePath, card.BvID, today)
	if file.IsExist(videoFile) {
		msg = append(msg, message.Video("file:///"+file.BOTPATH+"/"+videoFile))
		return
	}
	data, err = web.RequestDataWithHeaders(web.NewDefaultClient(), bz.SignURL(fmt.Sprintf(bz.VideoDownloadURL, card.BvID, card.CID)), "GET", func(req *http.Request) error {
		if cookiecfg != nil {
			cookie := ""
			cookie, err = cookiecfg.Load()
			if err != nil {
				return err
			}
			req.Header.Add("cookie", cookie)
		}
		req.Header.Set("User-Agent", ua)
		return nil
	}, nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &videoDownload)
	if err != nil {
		return
	}
	headers := fmt.Sprintf("User-Agent: %s\nReferer: %s", ua, bilibiliparseReferer)
	// 限制最多下载8分钟视频
	cmd := exec.Command("ffmpeg", "-ss", "0", "-t", "480", "-headers", headers, "-i", videoDownload.Data.Durl[0].URL, "-c", "copy", videoFile)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		// err = errors.Errorf("未配置ffmpeg，%v", stderr)
		err = errors.Errorf("FFmpeg繁忙，如需下载视频请稍后重试")
		return
	}
	msg = append(msg, message.Video("file:///"+file.BOTPATH+"/"+videoFile))
	return
}
