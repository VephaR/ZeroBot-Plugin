// Package bilibili 图片渲染工具
package bilibili

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/sirupsen/logrus"
)

// 日志实例重命名，避免与包内其他 log 冲突
var renderLog = logrus.WithField("module", "bilibili-render-utils")

// 全局渲染配置（统一风格，供 card2msg.go 复用）
const (
	RenderWidth     = 1200                  // 图片固定宽度
	TitleFontSize   = 36.0                  // 标题字体大小
	SubTitleSize    = 28.0                  // 副标题字体大小
	ContentFontSize = 22.0                  // 内容字体大小
	SmallFontSize   = 18.0                  // 小字体大小
	Padding         = 40.0                  // 内边距
	LineHeight      = 1.5                   // 行高倍数
	TitleColor      = "#1E293B"             // 标题颜色（深灰蓝）
	SubTitleColor   = "#334155"             // 副标题颜色
	ContentColor    = "#475569"             // 内容颜色
	HighlightColor  = "#2563EB"             // 强调色（蓝色）
	DataColor       = "#F97316"             // 数据颜色（橙色）
	BgColor         = "#FFFFFF"             // 背景色（白色）
	BorderColor     = "#F1F5F9"             // 边框/分割线颜色
	MaxContentLines = 5                     // 长文本最大显示行数
	MaxImages       = 3                     // 多图最大显示数量
)

// 全局字体缓存（避免重复加载）
var (
	titleFont   []byte
	contentFont []byte
)

func init() {
	// 预加载字体（复用现有字体文件）
	var err error
	titleFont, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		renderLog.Errorln("[bilibili-render] 加载标题字体失败:", err)
		titleFont = []byte{}
	}
	contentFont, err = file.GetLazyData(text.FontFile, control.Md5File, true)
	if err != nil {
		renderLog.Errorln("[bilibili-render] 加载内容字体失败:", err)
		contentFont = []byte{}
	}
}

// NewRenderContext 初始化渲染上下文（自动计算高度）
func NewRenderContext(minHeight float64) *gg.Context {
	ctx := gg.NewContext(RenderWidth, int(minHeight))
	ctx.SetColor(colorHex(BgColor))
	ctx.Clear()
	// 绘制轻微阴影边框
	ctx.SetColor(colorHex(BorderColor))
	ctx.DrawRoundedRectangle(10, 10, RenderWidth-20, minHeight-20, 20)
	ctx.Fill()
	// 绘制白色背景
	ctx.SetColor(colorHex(BgColor))
	ctx.DrawRoundedRectangle(20, 20, RenderWidth-40, minHeight-40, 15)
	ctx.Fill()
	return ctx
}

// LoadFont 加载字体（自动降级：移除 text.DefaultFontData，用空字节降级）
func LoadFont(ctx *gg.Context, fontSize float64, isBold bool) error {
	fontData := contentFont
	if isBold && len(titleFont) > 0 {
		fontData = titleFont
	}
	if len(fontData) == 0 {
		// 字体加载失败时，使用 gg 内置默认字体（空字节降级）
		return ctx.LoadFontFace("", fontSize)
	}
	return ctx.ParseFontFace(fontData, fontSize)
}

// WrapText 文本自动换行
func WrapText(ctx *gg.Context, text string, maxWidth float64) []string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t'
	})
	var lines []string
	currentLine := ""
	lineWidth := 0.0

	for _, word := range words {
		wordWidth, _ := ctx.MeasureString(word)
		if currentLine == "" {
			currentLine = word
			lineWidth = wordWidth
		} else if lineWidth+wordWidth+10 < maxWidth { // 10为单词间距
			currentLine += " " + word
			lineWidth += wordWidth + 10
		} else {
			lines = append(lines, currentLine)
			currentLine = word
			lineWidth = wordWidth
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

// DrawImageWithLimit 绘制图片（限制尺寸+居中）
func DrawImageWithLimit(ctx *gg.Context, imgURL string, maxWidth, maxHeight float64) (image.Image, error) {
	data, err := web.GetData(imgURL)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	// 按比例缩放
	return imgfactory.Size(img, int(maxWidth), int(maxHeight)).Image(), nil
}

// DrawMultiImages 绘制多张图片（横向拼接）
func DrawMultiImages(ctx *gg.Context, imgURLs []string, startX, startY, maxHeight float64) (float64, error) {
	if len(imgURLs) == 0 {
		return startY, nil
	}
	// 计算单张图片宽度（均分）
	imgCount := len(imgURLs)
	if imgCount > MaxImages {
		imgCount = MaxImages
	}
	singleWidth := (RenderWidth - 2*Padding - float64(imgCount-1)*10) / float64(imgCount)

	currentX := startX
	maxImgHeight := 0.0
	for i := 0; i < imgCount; i++ {
		img, err := DrawImageWithLimit(ctx, imgURLs[i], singleWidth, maxHeight)
		if err != nil {
			renderLog.Warnln("[bilibili-render] 加载图片失败:", err)
			continue
		}
		imgBounds := img.Bounds()
		imgWidth := float64(imgBounds.Max.X - imgBounds.Min.X)
		imgHeight := float64(imgBounds.Max.Y - imgBounds.Min.Y)
		// 居中绘制
		ctx.DrawImage(img, int(currentX), int(startY+(maxHeight-imgHeight)/2))
		currentX += imgWidth + 10 // 10为图片间距
		if imgHeight > maxImgHeight {
			maxImgHeight = imgHeight
		}
	}
	// 绘制图片数量提示（超过限制时）
	if len(imgURLs) > MaxImages {
		ctx.SetColor(colorHex(SubTitleColor))
		if err := LoadFont(ctx, SmallFontSize, false); err != nil {
			return startY + maxImgHeight + 10, err
		}
		tip := fmt.Sprintf("共%d张图片，仅展示前%d张", len(imgURLs), MaxImages)
		tipWidth, _ := ctx.MeasureString(tip)
		ctx.DrawString(tip, (RenderWidth-tipWidth)/2, startY+maxImgHeight+30)
		return startY + maxImgHeight + 40, nil
	}
	return startY + maxImgHeight + 20, nil
}

// colorHex 十六进制颜色转color.RGBA
func colorHex(hex string) color.RGBA {
	var r, g, b, a uint8
	if len(hex) == 7 {
		fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
		a = 255
	} else if len(hex) == 9 {
		fmt.Sscanf(hex, "#%02x%02x%02x%02x", &r, &g, &b, &a)
	}
	return color.RGBA{R: r, G: g, B: b, A: a}
}

// TruncateText 长文本截断（带省略号）
func TruncateText(lines []string, maxLines int) []string {
	if len(lines) <= maxLines {
		return lines
	}
	truncated := lines[:maxLines-1]
	truncated = append(truncated, "...（查看更多请访问原链接）")
	return truncated
}
