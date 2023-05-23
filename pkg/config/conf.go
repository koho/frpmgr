package config

import (
	"os"
	"time"

	"gopkg.in/ini.v1"

	"github.com/koho/frpmgr/pkg/consts"
	"github.com/koho/frpmgr/pkg/util"
)

func init() {
	ini.PrettyFormat = false
	ini.PrettyEqual = true
}

// Config is the interface that a config must implement to support management.
type Config interface {
	// Items returns all sections in this config. which must be a slice of pointer to struct.
	Items() interface{}
	// ItemAt returns the section in this config for the given index.
	ItemAt(index int) interface{}
	// DeleteItem deletes the section for the given index.
	DeleteItem(index int)
	// AddItem adds a section to this config.
	AddItem(item interface{}) bool
	// Save serializes this config and saves to the given path.
	Save(path string) error
	// Complete prunes and completes this config.
	// When "read" is true, the config should be completed for a file loaded from source.
	// Otherwise, it should be completed for file written to disk.
	Complete(read bool)
	// GetLogFile returns the log file path of this config.
	GetLogFile() string
	// AutoStart indicates whether this config should be started at boot.
	AutoStart() bool
	// Expiry indicates whether the config has an expiry date.
	Expiry() bool
	// Copy creates a new copy of this config.
	Copy(all bool) Config
}

type AutoDelete struct {
	// DeleteMethod specifies what delete method to use to delete the config.
	// If "absolute" is specified, the expiry date is set in config. If "relative" is specified, the expiry date
	// is calculated by adding the days to the file modification time. If it's empty, the config has no expiry date.
	DeleteMethod string `ini:"frpmgr_delete_method,omitempty"`
	// DeleteAfterDays is the number of days a config will be kept, after which it may be stopped and deleted.
	DeleteAfterDays uint `ini:"frpmgr_delete_after_days,omitempty" relative:"true"`
	// DeleteAfterDate is the last date the config will be valid, after which it may be stopped and deleted.
	DeleteAfterDate time.Time `ini:"frpmgr_delete_after_date,omitempty" absolute:"true"`
}

func (ad AutoDelete) Complete() AutoDelete {
	deleteMethod := ad.DeleteMethod
	if deleteMethod != "" {
		if d, err := util.PruneByTag(ad, "true", deleteMethod); err == nil {
			ad = d.(AutoDelete)
			ad.DeleteMethod = deleteMethod
		}
		// Reset zero day
		if deleteMethod == consts.DeleteRelative && ad.DeleteAfterDays == 0 {
			ad.DeleteMethod = ""
		}
	} else {
		ad = AutoDelete{}
	}
	return ad
}

// Expiry returns the remaining duration, after which a config will expire.
// If a config has no expiry date, an `ErrNoDeadline` error is returned.
func Expiry(configPath string, del AutoDelete) (time.Duration, error) {
	fInfo, err := os.Stat(configPath)
	if err != nil {
		return 0, err
	}
	switch del.DeleteMethod {
	case consts.DeleteAbsolute:
		return time.Until(del.DeleteAfterDate), nil
	case consts.DeleteRelative:
		if del.DeleteAfterDays > 0 {
			elapsed := time.Since(fInfo.ModTime())
			total := time.Hour * 24 * time.Duration(del.DeleteAfterDays)
			return total - elapsed, nil
		}
	}
	return 0, os.ErrNoDeadline
}
