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

type WebConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Enabled  bool
}

type ScheduleConfig struct {
	Enabled bool `json:"enabled,omitempty"`
	// Cron expression
	Cron string `json:"cron,omitempty"`
	// Every
	Every string `json:"every,omitempty"`
	// At time
	At string `json:"at,omitempty"`
}

func (sc ScheduleConfig) String() string {
	if sc.Enabled {
		if len(sc.Cron) > 0 {
			return fmt.Sprintf("cron %s", sc.Cron)
		} else {
			if len(sc.At) > 0 {
				return fmt.Sprintf("every %s at %s", sc.Every, sc.At)
			} else {
				return fmt.Sprintf("every %s", sc.Every)
			}
		}
	}

	return "disabled"
}

// ModelConfig for special case
type ModelConfig struct {
	Name        string
	Description string
	// WorkDir of the gobackup started
	WorkDir        string
	TempPath       string
	DumpPath       string
	Schedule       ScheduleConfig
	CompressWith   SubConfig
	EncryptWith    SubConfig
	Archive        *viper.Viper
	Splitter       *viper.Viper
	Databases      map[string]SubConfig
	Storages       map[string]SubConfig
	DefaultStorage string
	Notifiers      map[string]SubConfig
	Viper          *viper.Viper
	BeforeScript   string
	AfterScript    string
	// Hooks around the whole model run
	Hooks []HookConfig
	// ArchiveHooks around the archive stage (from `archive.hooks`)
	ArchiveHooks []HookConfig
	// SplitterHooks around the split stage (from `split_with.hooks`)
	SplitterHooks []HookConfig
}

const (
	HookBefore = "before"
	HookAfter  = "after"

	HookOnExitAlways  = "always"
	HookOnExitSuccess = "success"
	HookOnExitFailure = "failure"
)

// HookConfig describes a user script executed before or after a unit of work
// (the whole model run, a stage, or a single database / storage step).
type HookConfig struct {
	Name            string `mapstructure:"name"`
	When            string `mapstructure:"when"`
	OnExit          string `mapstructure:"on_exit"`
	Script          string `mapstructure:"script"`
	Shell           string `mapstructure:"shell"`
	Timeout         int    `mapstructure:"timeout"`
	ContinueOnError bool   `mapstructure:"continue_on_error"`
}

// DisplayName returns hook name for logs. Hooks loaded from config always have
// a name (`#N`, 1-based position in the list, when the user set none); hooks
// built in code without a name fall back to the given slice index.
func (h HookConfig) DisplayName(index int) string {
	if len(h.Name) > 0 {
		return h.Name
	}
	return fmt.Sprintf("#%d", index+1)
}

// loadHooks reads and validates `hooks` list from the given viper section.
// scope is used for error messages, e.g. `models.demo.databases.pg`.
func loadHooks(v *viper.Viper, scope string) ([]HookConfig, error) {
	if v == nil {
		return nil, nil
	}

	raw := v.Get("hooks")
	if raw == nil {
		return nil, nil
	}

	items, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("%s.hooks must be a list", scope)
	}

	var hooks []HookConfig
	if err := v.UnmarshalKey("hooks", &hooks); err != nil {
		return nil, fmt.Errorf("%s.hooks: %v", scope, err)
	}

	for i, hook := range hooks {
		if len(hook.Name) == 0 {
			hooks[i].Name = fmt.Sprintf("#%d", i+1)
		}
		name := fmt.Sprintf("%s.hooks[%s]", scope, hooks[i].Name)

		keys := map[string]bool{}
		if item, ok := items[i].(map[string]interface{}); ok {
			for k := range item {
				keys[strings.ToLower(k)] = true
			}
		}

		switch hook.When {
		case HookBefore:
			if keys["on_exit"] {
				return nil, fmt.Errorf("%s: on_exit is only allowed for after hooks", name)
			}
		case HookAfter:
			if keys["continue_on_error"] {
				return nil, fmt.Errorf("%s: continue_on_error is only allowed for before hooks", name)
			}
			switch hook.OnExit {
			case "", HookOnExitAlways, HookOnExitSuccess, HookOnExitFailure:
				// empty means always, defaulted by hook.RunAfter
			default:
				return nil, fmt.Errorf("%s: on_exit must be one of always, success, failure", name)
			}
		case "":
			return nil, fmt.Errorf("%s: when is required (before or after)", name)
		default:
			return nil, fmt.Errorf("%s: when must be before or after", name)
		}

		if len(strings.TrimSpace(hook.Script)) == 0 {
			return nil, fmt.Errorf("%s: script is required", name)
		}

		if hook.Timeout < 0 {
			return nil, fmt.Errorf("%s: timeout must not be negative", name)
		}
	}

	return hooks, nil
}

func getGoBackupDir() string {
	dir := os.Getenv("GOBACKUP_DIR")
	if len(dir) == 0 {
		dir = filepath.Join(os.Getenv("HOME"), ".gobackup")
	}
	return dir
}

// SubConfig sub config info
type SubConfig struct {
	Name  string
	Type  string
	Viper *viper.Viper
	// Hooks around this step (database / storage) or stage (compress / encrypt)
	Hooks []HookConfig
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

	// TODO: Here the `useTempWorkDir` and `workdir`, is not in config document. We need removed it.
	viper.Set("useTempWorkDir", false)
	if workdir := viper.GetString("workdir"); len(workdir) == 0 {
		// use temp dir as workdir
		dir, err := os.MkdirTemp("", "gobackup")
		if err != nil {
			return err
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
	if model.Viper == nil {
		return ModelConfig{}, fmt.Errorf("model %s is empty or not a map", key)
	}

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

	scope := "models." + key
	var err error
	if model.Hooks, err = loadHooks(model.Viper, scope); err != nil {
		return ModelConfig{}, err
	}
	if model.CompressWith.Hooks, err = loadHooks(model.CompressWith.Viper, scope+".compress_with"); err != nil {
		return ModelConfig{}, err
	}
	if model.EncryptWith.Hooks, err = loadHooks(model.EncryptWith.Viper, scope+".encrypt_with"); err != nil {
		return ModelConfig{}, err
	}
	if model.ArchiveHooks, err = loadHooks(model.Archive, scope+".archive"); err != nil {
		return ModelConfig{}, err
	}
	if model.SplitterHooks, err = loadHooks(model.Splitter, scope+".split_with"); err != nil {
		return ModelConfig{}, err
	}

	loadScheduleConfig(&model)
	if err = loadDatabasesConfig(&model); err != nil {
		return ModelConfig{}, err
	}
	if err = loadStoragesConfig(&model); err != nil {
		return ModelConfig{}, err
	}

	if len(model.Databases) == 0 && model.Archive == nil {
		return ModelConfig{}, fmt.Errorf("model %s must configure databases or archive", model.Name)
	}

	if len(model.Storages) == 0 {
		return ModelConfig{}, fmt.Errorf("no storage found in model %s", model.Name)
	}

	loadNotifiersConfig(&model)

	return model, nil
}

func loadScheduleConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("schedule")
	model.Schedule = ScheduleConfig{Enabled: false}
	if subViper == nil {
		return
	}

	model.Schedule = ScheduleConfig{
		Enabled: true,
		Cron:    subViper.GetString("cron"),
		Every:   subViper.GetString("every"),
		At:      subViper.GetString("at"),
	}
}

func loadDatabasesConfig(model *ModelConfig) error {
	subViper := model.Viper.Sub("databases")
	if subViper == nil {
		return nil
	}
	model.Databases = map[string]SubConfig{}
	for key := range model.Viper.GetStringMap("databases") {
		dbViper := subViper.Sub(key)
		hooks, err := loadHooks(dbViper, fmt.Sprintf("models.%s.databases.%s", model.Name, key))
		if err != nil {
			return err
		}
		model.Databases[key] = SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
			Hooks: hooks,
		}
	}
	return nil
}

func loadStoragesConfig(model *ModelConfig) error {
	storageConfigs := map[string]SubConfig{}

	model.DefaultStorage = model.Viper.GetString("default_storage")

	subViper := model.Viper.Sub("storages")
	if subViper == nil {
		model.Storages = storageConfigs
		return nil
	}
	for key := range model.Viper.GetStringMap("storages") {
		storageViper := subViper.Sub(key)
		hooks, err := loadHooks(storageViper, fmt.Sprintf("models.%s.storages.%s", model.Name, key))
		if err != nil {
			return err
		}
		storageConfigs[key] = SubConfig{
			Name:  key,
			Type:  storageViper.GetString("type"),
			Viper: storageViper,
			Hooks: hooks,
		}

		// Set default storage
		if len(model.DefaultStorage) == 0 {
			model.DefaultStorage = key
		}
	}
	model.Storages = storageConfigs

	return nil
}

func loadNotifiersConfig(model *ModelConfig) {
	subViper := model.Viper.Sub("notifiers")
	if subViper == nil {
		return
	}
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
