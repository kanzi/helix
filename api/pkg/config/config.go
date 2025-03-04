package config

import (
	"fmt"
	"time"

	"github.com/helixml/helix/api/pkg/types"
	"github.com/kelseyhightower/envconfig"
)

type ServerConfig struct {
	Inference          Inference
	Providers          Providers
	Tools              Tools
	Keycloak           Keycloak
	Notifications      Notifications
	Janitor            Janitor
	Stripe             Stripe
	DataPrepText       DataPrepText
	TextExtractor      TextExtractor
	RAG                RAG
	Controller         Controller
	FileStore          FileStore
	Store              Store
	PubSub             PubSub
	WebServer          WebServer
	SubscriptionQuotas SubscriptionQuotas
	GitHub             GitHub
	FineTuning         FineTuning
	Apps               Apps
	GPTScript          GPTScript
	Triggers           Triggers
	SSL                SSL

	DisableLLMCallLogging bool `envconfig:"DISABLE_LLM_CALL_LOGGING" default:"false"`
}

func LoadServerConfig() (ServerConfig, error) {
	var cfg ServerConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		return ServerConfig{}, err
	}
	return cfg, nil
}

type Inference struct {
	Provider types.Provider `envconfig:"INFERENCE_PROVIDER" default:"helix" description:"One of helix, openai, or togetherai"`
}

// Providers is used to configure the various AI providers that we use
type Providers struct {
	OpenAI     OpenAI
	TogetherAI TogetherAI
	Helix      Helix
}

type OpenAI struct {
	BaseURL               string        `envconfig:"OPENAI_BASE_URL" default:"https://api.openai.com/v1"`
	APIKey                string        `envconfig:"OPENAI_API_KEY"`
	APIKeyFromFile        string        `envconfig:"OPENAI_API_KEY_FILE"` // i.e. /run/secrets/openai-api-key
	APIKeyRefreshInterval time.Duration `envconfig:"OPENAI_API_KEY_REFRESH_INTERVAL" default:"3s"`
}

type TogetherAI struct {
	BaseURL               string        `envconfig:"TOGETHER_BASE_URL" default:"https://api.together.xyz/v1"`
	APIKey                string        `envconfig:"TOGETHER_API_KEY"`
	APIKeyFromFile        string        `envconfig:"TOGETHER_API_KEY_FILE"` // i.e. /run/secrets/together-api-key
	APIKeyRefreshInterval time.Duration `envconfig:"TOGETHER_API_KEY_REFRESH_INTERVAL" default:"3s"`
}

type Helix struct {
	OwnerID            string        `envconfig:"TOOLS_PROVIDER_HELIX_OWNER_ID" default:"helix-internal"` // Will be used for sesions
	OwnerType          string        `envconfig:"TOOLS_PROVIDER_HELIX_OWNER_TYPE" default:"system"`       // Will be used for sesions
	ModelTTL           time.Duration `envconfig:"HELIX_MODEL_TTL" default:"10s"`                          // How long to keep models warm before allowing other work to be scheduled
	SlotTTL            time.Duration `envconfig:"HELIX_SLOT_TTL" default:"300s"`                          // How long to wait for work to complete before slots are considered dead
	RunnerTTL          time.Duration `envconfig:"HELIX_RUNNER_TTL" default:"30s"`                         // How long before runners are considered dead
	SchedulingStrategy string        `envconfig:"HELIX_SCHEDULING_STRATEGY" default:"max_spread" description:"The strategy to use for scheduling workloads."`
	QueueSize          int           `envconfig:"HELIX_QUEUE_SIZE" default:"100" description:"The size of the queue when buffering workloads."`
}

type Tools struct {
	Enabled bool `envconfig:"TOOLS_ENABLED" default:"true"` // Enable/disable tools for the server

	// Suggestions based on provider (now set by INFERENCE_PROVIDER):
	// - OpenAI: gpt-4-1106-preview
	// - Together AI: meta-llama/Llama-3-8b-chat-hf
	// - Helix: llama3:instruct
	Model string `envconfig:"TOOLS_MODEL" default:"llama3:instruct"`

	// IsActionableTemplate is used to determine whether Helix should
	// use a tool or not. Leave empty for default
	IsActionableTemplate string `envconfig:"TOOLS_IS_ACTIONABLE_TEMPLATE"` // Either plain text, base64 or path to a file
}

// Keycloak is used for authentication. You can find keycloak documentation
// at https://www.keycloak.org/guides
type Keycloak struct {
	KeycloakURL         string `envconfig:"KEYCLOAK_URL" default:"http://keycloak:8080/auth"`
	KeycloakFrontEndURL string `envconfig:"KEYCLOAK_FRONTEND_URL" default:"http://localhost:8080/auth"`
	ServerURL           string `envconfig:"SERVER_URL" description:"The URL the api server is listening on."`
	APIClientID         string `envconfig:"KEYCLOAK_CLIENT_ID" default:"api"`
	ClientSecret        string `envconfig:"KEYCLOAK_CLIENT_SECRET"` // If not set, will be looked up using admin API
	FrontEndClientID    string `envconfig:"KEYCLOAK_FRONTEND_CLIENT_ID" default:"frontend"`
	AdminRealm          string `envconfig:"KEYCLOAK_ADMIN_REALM" default:"master"`
	Realm               string `envconfig:"KEYCLOAK_REALM" default:"helix"`
	Username            string `envconfig:"KEYCLOAK_USER"`
	Password            string `envconfig:"KEYCLOAK_PASSWORD"`
}

// Notifications is used for sending notifications to users when certain events happen
// such as finetuning starting or completing.
type Notifications struct {
	AppURL string `envconfig:"APP_URL" default:"https://app.tryhelix.ai"`
	Email  EmailConfig
	// TODO: Slack, Discord, etc.
}

type EmailConfig struct {
	SenderAddress string `envconfig:"EMAIL_SENDER_ADDRESS" default:"chris@helix.ml"`

	SMTP struct {
		Host     string `envconfig:"EMAIL_SMTP_HOST"`
		Port     string `envconfig:"EMAIL_SMTP_PORT"`
		Identity string `envconfig:"EMAIL_SMTP_IDENTITY"`
		Username string `envconfig:"EMAIL_SMTP_USERNAME"`
		Password string `envconfig:"EMAIL_SMTP_PASSWORD"`
	}

	Mailgun struct {
		Domain string `envconfig:"EMAIL_MAILGUN_DOMAIN"`
		APIKey string `envconfig:"EMAIL_MAILGUN_API_KEY"`
		Europe bool   `envconfig:"EMAIL_MAILGUN_EUROPE" default:"false"` // use EU region
	}
}

type Janitor struct {
	AppURL                  string
	SentryDsnAPI            string   `envconfig:"SENTRY_DSN_API" description:"The api sentry DSN."`
	SentryDsnFrontend       string   `envconfig:"SENTRY_DSN_FRONTEND" description:"The frontend sentry DSN."`
	GoogleAnalyticsFrontend string   `envconfig:"GOOGLE_ANALYTICS_FRONTEND" description:"The frontend Google analytics id."`
	SlackWebhookURL         string   `envconfig:"JANITOR_SLACK_WEBHOOK_URL" description:"The slack webhook URL to ping messages to."`
	SlackIgnoreUser         []string `envconfig:"JANITOR_SLACK_IGNORE_USERS" description:"Ignore keycloak user ids for slack messages."`
	RudderStackWriteKey     string   `envconfig:"RUDDERSTACK_WRITE_KEY" description:"The write key for rudderstack."`
	RudderStackDataPlaneURL string   `envconfig:"RUDDERSTACK_DATA_PLANE_URL" description:"The data plane URL for rudderstack."`
}

type Stripe struct {
	AppURL               string
	SecretKey            string `envconfig:"STRIPE_SECRET_KEY" description:"The secret key for stripe."`
	WebhookSigningSecret string `envconfig:"STRIPE_WEBHOOK_SIGNING_SECRET" description:"The webhook signing secret for stripe."`
	PriceLookupKey       string `envconfig:"STRIPE_PRICE_LOOKUP_KEY" description:"The lookup key for the stripe price."`
}

type DataPrepText struct {
	Module            types.DataPrepModule `envconfig:"DATA_PREP_TEXT_MODULE" default:"dynamic" description:"Which module to use for text data prep."`
	OverflowSize      int                  `envconfig:"DATA_PREP_TEXT_OVERFLOW_SIZE" default:"256" description:"The overflow size for the text data prep."`
	QuestionsPerChunk int                  `envconfig:"DATA_PREP_TEXT_QUESTIONS_PER_CHUNK" default:"30" description:"The questions per chunk for the text data prep."`
	Temperature       float32              `envconfig:"DATA_PREP_TEXT_TEMPERATURE" default:"0.5" description:"The temperature for the text data prep prompt."`
}

type TextExtractor struct {
	Provider types.Extractor `envconfig:"TEXT_EXTRACTION_PROVIDER" default:"tika"`
	// the URL we post documents to so we can get the text back from them
	Unstructured struct {
		URL string `envconfig:"TEXT_EXTRACTION_URL" default:"http://llamaindex:5000/api/v1/extract" description:"The URL to extract text from a document."`
	}

	Tika struct {
		URL string `envconfig:"TEXT_EXTRACTION_TIKA_URL" default:"http://tika:9998" description:"The URL to extract text from a document."`
	}
}

type RAG struct {
	IndexingConcurrency int `envconfig:"RAG_INDEXING_CONCURRENCY" default:"1" description:"The number of concurrent indexing tasks."`

	// DefaultRagProvider is the default RAG provider to use if not specified
	DefaultRagProvider string `envconfig:"RAG_DEFAULT_PROVIDER" default:"typesense" description:"The default RAG provider to use if not specified."`

	MaxVersions int `envconfig:"RAG_MAX_VERSIONS" default:"3" description:"The maximum number of versions to keep for a knowledge."`

	// Typesense is used to store RAG records in a Typesense index
	Typesense struct {
		URL    string `envconfig:"RAG_TYPESENSE_URL" default:"http://typesense:8108" description:"The URL to the Typesense server."`
		APIKey string `envconfig:"RAG_TYPESENSE_API_KEY" default:"typesense" description:"The API key to the Typesense server."`
	}

	Llamaindex struct {
		// the URL we can post a chunk of text to for RAG indexing
		RAGIndexingURL string `envconfig:"RAG_INDEX_URL" default:"http://llamaindex:5000/api/v1/rag/chunk" description:"The URL to index text with RAG."`
		// the URL we can post a prompt to to match RAG records
		RAGQueryURL string `envconfig:"RAG_QUERY_URL" default:"http://llamaindex:5000/api/v1/rag/query" description:"The URL to query RAG records."`
		// the URL we can post a delete request to for RAG records,
		// this is a prefix, full path is http://llamaindex:5000/api/v1/rag/<data_entity_id>
		RAGDeleteURL string `envconfig:"RAG_DELETE_URL" default:"http://llamaindex:5000/api/v1/rag" description:"The URL to delete RAG records."`
	}

	Crawler struct {
		ChromeURL       string `envconfig:"RAG_CRAWLER_CHROME_URL" default:"http://chrome:9222" description:"The URL to the Chrome instance."`
		LauncherEnabled bool   `envconfig:"RAG_CRAWLER_LAUNCHER_ENABLED" default:"true" description:"Whether to use the Launcher to start the browser."`
		LauncherURL     string `envconfig:"RAG_CRAWLER_LAUNCHER_URL" default:"http://chrome:7317" description:"The URL to the Launcher instance."`
		BrowserPoolSize int    `envconfig:"RAG_CRAWLER_BROWSER_POOL_SIZE" default:"5" description:"The number of browsers to keep in the pool."`
		PagePoolSize    int    `envconfig:"RAG_CRAWLER_PAGE_POOL_SIZE" default:"50" description:"The number of pages to keep in the pool."`

		// Limits
		MaxFrequency time.Duration `envconfig:"RAG_CRAWLER_MAX_FREQUENCY" default:"60m" description:"The maximum frequency to crawl."`
		MaxDepth     int           `envconfig:"RAG_CRAWLER_MAX_DEPTH" default:"100" description:"The maximum depth to crawl."`
	}
}

type Controller struct {
	FilestorePresignSecret string `envconfig:"FILESTORE_PRESIGN_SECRET" description:""`
	// this is an "env" prefix like "dev"
	// the user prefix is handled inside the controller
	// (see getFilestorePath)
	FilePrefixGlobal string `envconfig:"FILE_PREFIX_GLOBAL" default:"dev" description:"The global prefix path for the filestore."`
	// this is a golang template that is used to prefix the user
	// path in the filestore - it is passed Owner and OwnerType values
	// write me an example FilePrefixUser as a go template
	// e.g. "users/{{.Owner}}"
	// FilePrefixUser string `envconfig:"FILE_PREFIX_USER" default:"users/{{.Owner}}" description:"The go template that produces the prefix path for a user."`

	// a static path used to denote what sub-folder job results live in
	FilePrefixResults string `envconfig:"FILE_PREFIX_RESULTS" default:"results" description:"The go template that produces the prefix path for a user."`

	// how many scheduler decisions to buffer before we start dropping them
	SchedulingDecisionBufferSize int `envconfig:"SCHEDULING_DECISION_BUFFER_SIZE" default:"10" description:"How many scheduling decisions to buffer before we start dropping them."`
}

type FileStore struct {
	Type         types.FileStoreType `envconfig:"FILESTORE_TYPE" default:"fs" description:"What type of filestore should we use (fs | gcs)."`
	LocalFSPath  string              `envconfig:"FILESTORE_LOCALFS_PATH" default:"/tmp/helix/filestore" description:"The local path that is the root for the local fs filestore."`
	GCSKeyBase64 string              `envconfig:"FILESTORE_GCS_KEY_BASE64" description:"The base64 encoded service account json file for GCS."`
	GCSKeyFile   string              `envconfig:"FILESTORE_GCS_KEY_FILE" description:"The local path to the service account json file for GCS."`
	GCSBucket    string              `envconfig:"FILESTORE_GCS_BUCKET" description:"The bucket we are storing things in GCS."`
}

type PubSub struct {
	StoreDir string `envconfig:"NATS_STORE_DIR" default:"/filestore/nats" description:"The directory to store nats data."`
}

type Store struct {
	Host     string `envconfig:"POSTGRES_HOST" description:"The host to connect to the postgres server."`
	Port     int    `envconfig:"POSTGRES_PORT" default:"5432" description:"The port to connect to the postgres server."`
	Database string `envconfig:"POSTGRES_DATABASE" default:"helix" description:"The database to connect to the postgres server."`
	Username string `envconfig:"POSTGRES_USER" description:"The username to connect to the postgres server."`
	Password string `envconfig:"POSTGRES_PASSWORD" description:"The password to connect to the postgres server."`
	SSL      bool   `envconfig:"POSTGRES_SSL" default:"false"`
	Schema   string `envconfig:"POSTGRES_SCHEMA"` // Defaults to public

	AutoMigrate     bool          `envconfig:"DATABASE_AUTO_MIGRATE" default:"true" description:"Should we automatically run the migrations?"`
	MaxConns        int           `envconfig:"DATABASE_MAX_CONNS" default:"50"`
	IdleConns       int           `envconfig:"DATABASE_IDLE_CONNS" default:"25"`
	MaxConnLifetime time.Duration `envconfig:"DATABASE_MAX_CONN_LIFETIME" default:"1h"`
	MaxConnIdleTime time.Duration `envconfig:"DATABASE_MAX_CONN_IDLE_TIME" default:"1m"`
}

type WebServer struct {
	URL  string `envconfig:"SERVER_URL" description:"The URL the api server is listening on."`
	Host string `envconfig:"SERVER_HOST" default:"0.0.0.0" description:"The host to bind the api server to."`
	Port int    `envconfig:"SERVER_PORT" default:"80" description:""`
	// Can either be a URL to frontend or a path to static files
	FrontendURL string `envconfig:"FRONTEND_URL" default:"http://frontend:8081" description:""`

	RunnerToken string `envconfig:"RUNNER_TOKEN" description:"The token for runner auth."`
	// a list of keycloak ids that are considered admins
	// if the string 'all' is included it means ALL users
	AdminIDs []string `envconfig:"ADMIN_USER_IDS" description:"Keycloak admin IDs."`
	// Specifies the source of the Admin user IDs.
	// By default AdminSrc is set to env.
	AdminSrc AdminSrcType `envconfig:"ADMIN_USER_SOURCE" default:"env" description:"Source of admin IDs (env or jwt)"`
	// if this is specified then we provide the option to clone entire
	// sessions into this user without having to logout and login
	EvalUserID string `envconfig:"EVAL_USER_ID" description:""`
	// this is for when we are running localfs filesystem
	// and we need to add a route to view files based on their path
	// we are assuming all file storage is open right now
	// so we just deep link to the object path and don't apply auth
	// (this is so helix nodes can see files)
	// later, we might add a token to the URLs
	LocalFilestorePath string
}

// AdminSrcType is an enum specifyin the type of Admin ID source.
// It currently supports only two sources:
// * env: ADMIN_USER_IDS env var
// * jwt: admin JWT token claim
type AdminSrcType string

const (
	AdminSrcTypeEnv AdminSrcType = "env"
	AdminSrcTypeJWT AdminSrcType = "jwt"
)

// String implements fmt.Stringer
func (a AdminSrcType) String() string {
	return string(a)
}

// Decode implements envconfig.Decoder for value validation.
func (a *AdminSrcType) Decode(value string) error {
	if value == "" {
		*a = AdminSrcTypeEnv
		return nil
	}
	switch value {
	case string(AdminSrcTypeEnv), string(AdminSrcTypeJWT):
		*a = AdminSrcType(value)
		return nil
	default:
		return fmt.Errorf("invalid source of admin IDs: %q", value)
	}
}

func (a *AdminSrcType) UnmarshalText(text []byte) error {
	return a.Decode(string(text))
}

type SubscriptionQuotas struct {
	Enabled    bool `envconfig:"SUBSCRIPTION_QUOTAS_ENABLED" default:"true"`
	Finetuning struct {
		Free struct {
			Strict        bool `envconfig:"SUBSCRIPTION_QUOTAS_FINETUNING_FREE_STRICT" default:"true"` // If set, will now allow any finetuning if the user is over quota
			MaxConcurrent int  `envconfig:"SUBSCRIPTION_QUOTAS_FINETUNING_FREE_MAX_CONCURRENT" default:"1"`
			MaxChunks     int  `envconfig:"SUBSCRIPTION_QUOTAS_FINETUNING_FREE_MAX_CHUNKS" default:"5"`
		}
		Pro struct {
			Strict        bool `envconfig:"SUBSCRIPTION_QUOTAS_FINETUNING_PRO_STRICT" default:"false"` // If set, will now allow any finetuning if the user is over quota
			MaxConcurrent int  `envconfig:"SUBSCRIPTION_QUOTAS_FINETUNING_PRO_MAX_CONCURRENT" default:"3"`
			MaxChunks     int  `envconfig:"SUBSCRIPTION_QUOTAS_FINETUNING_PRO_MAX_CHUNKS" default:"100"`
		}
	}
}

type GitHub struct {
	Enabled      bool   `envconfig:"GITHUB_INTEGRATION_ENABLED" default:"false" description:"Enable github integration."`
	ClientID     string `envconfig:"GITHUB_INTEGRATION_CLIENT_ID" description:"The github app client id."`
	ClientSecret string `envconfig:"GITHUB_INTEGRATION_CLIENT_SECRET" description:"The github app client secret."`
	RepoFolder   string `envconfig:"GITHUB_INTEGRATION_REPO_FOLDER" default:"/filestore/github/repos" description:"What folder do we use to clone github repos."`
	WebhookURL   string `envconfig:"GITHUB_INTEGRATION_WEBHOOK_URL" description:"The URL to receive github webhooks."`
}

type FineTuning struct {
	Enabled  bool           `envconfig:"FINETUNING_ENABLED" default:"true" description:"Enable QA pairs."` // Enable/disable QA pairs for the server
	Provider types.Provider `envconfig:"FINETUNING_PROVIDER" default:"togetherai" description:"Which LLM provider to use for QA pairs."`
	// Suggestions based on provider:
	// - Together AI: meta-llama/Llama-3-8b-chat-hf
	// - Helix: llama3:instruct
	QAPairGenModel string `envconfig:"FINETUNING_QA_PAIR_GEN_MODEL" default:"mistralai/Mixtral-8x7B-Instruct-v0.1" description:"Which LLM model to use for QA pairs."`
}

type Apps struct {
	Enabled  bool           `envconfig:"APPS_ENABLED" default:"true" description:"Enable apps."` // Enable/disable apps for the server
	Provider types.Provider `envconfig:"APPS_PROVIDER" default:"togetherai" description:"Which LLM provider to use for apps."`
	Model    string         `envconfig:"APPS_MODEL" default:"mistralai/Mixtral-8x7B-Instruct-v0.1" description:"Which LLM model to use for apps."` // gpt-4-1106-preview
}

type GPTScript struct {
	Enabled bool `envconfig:"GPTSCRIPT_ENABLED" default:"true" description:"Enable gptscript."` // Enable/disable gptscript for the server

	Runner struct {
		RequestTimeout time.Duration `envconfig:"GPTSCRIPT_RUNNER_REQUEST_TIMEOUT" default:"10s" description:"How long to wait for the script response."`
		Retries        uint          `envconfig:"GPTSCRIPT_RUNNER_RETRIES" default:"3" description:"How many retries."`
	}

	TestFaster struct {
		URL   string `envconfig:"HELIX_TESTFASTER_URL" description:"The URL to the testfaster cluster."`
		Token string `envconfig:"HELIX_TESTFASTER_TOKEN"`
	}
}

type Triggers struct {
	Discord Discord
	Cron    Cron
}

type Discord struct {
	Enabled  bool   `envconfig:"DISCORD_ENABLED" default:"false"`
	BotToken string `envconfig:"DISCORD_BOT_TOKEN"`
}

type Cron struct {
	Enabled bool `envconfig:"CRON_ENABLED" default:"true"`
}

type SSL struct {
	// certFileEnv is the environment variable which identifies where to locate
	// the SSL certificate file. If set this overrides the system default.
	SSLCertFile string `envconfig:"SSL_CERT_FILE"`

	// certDirEnv is the environment variable which identifies which directory
	// to check for SSL certificate files. If set this overrides the system default.
	// It is a colon separated list of directories.
	// See https://www.openssl.org/docs/man1.0.2/man1/c_rehash.html.
	SSLCertDir string `envconfig:"SSL_CERT_DIR"`
}
