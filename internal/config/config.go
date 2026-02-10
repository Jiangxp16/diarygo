package config

import (
	"crypto/md5"
	"diarygo/internal/db"
	"diarygo/internal/entity/bill"
	"diarygo/internal/entity/diary"
	"diarygo/internal/entity/interest"
	"diarygo/internal/entity/note"
	"diarygo/internal/entity/sport"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"sync"

	"gopkg.in/ini.v1"
)

func configPath() string {
	if p := os.Getenv("DIARYGO_PATH"); p != "" {
		return p
	}
	return "config/config.ini"
}

type Repository struct {
	filePath string
	cfg      *ini.File
	mutex    sync.Mutex
}

type ConfigRule struct {
	AllowedValues []string
	MaxLen        int
}

var defaultConfig = map[string]map[string]string{
	"global": {
		"db_name":           "data/diary.db",
		"port":              "8080",
		"first_day_of_week": "1",
		"location":          "",
		"language":          "en",
		"login_expired":     "1h", //seconds
		"show_lunar":        "1",
		"show_bill":         "1",
		"show_note":         "1",
		"show_interest":     "1",
		"show_sport":        "1",
		"password":          "",
		"ui_default":        "diary",
	},
	"style": {
		"font": "Arial, sans-serif",
		"logo": "/static/imgs/logo.png",
	},
	"backup": {
		"interval": "24h",
		"keep":     "7",
		"dir":      "data/backup",
	},
}

var editableConfig = map[string]map[string]ConfigRule{
	"global": {
		"first_day_of_week": {
			AllowedValues: []string{"0", "1"},
		},
		"location": {},
		"language": {
			AllowedValues: []string{"en", "zh"},
		},
		"login_expired": {},
		"show_lunar": {
			AllowedValues: []string{"0", "1"},
		},
		"show_bill": {
			AllowedValues: []string{"0", "1"},
		},
		"show_note": {
			AllowedValues: []string{"0", "1"},
		},
		"show_interest": {
			AllowedValues: []string{"0", "1"},
		},
		"show_sport": {
			AllowedValues: []string{"0", "1"},
		},
		"ui_default": {
			AllowedValues: []string{"config", "diary", "bill", "interest", "note"},
		},
	},
	"style": {
		"font": {MaxLen: 32},
		"logo": {},
	},
	"backup": {
		"interval": {},
		"keep":     {},
		"dir":      {},
	},
}

var (
	instance *Repository
	once     sync.Once
	initErr  error
)

func GetRepository() *Repository {
	once.Do(func() {
		instance, initErr = newRepository()
	})
	if initErr != nil {
		panic(initErr)
	}
	return instance
}

// NewRepository 读取或创建配置文件
func newRepository() (*Repository, error) {
	var err error
	r := &Repository{filePath: configPath()}

	if _, err = os.Stat(r.filePath); os.IsNotExist(err) {
		r.cfg = ini.Empty()
		for sectionName, keys := range defaultConfig {
			section, _ := r.cfg.NewSection(sectionName)
			for key, value := range keys {
				section.Key(key).SetValue(value)
			}
		}
		err = r.cfg.SaveTo(r.filePath)
		if err != nil {
			return nil, err
		}
	} else {
		r.cfg, err = ini.Load(r.filePath)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Repository) CheckValid(section, key, value string) error {
	secRules, ok := editableConfig[section]
	if !ok {
		return errors.New("Section not editable: " + section)
	}
	keyRule, ok := secRules[key]
	if !ok {
		return errors.New("Key not editable: " + key)
	}
	if len(keyRule.AllowedValues) > 0 {
		valid := false
		for _, v := range keyRule.AllowedValues {
			if value == v {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("Invalid value: " + value)
		}
	}
	if keyRule.MaxLen > 0 && len(value) > keyRule.MaxLen {
		return errors.New("Value too long: " + value)
	}
	return nil
}

func (r *Repository) Get(section, key string) string {
	def := defaultConfig[section][key]
	return r.GetWithDefault(section, key, def)
}

func (r *Repository) GetWithDefault(section, key, def string) string {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	sec := r.cfg.Section(section)
	if sec == nil {
		return def
	}

	keyObj := sec.Key(key)
	if keyObj == nil {
		return def
	}

	val := keyObj.String()
	return val
}

func (r *Repository) GetBool(section, key string) bool {
	v := r.Get(section, key)
	return v == "1" || v == "true" || v == "yes"
}

func (r *Repository) GetInt(section, key string, def int) int {
	v := r.Get(section, key)
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func (r *Repository) Set(section, key, value string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.cfg.Section(section).Key(key).SetValue(value)
	return r.cfg.SaveTo(r.filePath)
}

func (r *Repository) Delete(section, key string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.cfg.Section(section).DeleteKey(key)
	return r.cfg.SaveTo(r.filePath)
}

// -------------------- Password Operation --------------------

func (r *Repository) SetPassword(password string) error {
	hash := md5.Sum([]byte(password))
	hashStr := hex.EncodeToString(hash[:])
	return r.Set("global", "password", hashStr)
}

func (r *Repository) GetPassword() string {
	return r.Get("global", "password")
}

func (r *Repository) CheckPassword(input string) bool {
	hash := md5.Sum([]byte(input))
	hashStr := hex.EncodeToString(hash[:])
	return r.GetPassword() == hashStr
}

func (r *Repository) ChangePassword(d *db.DB, oldPwd, newPwd string) error {
	db.GlobalWriteMutex.Lock()
	defer db.GlobalWriteMutex.Unlock()

	if !r.CheckPassword(oldPwd) {
		return errors.New("old password incorrect")
	}

	diaryList, err := diary.NewRepository(d).List()
	if err != nil {
		return err
	}
	billList, err := bill.NewRepository(d).List()
	if err != nil {
		return err
	}
	interestList, err := interest.NewRepository(d).List()
	if err != nil {
		return err
	}
	noteList, err := note.NewRepository(d).List()
	if err != nil {
		return err
	}
	sportList, err := sport.NewRepository(d).List()
	if err != nil {
		return err
	}

	db.Key = newPwd

	if len(diaryList) > 0 {
		if err := diary.NewRepository(d).UpdateMany(diaryList); err != nil {
			return err
		}
	}
	if len(billList) > 0 {
		if err := bill.NewRepository(d).UpdateMany(billList); err != nil {
			return err
		}
	}
	if len(interestList) > 0 {
		if err := interest.NewRepository(d).UpdateMany(interestList); err != nil {
			return err
		}
	}
	if len(noteList) > 0 {
		if err := note.NewRepository(d).UpdateMany(noteList); err != nil {
			return err
		}
	}
	if len(sportList) > 0 {
		if err := sport.NewRepository(d).UpdateMany(sportList); err != nil {
			return err
		}
	}
	if err := r.SetPassword(newPwd); err != nil {
		return err
	}
	return nil
}
