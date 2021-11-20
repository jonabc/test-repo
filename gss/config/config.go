package config

import (
	"flag"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/github/go-config"
	"github.com/github/go-dbase"
	"github.com/github/go-exceptions"
	httpexporter "github.com/github/go-exceptions/exporters/http"
	"github.com/github/go-exceptions/exporters/writer"
	"github.com/github/go-exceptions/stacktracers/pkgerrors"
	"github.com/github/go-log"
	"github.com/github/go-stats"
	"github.com/lightstep/lightstep-tracer-go"
	"github.com/opentracing/opentracing-go"
)

// Config holds application configuration, including statting and tracing configs.
type Config struct {
	HTTPPort    int
	ServiceName string `config:"go-sample-service,env=APP_NAME"`
	Environment string `config:"development,env=HEAVEN_DEPLOYED_ENV"`
	Sha         string `config:",env=APP_SHA"`
	Ref         string `config:",env=APP_REF"`

	StatsAddr   string        `config:",env=STATS_ADDR"`
	StatsPeriod time.Duration `config:"10s,env=STATS_PERIOD"`

	TracingEnabled bool   `config:"false,env=TRACING_ENABLED"`
	TracingHost    string `config:",env=TRACING_HOST"`
	TracingPort    int    `config:"443,env=TRACING_PORT"`
	TracingToken   string `config:",env=TRACING_TOKEN"`

	SkeemaFile string `config:"schemas/.skeema,env=SKEEMA_FILE"`
}

// Load parses configuration from the environment and places it in a newly
// allocated Config struct.
func Load() (*Config, error) {
	// initialize configuration
	port := flag.Int("port", 8080, "port number to run http server on")
	flag.Parse()

	cfg := &Config{
		HTTPPort: *port,
	}

	if err := config.Load(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// NewLogger initializes and returns a new logger object.
func (cfg *Config) NewLogger() *log.Logger {
	return log.New(log.InfoLevel, &log.Logfmt{Sink: log.Stdout})
}

// NewExceptionReporter configures a new exceptions reporter based on the config.
// It prints reports to stderr in development environment, but sends them to
// Failbotg (which is proxied to Sentry) in production.
func (cfg *Config) NewExceptionReporter() (*exceptions.Reporter, error) {
	var err error
	var exporter exceptions.Exporter = writer.NewExporter(os.Stderr)

	if cfg.Environment == "production" {
		exporter, err = httpexporter.NewExporter()
		if err != nil {
			return nil, err
		}
	}

	reporter, err := exceptions.NewReporter(
		exceptions.WithExporter(exporter),
		exceptions.WithApplication(cfg.ServiceName),
		exceptions.WithStacktraceFunc(pkgerrors.NewStackTracer()),
		exceptions.WithValues(map[string]string{
			"deployed_to": cfg.Environment,
			"release":     cfg.Sha,
			"ref":         cfg.Ref,
		}),
	)
	if err != nil {
		return nil, err
	}

	return reporter, nil
}

// NewStatsClient generates a new statter for use with DataDog.
func (cfg *Config) NewStatsClient() (stats.Client, error) {
	if cfg.StatsAddr == "" {
		return stats.NullStatter, nil
	}

	return stats.NewClient(stats.UDPSink(cfg.StatsAddr), cfg.StatsPeriod, "go-sample-service"), nil
}

// NewTracer initializes distributed tracing based on config.
func (cfg *Config) NewTracer() opentracing.Tracer {
	if !cfg.TracingEnabled || cfg.TracingToken == "" {
		return opentracing.NoopTracer{}
	}

	opts := lightstep.Options{
		AccessToken: cfg.TracingToken,
		Collector: lightstep.Endpoint{
			Host:      cfg.TracingHost,
			Port:      cfg.TracingPort,
			Plaintext: false,
		},
		Tags: opentracing.Tags{
			lightstep.ComponentNameKey: cfg.ServiceName,
			"service.version":          cfg.Sha,
			"env":                      cfg.Environment,
		},
	}

	return lightstep.NewTracer(opts)
}

// NewDatabaseConfig returns a new MySQL configuration
func (cfg *Config) NewDatabaseConfig() (*mysql.Config, error) {

	dbCfg, err := dbase.Config(cfg.SkeemaFile, cfg.Environment)
	if err != nil {
		return nil, err
	}

	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = dbCfg.User
	mysqlConfig.Passwd = dbCfg.Passwd
	mysqlConfig.DBName = dbCfg.DBName
	mysqlConfig.Net = dbCfg.Net
	mysqlConfig.Addr = dbCfg.Addr
	mysqlConfig.AllowNativePasswords = true
	mysqlConfig.InterpolateParams = true
	mysqlConfig.ParseTime = true

	return mysqlConfig, nil
}
