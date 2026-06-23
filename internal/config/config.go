package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 汇聚所有子配置项
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Workflow WorkflowConfig
}

type AppConfig struct {
	Env  string
	Port string
	Name string
}

type DatabaseConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	Name           string
	MaxConns       int32
	MinConns       int32
	MigrationsPath string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// WorkflowConfig 控制审批流程的业务参数
type WorkflowConfig struct {
	// AutoApproveMaxDays 不超过该天数的请假走自动审批
	AutoApproveMaxDays int
	// MinimumOnSiteRatio 部门最低在岗比例（0~1），优先于 MinimumOnSiteCount
	MinimumOnSiteRatio float64
	// MinimumOnSiteCount 部门最低在岗绝对人数，当比例计算结果低于此值时取此值
	MinimumOnSiteCount int
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable pool_max_conns=%d pool_min_conns=%d",
		d.Host, d.Port, d.User, d.Password, d.Name, d.MaxConns, d.MinConns,
	)
}

// Load 从环境变量和配置文件加载配置
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/app/config")

	// 允许通过环境变量覆盖，例如 OA_DATABASE_HOST
	v.SetEnvPrefix("OA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 设置默认值
	v.SetDefault("app.env", "development")
	v.SetDefault("app.port", "8080")
	v.SetDefault("app.name", "oa-leave-system")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.maxConns", 20)
	v.SetDefault("database.minConns", 2)
	v.SetDefault("database.migrationsPath", "/app/migrations")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.expiration", "24h")
	v.SetDefault("workflow.autoApproveMaxDays", 1)
	v.SetDefault("workflow.minimumOnSiteRatio", 0.5)
	v.SetDefault("workflow.minimumOnSiteCount", 2)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	expStr := v.GetString("jwt.expiration")
	exp, err := time.ParseDuration(expStr)
	if err != nil {
		return nil, fmt.Errorf("JWT 过期时间格式无效: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Env:  v.GetString("app.env"),
			Port: v.GetString("app.port"),
			Name: v.GetString("app.name"),
		},
		Database: DatabaseConfig{
			Host:           v.GetString("database.host"),
			Port:           v.GetInt("database.port"),
			User:           v.GetString("database.user"),
			Password:       v.GetString("database.password"),
			Name:           v.GetString("database.name"),
			MaxConns:       int32(v.GetInt("database.maxConns")),
			MinConns:       int32(v.GetInt("database.minConns")),
			MigrationsPath: v.GetString("database.migrationsPath"),
		},
		Redis: RedisConfig{
			Addr:     v.GetString("redis.addr"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
		},
		JWT: JWTConfig{
			Secret:     v.GetString("jwt.secret"),
			Expiration: exp,
		},
		Workflow: WorkflowConfig{
			AutoApproveMaxDays: v.GetInt("workflow.autoApproveMaxDays"),
			MinimumOnSiteRatio: v.GetFloat64("workflow.minimumOnSiteRatio"),
			MinimumOnSiteCount: v.GetInt("workflow.minimumOnSiteCount"),
		},
	}

	return cfg, nil
}
