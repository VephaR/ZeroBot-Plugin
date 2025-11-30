package bilibili

import (
	"database/sql"
	"errors"
	"fmt"
	"strings" // 移到这里！所有import必须集中在文件开头
	"time"

	_ "github.com/go-sql-driver/mysql" // 示例数据库驱动，根据实际调整
)

// 补全缺失的变量：数据库连接实例（bdb）
var bdb *sql.DB

// 补全缺失的变量：上次拉取动态的时间
var lastTime time.Time

// 动态卡片结构体（根据实际需求调整字段）
type DynamicCard struct {
	ID          string
	Content     string
	ImageURLs   []string
	PublishedAt time.Time
	LikeCount   int64
}

// 初始化数据库连接（程序启动时调用）
func InitDB(dsn string) error {
	var err error
	bdb, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("数据库连接失败：%w", err)
	}

	// 验证连接
	if err := bdb.Ping(); err != nil {
		return fmt.Errorf("数据库 ping 失败：%w", err)
	}

	// 初始化lastTime：从数据库读取上次拉取时间，默认取1小时前
	lastTime = time.Now().Add(-1 * time.Hour)
	err = bdb.QueryRow("SELECT last_pull_time FROM bilibili_config LIMIT 1").Scan(&lastTime)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("读取上次拉取时间失败：%w", err)
	}

	return nil
}

// 补全缺失的函数：获取用户动态卡片
func getUserDynamicCard(userID string) ([]*DynamicCard, error) {
	if userID == "" {
		return nil, errors.New("用户ID不能为空")
	}

	// 验证数据库连接是否初始化
	if bdb == nil {
		return nil, errors.New("数据库未初始化，请先调用InitDB")
	}

	// 示例实现：查询用户最新动态（实际需替换为B站动态API或数据库查询逻辑）
	query := `SELECT id, content, image_urls, published_at, like_count 
			  FROM bilibili_dynamics 
			  WHERE user_id = ? AND published_at > ? 
			  ORDER BY published_at DESC`

	rows, err := bdb.Query(query, userID, lastTime)
	if err != nil {
		return nil, fmt.Errorf("查询动态失败：%w", err)
	}
	defer rows.Close()

	var dynamics []*DynamicCard
	for rows.Next() {
		var card DynamicCard
		var imageURLs string // 假设数据库中是逗号分隔的URL字符串
		err := rows.Scan(&card.ID, &card.Content, &imageURLs, &card.PublishedAt, &card.LikeCount)
		if err != nil {
			return nil, fmt.Errorf("解析动态数据失败：%w", err)
		}
		// 处理图片URL（示例：分割逗号）
		card.ImageURLs = strings.Split(imageURLs, ",")
		dynamics = append(dynamics, &card)
	}

	// 更新上次拉取时间
	if len(dynamics) > 0 {
		lastTime = dynamics[0].PublishedAt
		_, err := bdb.Exec("UPDATE bilibili_config SET last_pull_time = ?", lastTime)
		if err != nil {
			return dynamics, fmt.Errorf("更新上次拉取时间失败：%w", err)
		}
	}

	return dynamics, nil
}

// 原文件中的推送函数（示例，根据实际逻辑调整）
func PushUserDynamic(userID string) error {
	// 补全bdb的初始化检查（原第28行错误修复）
	if bdb == nil {
		return errors.New("数据库未初始化")
	}

	// 补全getUserDynamicCard调用（原第31行错误修复）
	cards, err := getUserDynamicCard(userID)
	if err != nil {
		return fmt.Errorf("获取用户动态失败：%w", err)
	}

	// 补全lastTime使用（原第38行错误修复）
	fmt.Printf("上次拉取时间：%s，本次获取动态数：%d\n", lastTime.Format(time.RFC3339), len(cards))

	// 推送逻辑（示例）
	for _, card := range cards {
		fmt.Printf("推送动态：%s\n", card.Content)
		// 实际推送逻辑（如HTTP请求、消息队列等）
	}

	return nil
}
