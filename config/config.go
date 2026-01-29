package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

var (
	// Exist Is config file exist
	Exist bool
	// Models configs
	Models []ModelConfig
	// gobackup base dir
	GoBackupDir string = getGoBackupDir()

	PidFilePath string = filepath.Join(GoBackupDir, "gobackup.pid")
	LogFilePath string = filepath.Join(GoBackupDir, "gobackup.log")
	Web         WebConfig

	wLock = sync.Mutex{}

	// The config file loaded at
	UpdatedAt time.Time

	onConfigChanges = make([]func(fsnotify.Event), 0)
)

// WebConfig holds the web server configuration
type WebConfig struct {
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	Username string `json:"-"` // exclude from JSON for security
	Password string `json:"-"` // exclude from JSON for security
	Enabled  bool   `json:"enabled,omitempty"`
}

// HasAuth returns true if web authentication is configured
func (w WebConfig) HasAuth() bool {
	return len(w.Username) > 0 && len(w.Password) > 0
}

// Address returns the full address string (host:port)
func (w WebConfig) Address() string {
	return fmt.Sprintf("%s:%s", w.Host, w.Port)
}

// ScheduleConfig holds the scheduling configuration for a backup model
type ScheduleConfig struct {
	Enabled bool   `json:"enabled,omitempty"`
	Cron    string `json:"cron,omitempty"`  // Cron expression (e.g., "0 0 * * *")
	Every   string `json:"every,omitempty"` // Duration (e.g., "1day", "12h")
	At      string `json:"at,omitempty"`    // Time of day (e.g., "03:00")
}

// String returns a human-readable representation of the schedule
func (sc ScheduleConfig) String() string {
	if !sc.Enabled {
		return "disabled"
	}
	if len(sc.Cron) > 0 {
		return fmt.Sprintf("cron %s", sc.Cron)
	}
	if len(sc.At) > 0 {
		return fmt.Sprintf("every %s at %s", sc.Every, sc.At)
	}
	return fmt.Sprintf("every %s", sc.Every)
}

// Validate checks if the schedule configuration is valid
func (sc ScheduleConfig) Validate() error {
	if !sc.Enabled {
		return nil
	}
	// Must have either cron or every
	if len(sc.Cron) == 0 && len(sc.Every) == 0 {
		return fmt.Errorf("schedule must have either 'cron' or 'every' configured")
	}
	// Cannot have both cron and every
	if len(sc.Cron) > 0 && len(sc.Every) > 0 {
		return fmt.Errorf("schedule cannot have both 'cron' and 'every' configured")
	}
	return nil
}

// ModelConfig represents a backup model configuration
type ModelConfig struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	WorkDir     string         `json:"-"` // runtime, not serialized
	TempPath    string         `json:"-"` // runtime, not serialized
	DumpPath    string         `json:"-"` // runtime, not serialized
	Schedule    ScheduleConfig `json:"schedule,omitempty"`

	CompressWith   SubConfig `json:"compress_with,omitempty"`
	EncryptWith    SubConfig `json:"encrypt_with,omitempty"`
	DefaultStorage string    `json:"default_storage,omitempty"`

	// These use viper.Viper internally, exclude from JSON
	Archive  *viper.Viper `json:"-"`
	Splitter *viper.Viper `json:"-"`
	Viper    *viper.Viper `json:"-"`

	Databases map[string]SubConfig `json:"databases,omitempty"`
	Storages  map[string]SubConfig `json:"storages,omitempty"`
	Notifiers map[string]SubConfig `json:"notifiers,omitempty"`

	BeforeScript string `json:"before_script,omitempty"`
	AfterScript  string `json:"after_script,omitempty"`
}

// String returns a human-readable representation of the model
func (m ModelConfig) String() string {
	return fmt.Sprintf("Model{name=%s, databases=%d, storages=%d, schedule=%s}",
		m.Name, len(m.Databases), len(m.Storages), m.Schedule.String())
}

// HasSchedule returns true if the model has scheduling enabled
func (m ModelConfig) HasSchedule() bool {
	return m.Schedule.Enabled
}

// HasEncryption returns true if encryption is configured
func (m ModelConfig) HasEncryption() bool {
	return len(m.EncryptWith.Type) > 0
}

// DatabaseNames returns a list of configured database names
func (m ModelConfig) DatabaseNames() []string {
	names := make([]string, 0, len(m.Databases))
	for name := range m.Databases {
		names = append(names, name)
	}
	return names
}

// StorageNames returns a list of configured storage names
func (m ModelConfig) StorageNames() []string {
	names := make([]string, 0, len(m.Storages))
	for name := range m.Storages {
		names = append(names, name)
	}
	return names
}

func getGoBackupDir() string {
	dir := os.Getenv("GOBACKUP_DIR")
	if len(dir) == 0 {
		dir = filepath.Join(os.Getenv("HOME"), ".gobackup")
	}
	return dir
}

// SubConfig holds configuration for a sub-component (database, storage, notifier)
type SubConfig struct {
	Name  string       `json:"name,omitempty"`
	Type  string       `json:"type,omitempty"`
	Viper *viper.Viper `json:"-"` // internal, not serialized
}

// String returns a human-readable representation
func (s SubConfig) String() string {
	return fmt.Sprintf("%s(%s)", s.Name, s.Type)
}

// Init
// loadConfig from:
// - ./gobackup.yml
// - ~/.gobackup/gobackup.yml
// - /etc/gobackup/gobackup.yml
func Init(configFile string) error {
	logger := logger.Tag("Config")

	viper.SetConfigType("yaml")

	// set config file directly
	if len(configFile) > 0 {
		configFile = helper.AbsolutePath(configFile)
		logger.Info("Load config:", configFile)

		viper.SetConfigFile(configFile)
	} else {
		logger.Info("Load config from default path.")
		viper.SetConfigName("gobackup")

		// ./gobackup.yml
		viper.AddConfigPath(".")
		// ~/.gobackup/gobackup.yml
		viper.AddConfigPath("$HOME/.gobackup") // call multiple times to add many search paths
		// /etc/gobackup/gobackup.yml
		viper.AddConfigPath("/etc/gobackup/") // path to look for the config file in
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		logger.Info("Config file changed:", in.Name)
		defer onConfigChanged(in)
		if err := loadConfig(); err != nil {
			logger.Error(err.Error())
		}
	})

	return loadConfig()
}

// OnConfigChange add callback when config changed
func OnConfigChange(run func(in fsnotify.Event)) {
	onConfigChanges = append(onConfigChanges, run)
}

// Invoke callbacks when config changed
func onConfigChanged(in fsnotify.Event) {
	for _, fn := range onConfigChanges {
		fn(in)
	}
}

func loadConfig() error {
	wLock.Lock()
	defer wLock.Unlock()

	logger := logger.Tag("Config")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("Load gobackup config failed: ", err)
		return err
	}

	viperConfigFile := viper.ConfigFileUsed()
	if info, err := os.Stat(viperConfigFile); err == nil {
		// max permission: 0770
		if info.Mode()&(1<<2) != 0 {
			logger.Warnf("Other users are able to access %s with mode %v", viperConfigFile, info.Mode())
		}
	}

	logger.Info("Config file:", viperConfigFile)

	// load .env if exists in the same directory of used config file and expand variables in the config
	dotEnv := filepath.Join(filepath.Dir(viperConfigFile), ".env")
	if _, err := os.Stat(dotEnv); err == nil {
		if err := godotenv.Load(dotEnv); err != nil {
			logger.Errorf("Load %s failed: %v", dotEnv, err)
			return err
		}
	}

	cfg, _ := os.ReadFile(viperConfigFile)
	if err := viper.ReadConfig(strings.NewReader(os.ExpandEnv(string(cfg)))); err != nil {
		logger.Errorf("Load expanded config failed: %v", err)
		return err
	}

	// Setup working directory for temporary backup files
	// If not configured, create a temporary directory that will be cleaned up after backup
	viper.Set("useTempWorkDir", false)
	if workdir := viper.GetString("workdir"); len(workdir) == 0 {
		dir, err := os.MkdirTemp("", "gobackup")
		if err != nil {
			return fmt.Errorf("failed to create temp workdir: %w", err)
		}
		viper.Set("workdir", dir)
		viper.Set("useTempWorkDir", true)
	}

	Exist = true
	Models = []ModelConfig{}
	for key := range viper.GetStringMap("models") {
		model, err := loadModel(key)
		if err != nil {
			return fmt.Errorf("load model %s: %v", key, err)
		}

		Models = append(Models, model)
	}

	if len(Models) == 0 {
		return fmt.Errorf("no model found in %s", viperConfigFile)
	}

	// Load web config
	Web = WebConfig{}
	viper.SetDefault("web.host", "0.0.0.0")
	viper.SetDefault("web.port", 2703)
	viper.SetDefault("web.enabled", true)
	Web.Host = viper.GetString("web.host")
	Web.Port = viper.GetString("web.port")
	Web.Username = viper.GetString("web.username")
	Web.Password = viper.GetString("web.password")
	Web.Enabled = viper.GetBool("web.enabled")

	UpdatedAt = time.Now()
	logger.Infof("Config loaded, found %d models.", len(Models))

	return nil
}

func loadModel(key string) (ModelConfig, error) {
	var model ModelConfig
	model.Name = key

	workdir, _ := os.Getwd()

	model.WorkDir = workdir
	model.TempPath = filepath.Join(viper.GetString("workdir"), fmt.Sprintf("%d", time.Now().UnixNano()))
	model.DumpPath = filepath.Join(model.TempPath, key)
	model.Viper = viper.Sub("models." + key)

	model.Description = model.Viper.GetString("description")
	model.Schedule = ScheduleConfig{Enabled: false}

	compressViper := model.Viper.Sub("compress_with")
	if compressViper == nil {
		compressViper = viper.New()
	}
	compressViper.SetDefault("type", "tar")
	compressViper.SetDefault("filename_format", "2006.01.02.15.04.05")
	model.CompressWith = SubConfig{
		Type:  compressViper.GetString("type"),
		Viper: compressViper,
	}

	model.EncryptWith = SubConfig{
		Type:  model.Viper.GetString("encrypt_with.type"),
		Viper: model.Viper.Sub("encrypt_with"),
	}

	model.Archive = model.Viper.Sub("archive")
	model.Splitter = model.Viper.Sub("split_with")

	model.BeforeScript = model.Viper.GetString("before_script")
	model.AfterScript = model.Viper.GetString("after_script")

	if err := loadScheduleConfig(&model); err != nil {
		return ModelConfig{}, fmt.Errorf("model %s: %w", model.Name, err)
	}
	loadDatabasesConfig(&model)
	loadStoragesConfig(&model)

	if len(model.Storages) == 0 {
		return ModelConfig{}, fmt.Errorf("model %s: no storage configured", model.Name)
	}

	loadNotifiersConfig(&model)

	return model, nil
}

func loadScheduleConfig(model *ModelConfig) error {
	subViper := model.Viper.Sub("schedule")
	model.Schedule = ScheduleConfig{Enabled: false}
	if subViper == nil {
		return nil
	}

	model.Schedule = ScheduleConfig{
		Enabled: true,
		Cron:    subViper.GetString("cron"),
		Every:   subViper.GetString("every"),
		At:      subViper.GetString("at"),
	}

	return model.Schedule.Validate()
}

func loadDatabasesConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("databases")
	model.Databases = map[string]SubConfig{}
	for key := range model.Viper.GetStringMap("databases") {
		dbViper := subViper.Sub(key)
		model.Databases[key] = SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		}
	}
}

func loadStoragesConfig(model *ModelConfig) {
	storageConfigs := map[string]SubConfig{}

	model.DefaultStorage = model.Viper.GetString("default_storage")

	subViper := model.Viper.Sub("storages")
	for key := range model.Viper.GetStringMap("storages") {
		storageViper := subViper.Sub(key)
		storageConfigs[key] = SubConfig{
			Name:  key,
			Type:  storageViper.GetString("type"),
			Viper: storageViper,
		}

		// Set default storage
		if len(model.DefaultStorage) == 0 {
			model.DefaultStorage = key
		}
	}
	model.Storages = storageConfigs

}

func loadNotifiersConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("notifiers")
	model.Notifiers = map[string]SubConfig{}
	for key := range model.Viper.GetStringMap("notifiers") {
		dbViper := subViper.Sub(key)
		model.Notifiers[key] = SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		}
	}
}

// GetModelConfigByName get model config by name
func GetModelConfigByName(name string) (model *ModelConfig) {
	for _, m := range Models {
		if m.Name == name {
			model = &m
			return
		}
	}
	return
}

// GetDatabaseByName get database config by name
func (model *ModelConfig) GetDatabaseByName(name string) (subConfig *SubConfig) {
	for _, m := range model.Databases {
		if m.Name == name {
			subConfig = &m
			return
		}
	}
	return
}
