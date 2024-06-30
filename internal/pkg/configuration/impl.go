package configuration

import (
	"fmt"
	"log/slog"
	"net"
	"runtime/debug"
)

type Database struct {
	Host     string `json:"database_host" mapstructure:"database_host"`
	Port     string `json:"database_port" mapstructure:"database_port"`
	User     string `json:"database_user" mapstructure:"database_user"`
	Password string `json:"database_password" mapstructure:"database_password"`
	Name     string `json:"database_name" mapstructure:"database_name"`
	URL      string `json:"database_url" mapstructure:"database_url"`
}

func (db Database) ToURL() string {
	if db.URL != "" {
		return db.URL
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		db.User,
		db.Password,
		db.Host,
		db.Port,
		db.Name,
	)
}

type Service struct {
	HostName    string `json:"service_host" mapstructure:"service_host"`
	Port        string `json:"service_port" mapstructure:"service_port"`
	Name        string `json:"service_name" mapstructure:"service_name"`
	Environment string `json:"service_environment" mapstructure:"service_environment"`
	CommitHash  string `json:"service_commit_hash"`
	GoVersion   string `json:"service_go_version"`
}

func (s *Service) SetDefaults() {
	s.Port = "8080"
	s.Environment = "local"
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, v := range buildInfo.Settings {
			if v.Key == "vcs.revision" && len(v.Value) >= 7 {
				s.CommitHash = v.Value[:7] // git short hash
			}
		}
		s.GoVersion = buildInfo.GoVersion
	}
}

func (s Service) ToLogFields() []any {
	return []any{slog.String("name", s.Name), slog.String("env", s.Environment), slog.String("commit", s.CommitHash), slog.String("go_version", s.GoVersion)}
}

func (s Service) ToHost() string {
	return net.JoinHostPort(s.HostName, s.Port)
}

type NATS struct {
	Address string `json:"nats_address" mapstructure:"nats_address"`
	Port    string `json:"nats_port" mapstructure:"nats_port"`
	URL     string `json:"nats_url" mapstructure:"nats_url"`
	JWT     string `json:"nats_jwt" mapstructure:"nats_jwt"`
	Seed    string `json:"nats_seed" mapstructure:"nats_seed"`
}

func (n NATS) ToURL() string {
	if n.URL != "" {
		return n.URL
	}
	return fmt.Sprintf("nats://%s:%s",
		n.Address,
		n.Port,
	)
}
