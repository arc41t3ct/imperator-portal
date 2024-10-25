package imperator

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jet "github.com/CloudyKit/jet/v6"
	scs "github.com/alexedwards/scs/v2"
	"github.com/arc41t3ct/imperator/cache"
	"github.com/arc41t3ct/imperator/mailer"
	"github.com/arc41t3ct/imperator/render"
	"github.com/arc41t3ct/imperator/session"
	badger "github.com/dgraph-io/badger/v4"
	chi "github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	cron "github.com/robfig/cron/v3"
)

const version = "0.1.0"

// define some globl variables that we need in different places
var appRedisInstance *cache.RedisCache
var appBadgerInstance *cache.BadgerCache
var redisPool *redis.Pool
var badgerConn *badger.DB

// Imperator is the application wide type for the Imperator package. Members that are exported to this type
// are available to any application that uses it.
type Imperator struct {
	RootPath      string
	AppName       string
	Version       string
	Debug         bool
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	Routes        *chi.Mux
	Render        *render.Render
	Session       *scs.SessionManager
	Validator     *Validation
	DB            Database
	JetViews      *jet.Set
	EncryptionKey string
	Cache         cache.Cache
	Schedular     *cron.Cron
	Mail          mailer.Mail
	Server        Server
	// internal not accessible by implementors
	config config
}

type Server struct {
	ServerName string
	Port       string
	Secure     bool
	URL        string
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
	redis       redisConfig
}

// New reads the .env file, creates our application config, populates the Imperator type with configuration
// based on .env values, and creates the necessary folders and files if they don't exist yet.
func (i *Imperator) New(rootPath string) error {
	// most important root path
	i.RootPath = rootPath
	i.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	// folders that we will create if they don't exist yet
	if err := i.createMissingPaths(); err != nil {
		return err
	}
	// check and load .env
	if err := i.checkDotEnv(); err != nil {
		return err
	}
	// read .env
	if err := godotenv.Load(i.RootPath + "/.env"); err != nil {
		return err
	}
	// create loggers
	infoLog, errorLog := i.StartLoggers()
	i.InfoLog = infoLog
	i.ErrorLog = errorLog
	// connect to databases
	if err := i.createDatabasePool(); err != nil {
		return err
	}
	// create a schedular
	schedular := cron.New()
	i.Schedular = schedular
	// create session and cache
	if err := i.createCacheAndSessionStore(); err != nil {
		return err
	}
	// bootstrap imperitor
	i.AppName = os.Getenv("APP_NAME")
	i.Version = version
	i.Mail = i.createMailer()
	i.EncryptionKey = os.Getenv("ENCRYPTION_KEY")
	i.Routes = i.routes().(*chi.Mux)
	// create internal config
	if err := i.createInternalConfig(); err != nil {
		return err
	}
	// create server config
	if err := i.createInternalConfig(); err != nil {
		return err
	}
	// allows editing templates and reloading ok for development
	if err := i.createJetTemplatesConfig(); err != nil {
		return err
	}
	// createSession must come before createRenderer
	i.createSession()
	i.createRenderer()

	go i.Mail.ListenForMail()

	return nil
}

func (i *Imperator) createJetTemplatesConfig() error {
	var views *jet.Set
	views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", i.RootPath)),
	)
	if i.Debug {
		views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", i.RootPath)),
			jet.InDevelopmentMode(),
		)
	}
	i.JetViews = views
	return nil
}

func (i *Imperator) createServerConfig() error {
	secure := true
	if strings.ToLower(os.Getenv("SECURE")) == "false" {
		secure = false
	}
	i.Server = Server{
		ServerName: os.Getenv("SERVER_NAME"),
		Port:       os.Getenv("PORT"),
		Secure:     secure,
		URL:        os.Getenv("APP_URL"),
	}
	return nil
}

func (i *Imperator) createInternalConfig() error {
	i.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      i.BuildDSN(),
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}
	return nil
}

func (i *Imperator) createMissingPaths() error {
	var p = initPaths{
		rootPath: i.RootPath,
		folderNames: []string{
			"handlers", "migrations", "views", "views/layouts", "mail",
			"models", "public", "public/images", "public/ico", "middleware",
		},
	}
	root := p.rootPath
	for _, path := range p.folderNames {
		// create folder if not exists
		err := i.CreateDirIfNotExists(root + "/" + path)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListenAndServe starts the web server
func (i *Imperator) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", i.config.port),
		ErrorLog:     i.ErrorLog,
		Handler:      i.Routes,
		IdleTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * 60 * time.Second,
	}

	// close database pools when we stop running
	if i.DB.Pool != nil {
		defer i.DB.Pool.Close()
	}

	if redisPool != nil {
		defer redisPool.Close()
	}

	if badgerConn != nil {
		defer badgerConn.Close()
	}

	i.InfoLog.Println(i.AppName, "listening on port:", i.config.port)
	err := srv.ListenAndServe()
	i.ErrorLog.Fatal(err)
}

func (i *Imperator) checkDotEnv() error {
	err := i.CreateFileIfNotExists(fmt.Sprintf("%s/.env", i.RootPath))
	if err != nil {
		return err
	}
	return nil
}

func (i *Imperator) StartLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return infoLog, errorLog
}

func (i *Imperator) createRenderer() {
	renderer := render.Render{
		Renderer: i.config.renderer,
		RootPath: i.RootPath,
		Port:     i.config.port,
		JetViews: i.JetViews,
		Session:  i.Session,
	}
	i.Render = &renderer
}

func (i *Imperator) createMailer() mailer.Mail {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	return mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Templates:   i.RootPath + "/mail",
		Host:        os.Getenv("SMTP_HOST"),
		Port:        port,
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWOR"),
		Encryption:  os.Getenv("SMTP_ENCRYPTION"),
		FromName:    os.Getenv("MAIL_FROM_NAME"),
		FromAddress: os.Getenv("MAIL_FROM_ADDRESS"),
		Jobs:        make(chan mailer.Message, 20),
		Results:     make(chan mailer.Result, 20),
		API:         os.Getenv("MAILER_API"),
		APIKey:      os.Getenv("MAILER_KEY"),
		APIUrl:      os.Getenv("MAILER_URL"),
	}
}

func (i *Imperator) createSession() {
	sessionMgr := session.Session{
		CookieLifetime: i.config.cookie.lifetime,
		CookiePersist:  i.config.cookie.persist,
		CookieName:     i.config.cookie.name,
		CookieDomain:   i.config.cookie.domain,
		CookieSecure:   i.config.cookie.secure,
		SessionType:    i.config.sessionType,
		DBPool:         i.DB.Pool,
	}

	switch i.config.sessionType {
	case "redis":
		sessionMgr.RedisPool = appRedisInstance.Conn
	case "mysql", "postgres", "mariadb", "postgresql":
		sessionMgr.DBPool = i.DB.Pool
	}

	i.Session = sessionMgr.InitSession()
}

func (i *Imperator) createClientRedisCache() *cache.RedisCache {
	cacheClient := cache.RedisCache{
		Conn:   i.createRedisPool(),
		Prefix: i.config.redis.prefix,
	}
	return &cacheClient
}

func (i *Imperator) createClientBadgerCache() (*cache.BadgerCache, error) {
	conn, err := i.createBadgerConn()
	if err != nil {
		return nil, err
	}
	cacheClient := cache.BadgerCache{
		Conn: conn,
	}
	return &cacheClient, nil
}

func (i *Imperator) createBadgerConn() (*badger.DB, error) {
	db, err := badger.Open(badger.DefaultOptions(i.RootPath + "/tmp/badger"))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (i *Imperator) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", i.config.redis.host, redis.DialPassword(i.config.redis.password))
		},

		TestOnBorrow: func(c redis.Conn, lastUsed time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (i *Imperator) BuildDSN() string {
	var dsn string
	switch i.DB.DatabaseType {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"))
		if os.Getenv("DATABASE_PASSWORD") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASSWORD"))
		}
	default:
	}
	return dsn
}

func (i *Imperator) createDatabasePool() error {
	if os.Getenv("DATABASE_TYPE") != "" {
		i.DB = Database{}
		i.DB.DatabaseType = os.Getenv("DATABASE_TYPE")
		db, err := i.OpenDB(os.Getenv("DATABASE_TYPE"), i.BuildDSN())
		if err != nil {
			i.ErrorLog.Println(err)
			os.Exit(1)
		}
		i.DB.Pool = db
	}
	return nil
}

func (i *Imperator) createCacheAndSessionStore() error {
	if os.Getenv("CACHE_TYPE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		appRedisInstance = i.createClientRedisCache()
		i.Cache = appRedisInstance
		redisPool = appRedisInstance.Conn
	}

	if os.Getenv("CACHE_TYPE") == "badger" {
		appBadgerInstance, err := i.createClientBadgerCache()
		if err != nil {
			return err
		}
		i.Cache = appBadgerInstance
		badgerConn = appBadgerInstance.Conn

		_, err = i.Schedular.AddFunc("@daily", func() {
			_ = appBadgerInstance.Conn.RunValueLogGC(0.7)
		})
		if err != nil {
			return err
		}
	}
	return nil
}
