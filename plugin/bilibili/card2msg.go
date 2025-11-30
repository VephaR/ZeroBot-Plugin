// Package bilibili
package bilibili

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"time"

	bz "github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 日志实例（复用包内日志）
var log = logrus.WithField("module", "bilibili-card2msg")

// 补全缓存路径（与之前逻辑一致）
var cachePath = "./data/bilibili/cache/"

// 初始化缓存目录
func init() {
	if !file.IsExist(cachePath) {
		err := os.MkdirAll(cachePath, 0755)
		if err != nil {
			log.Errorln("创建缓存目录失败:", err)
		}
	}
}

// 渲染相关常量（仅保留本文件独有的，其余复用 render_utils.go）
const (
	coverMaxHeight = 400.0 // 封面图最大高度
	dataItemWidth  = 150.0 // 数据项宽度
)

// ------------------------------ 视频信息+总结渲染 ------------------------------
func RenderVideoCard(card bz.Card, summaryMsg []message.Segment) (imgData []byte, err error) {
	// 缓存文件名：BV号+日期
	cacheKey := fmt.Sprintf("video_%s_%s.png", card.BvID, time.Now().Format("20060102"))
	cacheFile := cachePath + cacheKey
	if file.IsExist(cacheFile) {
		return os.ReadFile(cacheFile)
	}

	// 1. 初始化画布（复用 render_utils.go 的 NewRenderContext）
	minHeight := 800.0
	ctx := NewRenderContext(minHeight)
	defer func() {
		if err == nil {
			// 保存缓存
			f, _ := os.Create(cacheFile)
			defer f.Close()
			imgfactory.WriteTo(ctx.Image(), f)
		}
	}()

	// 2. 绘制标题
	currentY := Padding
	ctx.SetColor(colorHex(TitleColor))
	if err := LoadFont(ctx, TitleFontSize, true); err != nil {
		return nil, err
	}
	titleLines := WrapText(ctx, card.Title, RenderWidth-2*Padding)
	for _, line := range titleLines {
		lineWidth, _ := ctx.MeasureString(line)
		ctx.DrawString(line, (RenderWidth-lineWidth)/2, currentY)
		currentY += TitleFontSize * LineHeight
	}
	currentY += 20

	// 3. 绘制封面图（复用 render_utils.go 的 DrawImageWithLimit）
	coverImg, err := DrawImageWithLimit(ctx, card.Pic, RenderWidth-2*Padding, coverMaxHeight)
	if err == nil {
		coverBounds := coverImg.Bounds()
		coverHeight := float64(coverBounds.Max.Y - coverBounds.Min.Y)
		ctx.DrawImage(coverImg, int(Padding), int(currentY))
		currentY += coverHeight + 30
	}

	// 4. 绘制UP主信息和数据统计
	ctx.SetColor(colorHex(SubTitleColor))
	if err := LoadFont(ctx, SubTitleSize, false); err != nil {
		return nil, err
	}
	// UP主信息
	upText := fmt.Sprintf("UP主：%s", card.Owner.Name)
	if card.Rights.IsCooperation == 1 {
		upText = "联合创作："
		for i, staff := range card.Staff {
			if i > 0 {
				upText += " | "
			}
			upText += fmt.Sprintf("%s（%s）", staff.Name, staff.Title)
		}
	}
	ctx.DrawString(upText, Padding, currentY)
	currentY += SubTitleSize * LineHeight

	// 数据统计（播放、点赞、投币等）
	dataItems := []struct {
		label string
		value string
	}{
		{"播放", bz.HumanNum(card.Stat.View)},
		{"弹幕", bz.HumanNum(card.Stat.Danmaku)},
		{"点赞", bz.HumanNum(card.Stat.Like)},
		{"投币", bz.HumanNum(card.Stat.Coin)},
		{"收藏", bz.HumanNum(card.Stat.Favorite)},
		{"分享", bz.HumanNum(card.Stat.Share)},
	}
	ctx.SetColor(colorHex(DataColor))
	currentX := Padding
	for _, item := range dataItems {
		text := fmt.Sprintf("%s：%s", item.label, item.value)
		ctx.DrawString(text, currentX, currentY)
		currentX += dataItemWidth
	}
	currentY += SubTitleSize * LineHeight + 20

	// 5. 绘制简介
	ctx.SetColor(colorHex(ContentColor))
	if err := LoadFont(ctx, ContentFontSize, false); err != nil {
		return nil, err
	}
	ctx.DrawString("简介：", Padding, currentY)
	currentY += ContentFontSize * LineHeight
	introLines := WrapText(ctx, card.Desc, RenderWidth-2*Padding-20)
	introLines = TruncateText(introLines, MaxContentLines)
	for _, line := range introLines {
		ctx.DrawString(line, Padding+20, currentY)
		currentY += ContentFontSize * LineHeight
	}
	currentY += 30

	// 6. 绘制AI总结（如果有）
	if len(summaryMsg) > 0 {
		ctx.SetColor(colorHex(HighlightColor))
		if err := LoadFont(ctx, SubTitleSize, true); err != nil {
			return nil, err
		}
		ctx.DrawString("视频总结：", Padding, currentY)
		currentY += SubTitleSize * LineHeight

		// 解析总结文本（从summaryMsg中提取）
		summaryText := ""
		for _, seg := range summaryMsg {
			if seg.Type == "text" {
				// 直接使用string类型，无需类型断言
				summaryText += seg.Data["text"]
			}
		}
		// 拆分总结内容（概述+大纲）
		summaryParts := strings.Split(summaryText, "\n\n")
		if len(summaryParts) > 0 {
			// 绘制概述
			ctx.SetColor(colorHex(ContentColor))
			if err := LoadFont(ctx, ContentFontSize, false); err != nil {
				return nil, err
			}
			overviewLines := WrapText(ctx, summaryParts[0], RenderWidth-2*Padding-20)
			for _, line := range overviewLines {
				ctx.DrawString(line, Padding+20, currentY)
				currentY += ContentFontSize * LineHeight
			}
			currentY += 20

			// 绘制大纲
			ctx.SetColor(colorHex(HighlightColor))
			if err := LoadFont(ctx, SubTitleSize, false); err != nil {
				return nil, err
			}
			ctx.DrawString("关键大纲：", Padding+20, currentY)
			currentY += SubTitleSize * LineHeight

			for i := 1; i < len(summaryParts); i++ {
				part := strings.TrimSpace(summaryParts[i])
				if part == "" {
					continue
				}
				// 匹配大纲项（● 标题 + 时间戳内容）
				outlineLines := WrapText(ctx, part, RenderWidth-2*Padding-40)
				for _, line := range outlineLines {
					ctx.SetColor(colorHex(ContentColor))
					if strings.HasPrefix(line, "●") {
						ctx.SetColor(colorHex(HighlightColor))
					} else if strings.Contains(line, ":") {
						// 时间戳部分用数据色
						ctx.SetColor(colorHex(DataColor))
					}
					ctx.DrawString(line, Padding+40, currentY)
					currentY += ContentFontSize * LineHeight
				}
			}
		}
		currentY += 30
	}

	// 7. 绘制视频链接
	ctx.SetColor(colorHex(HighlightColor))
	if err := LoadFont(ctx, SmallFontSize, false); err != nil {
		return nil, err
	}
	linkText := fmt.Sprintf("视频链接：%s%s", bz.VURL, card.BvID)
	linkWidth, _ := ctx.MeasureString(linkText)
	ctx.DrawString(linkText, (RenderWidth-linkWidth)/2, currentY)
	currentY += SmallFontSize * LineHeight + Padding

	// 调整画布高度
	ctx.Scale(1, 1)
	finalImg := ctx.Image().(*image.RGBA)
	finalImg = finalImg.SubImage(image.Rect(0, 0, int(RenderWidth), int(currentY))).(*image.RGBA)

	// 转成字节流
	var buf bytes.Buffer
	if err := imgfactory.WriteTo(finalImg, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ------------------------------ 动态渲染 ------------------------------
func RenderDynamicCard(dynamicCard *bz.DynamicCard) (imgData []byte, err error) {
	// 缓存文件名：动态ID
	cacheKey := fmt.Sprintf("dynamic_%s.png", dynamicCard.Desc.DynamicIDStr)
	cacheFile := cachePath + cacheKey
	if file.IsExist(cacheFile) {
		return os.ReadFile(cacheFile)
	}

	// 1. 解析动态内容
	var card bz.Card
	var vote bz.Vote
	if err := json.Unmarshal(binary.StringToBytes(dynamicCard.Card), &card); err != nil {
		return nil, err
	}
	if dynamicCard.Extension.Vote != "" {
		json.Unmarshal(binary.StringToBytes(dynamicCard.Extension.Vote), &vote)
	}
	cType := dynamicCard.Desc.Type
	dynamicType := msgType[cType]
	if dynamicType == "" {
		dynamicType = fmt.Sprintf("未知类型（%d）", cType)
	}

	// 2. 初始化画布（复用工具函数）
	minHeight := 600.0
	ctx := NewRenderContext(minHeight)
	defer func() {
		if err == nil {
			f, _ := os.Create(cacheFile)
			defer f.Close()
			imgfactory.WriteTo(ctx.Image(), f)
		}
	}()

	currentY := Padding

	// 3. 绘制标题栏（发布者+时间+动态类型）
	ctx.SetColor(colorHex(TitleColor))
	if err := LoadFont(ctx, TitleFontSize, true); err != nil {
		return nil, err
	}
	publisher := card.User.Uname
	if publisher == "" {
		publisher = dynamicCard.Desc.UserProfile.Info.Uname
	}
	titleText := fmt.Sprintf("%s %s", publisher, dynamicType)
	titleWidth, _ := ctx.MeasureString(titleText)
	ctx.DrawString(titleText, (RenderWidth-titleWidth)/2, currentY)
	currentY += TitleFontSize * LineHeight + 10

	// 时间
	ctx.SetColor(colorHex(SubTitleColor))
	if err := LoadFont(ctx, SmallFontSize, false); err != nil {
		return nil, err
	}
	pubTime := time.Unix(int64(dynamicCard.Desc.Timestamp), 0).Format("2006-01-02 15:04:05")
	timeWidth, _ := ctx.MeasureString(pubTime)
	ctx.DrawString(pubTime, (RenderWidth-timeWidth)/2, currentY)
	currentY += SmallFontSize * LineHeight + 30

	// 4. 绘制动态内容
	ctx.SetColor(colorHex(ContentColor))
	if err := LoadFont(ctx, ContentFontSize, false); err != nil {
		return nil, err
	}
	var contentLines []string
	switch cType {
	case 1: // 转发
		contentLines = WrapText(ctx, fmt.Sprintf("转发内容：%s", card.Item.Content), RenderWidth-2*Padding)
	case 2: // 图文
		contentLines = WrapText(ctx, card.Item.Description, RenderWidth-2*Padding)
	case 4: // 无图+投票
		contentLines = WrapText(ctx, card.Item.Content, RenderWidth-2*Padding)
		// 投票内容
		if dynamicCard.Extension.Vote != "" {
			voteText := fmt.Sprintf("\n【投票】%s（截止：%s，参与：%s人）",
				vote.Desc,
				time.Unix(int64(vote.Endtime), 0).Format("2006-01-02"),
				bz.HumanNum(vote.JoinNum))
			contentLines = append(contentLines, WrapText(ctx, voteText, RenderWidth-2*Padding)...)
			for i, opt := range vote.Options {
				optText := fmt.Sprintf("%d. %s", i+1, opt.Desc)
				contentLines = append(contentLines, optText)
			}
		}
	case 8: // 视频
		contentLines = WrapText(ctx, fmt.Sprintf("视频标题：%s\n简介：%s", card.Title, card.Desc), RenderWidth-2*Padding)
	case 16: // 短视频
		contentLines = WrapText(ctx, card.Item.Description, RenderWidth-2*Padding)
	case 64: // 文章
		contentLines = WrapText(ctx, fmt.Sprintf("文章标题：%s\n摘要：%s", card.Title, card.Summary), RenderWidth-2*Padding)
	case 256: // 音频
		contentLines = WrapText(ctx, fmt.Sprintf("音频标题：%s\n简介：%s", card.Title, card.Intro), RenderWidth-2*Padding)
	case 2048: // 简报
		contentLines = WrapText(ctx, fmt.Sprintf("%s\n%s", card.Vest.Content, card.Sketch.DescText), RenderWidth-2*Padding)
	case 4308: // 直播
		contentLines = WrapText(ctx, fmt.Sprintf("直播标题：%s\n分区：%s-%s\n状态：%s",
			card.LivePlayInfo.Title,
			card.LivePlayInfo.ParentAreaName,
			card.LivePlayInfo.AreaName,
			map[int]string{0: "未开播", 1: "直播中"}[card.LivePlayInfo.LiveStatus]), RenderWidth-2*Padding)
	default:
		contentLines = WrapText(ctx, fmt.Sprintf("动态ID：%s", dynamicCard.Desc.DynamicIDStr), RenderWidth-2*Padding)
	}
	contentLines = TruncateText(contentLines, MaxContentLines+3)
	for _, line := range contentLines {
		ctx.DrawString(line, Padding, currentY)
		currentY += ContentFontSize * LineHeight
	}
	currentY += 20

	// 5. 绘制图片（如果有，复用工具函数）
	var imgURLs []string
	switch cType {
	case 2:
		for _, pic := range card.Item.Pictures {
			imgURLs = append(imgURLs, pic.ImgSrc)
		}
	case 8:
		imgURLs = append(imgURLs, card.Pic)
	case 16:
		imgURLs = append(imgURLs, card.Item.Cover.Default)
	case 64:
		imgURLs = card.ImageUrls
	case 256:
		imgURLs = append(imgURLs, card.Cover)
	case 2048:
		imgURLs = append(imgURLs, card.Sketch.CoverURL)
	case 4308:
		imgURLs = append(imgURLs, card.LivePlayInfo.Cover)
	}
	if len(imgURLs) > 0 {
		currentY, err = DrawMultiImages(ctx, imgURLs, Padding, currentY, 300)
		if err != nil {
			log.Warnln("[bilibili-render] 绘制动态图片失败:", err)
		}
	}

	// 6. 绘制链接
	ctx.SetColor(colorHex(HighlightColor))
	if err := LoadFont(ctx, SmallFontSize, false); err != nil {
		return nil, err
	}
	linkText := fmt.Sprintf("动态链接：%s%s", bz.TURL, dynamicCard.Desc.DynamicIDStr)
	linkWidth, _ := ctx.MeasureString(linkText)
	ctx.DrawString(linkText, (RenderWidth-linkWidth)/2, currentY)
	currentY += SmallFontSize * LineHeight + Padding

	// 调整画布高度
	finalImg := ctx.Image().(*image.RGBA)
	finalImg = finalImg.SubImage(image.Rect(0, 0, int(RenderWidth), int(currentY))).(*image.RGBA)

	// 转字节流
	var buf bytes.Buffer
	if err := imgfactory.WriteTo(finalImg, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ------------------------------ 专栏渲染 ------------------------------
func RenderArticleCard(card bz.Card, cvID string) (imgData []byte, err error) {
	cacheKey := fmt.Sprintf("article_%s.png", cvID)
	cacheFile := cachePath + cacheKey
	if file.IsExist(cacheFile) {
		return os.ReadFile(cacheFile)
	}

	// 1. 初始化画布
	ctx := NewRenderContext(800)
	defer func() {
		if err == nil {
			f, _ := os.Create(cacheFile)
			defer f.Close()
			imgfactory.WriteTo(ctx.Image(), f)
		}
	}()

	currentY := Padding

	// 2. 绘制标题
	ctx.SetColor(colorHex(TitleColor))
	if err := LoadFont(ctx, TitleFontSize, true); err != nil {
		return nil, err
	}
	titleLines := WrapText(ctx, card.Title, RenderWidth-2*Padding)
	for _, line := range titleLines {
		lineWidth, _ := ctx.MeasureString(line)
		ctx.DrawString(line, (RenderWidth-lineWidth)/2, currentY)
		currentY += TitleFontSize * LineHeight
	}
	currentY += 20

	// 3. 绘制作者和数据
	ctx.SetColor(colorHex(SubTitleColor))
	if err := LoadFont(ctx, SubTitleSize, false); err != nil {
		return nil, err
	}
	authorText := fmt.Sprintf("作者：%s", card.AuthorName)
	ctx.DrawString(authorText, Padding, currentY)
	// 数据统计
	dataText := fmt.Sprintf("阅读：%s | 评论：%s | 发布时间：%s",
		bz.HumanNum(card.Stats.View),
		bz.HumanNum(card.Stats.Reply),
		time.Unix(int64(card.PublishTime), 0).Format("2006-01-02"))
	dataWidth, _ := ctx.MeasureString(dataText)
	ctx.DrawString(dataText, RenderWidth-Padding-dataWidth, currentY)
	currentY += SubTitleSize * LineHeight + 30

	// 4. 绘制摘要
	ctx.SetColor(colorHex(ContentColor))
	if err := LoadFont(ctx, ContentFontSize, false); err != nil {
		return nil, err
	}
	ctx.DrawString("摘要：", Padding, currentY)
	currentY += ContentFontSize * LineHeight
	summaryLines := WrapText(ctx, card.Summary, RenderWidth-2*Padding-20)
	summaryLines = TruncateText(summaryLines, MaxContentLines)
	for _, line := range summaryLines {
		ctx.DrawString(line, Padding+20, currentY)
		currentY += ContentFontSize * LineHeight
	}
	currentY += 30

	// 5. 绘制配图
	if len(card.OriginImageUrls) > 0 {
		currentY, err = DrawMultiImages(ctx, card.OriginImageUrls, Padding, currentY, 350)
		if err != nil {
			log.Warnln("[bilibili-render] 绘制专栏图片失败:", err)
		}
	}

	// 6. 绘制链接
	ctx.SetColor(colorHex(HighlightColor))
	if err := LoadFont(ctx, SmallFontSize, false); err != nil {
		return nil, err
	}
	linkText := fmt.Sprintf("文章链接：%s%s", bz.CVURL, cvID)
	linkWidth, _ := ctx.MeasureString(linkText)
	ctx.DrawString(linkText, (RenderWidth-linkWidth)/2, currentY)
	currentY += SmallFontSize * LineHeight + Padding

	// 调整画布高度
	finalImg := ctx.Image().(*image.RGBA)
	finalImg = finalImg.SubImage(image.Rect(0, 0, int(RenderWidth), int(currentY))).(*image.RGBA)

	// 转字节流
	var buf bytes.Buffer
	if err := imgfactory.WriteTo(finalImg, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ------------------------------ 直播间渲染 ------------------------------
func RenderLiveCard(card bz.RoomCard) (imgData []byte, err error) {
	cacheKey := fmt.Sprintf("live_%d.png", card.RoomInfo.RoomID)
	cacheFile := cachePath + cacheKey
	if file.IsExist(cacheFile) {
		return os.ReadFile(cacheFile)
	}

	// 1. 初始化画布
	ctx := NewRenderContext(800)
	defer func() {
		if err == nil {
			f, _ := os.Create(cacheFile)
			defer f.Close()
			imgfactory.WriteTo(ctx.Image(), f)
		}
	}()

	currentY := Padding

	// 2. 绘制标题
	ctx.SetColor(colorHex(TitleColor))
	if err := LoadFont(ctx, TitleFontSize, true); err != nil {
		return nil, err
	}
	titleText := fmt.Sprintf("直播间：%s", card.RoomInfo.Title)
	titleWidth, _ := ctx.MeasureString(titleText)
	ctx.DrawString(titleText, (RenderWidth-titleWidth)/2, currentY)
	currentY += TitleFontSize * LineHeight + 20

	// 3. 绘制封面图
	if card.RoomInfo.Keyframe != "" {
		coverImg, err := DrawImageWithLimit(ctx, card.RoomInfo.Keyframe, RenderWidth-2*Padding, coverMaxHeight)
		if err == nil {
			coverBounds := coverImg.Bounds()
			coverHeight := float64(coverBounds.Max.Y - coverBounds.Min.Y)
			ctx.DrawImage(coverImg, int(Padding), int(currentY))
			currentY += coverHeight + 30
		}
	}

	// 4. 绘制主播信息
	ctx.SetColor(colorHex(SubTitleColor))
	if err := LoadFont(ctx, SubTitleSize, false); err != nil {
		return nil, err
	}
	anchorText := fmt.Sprintf("主播：%s", card.AnchorInfo.BaseInfo.Uname)
	ctx.DrawString(anchorText, Padding, currentY)
	currentY += SubTitleSize * LineHeight

	// 直播间信息
	infoLines := []string{
		fmt.Sprintf("房间号：%d（短号：%d）", card.RoomInfo.RoomID, card.RoomInfo.ShortID),
		fmt.Sprintf("分区：%s-%s", card.RoomInfo.ParentAreaName, card.RoomInfo.AreaName),
		fmt.Sprintf("状态：%s", map[int]string{0: "未开播", 1: fmt.Sprintf("直播中（%s人气）", bz.HumanNum(card.RoomInfo.Online))}[card.RoomInfo.LiveStatus]),
	}
	ctx.SetColor(colorHex(ContentColor))
	if err := LoadFont(ctx, ContentFontSize, false); err != nil {
		return nil, err
	}
	for _, line := range infoLines {
		ctx.DrawString(line, Padding+20, currentY)
		currentY += ContentFontSize * LineHeight
	}
	currentY += 30

	// 5. 绘制链接
	ctx.SetColor(colorHex(HighlightColor))
	if err := LoadFont(ctx, SmallFontSize, false); err != nil {
		return nil, err
	}
	linkText := fmt.Sprintf("直播间链接：%s%d", bz.LURL, card.RoomInfo.RoomID)
	linkWidth, _ := ctx.MeasureString(linkText)
	ctx.DrawString(linkText, (RenderWidth-linkWidth)/2, currentY)
	currentY += SmallFontSize * LineHeight + Padding

	// 调整画布高度
	finalImg := ctx.Image().(*image.RGBA)
	finalImg = finalImg.SubImage(image.Rect(0, 0, int(RenderWidth), int(currentY))).(*image.RGBA)

	// 转字节流
	var buf bytes.Buffer
	if err := imgfactory.WriteTo(finalImg, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ------------------------------ 原有函数替换 ------------------------------
// dynamicDetail 用动态id查动态信息（返回图片）
func dynamicDetail(cookiecfg *bz.CookieConfig, dynamicIDStr string) (imgData []byte, err error) {
	dyc, err := bz.GetDynamicDetail(cookiecfg, dynamicIDStr)
	if err != nil {
		return nil, err
	}
	return RenderDynamicCard(&dyc)
}

// articleCard2msg 专栏转图片
func articleCard2msg(card bz.Card, defaultID string) (imgData []byte, err error) {
	return RenderArticleCard(card, defaultID)
}

// liveCard2msg 直播卡片转图片
func liveCard2msg(card bz.RoomCard) (imgData []byte, err error) {
	return RenderLiveCard(card)
}

// videoCard2msg 视频卡片转图片（含总结）
func videoCard2msg(card bz.Card, summaryMsg []message.Segment) (imgData []byte, err error) {
	return RenderVideoCard(card, summaryMsg)
}

// 恢复 msgType 定义（动态类型映射）
var msgType = map[int]string{
	1:    "转发了动态",
	2:    "发布了图文动态",
	4:    "发布了文字动态",
	8:    "投稿了视频",
	16:   "投稿了短视频",
	64:   "投稿了专栏",
	256:  "投稿了音频",
	2048: "发布了简报",
	4200: "发布了直播预告",
	4308: "发布了直播动态",
}
