// Package urldata provides the UrlData type.
package model

import (
	"encoding/json"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"strings"
	"time"
)

// CustomTime denotes the expiration time in format dd/mm/yyyy hh:mm
type CustomTime struct {
	time.Time `gorm:"column:expires; type:datetime"`
}

// UrlData denotes the url data that is sent by the user.
type UrlData struct {
	ShortSlug string     `json:"short-slug" gorm:"column:short_slug; type:varchar(50); primary_key"`
	RealUrl   string     `json:"real-url" gorm:"column:real_url; type:text"`
	Expires   CustomTime `json:"expires" gorm:"embedded"`
}

// UnmarshalJSON overrides the base method to handle dd/mm/yyyy hh:mm
func (customTime *CustomTime) UnmarshalJSON(input []byte) error {
	strInput := string(input)
	strInput = strings.Trim(strInput, `"`)
	if strInput == "" {
		return nil
	}

	newTime, err := time.ParseInLocation("02/01/2006 15:04", strInput, time.Local)
	if err != nil {
		return err
	}

	customTime.Time = newTime
	return nil
}

// MarshalJSON overrides the base method to handle dd/mm/yyyy hh:mm
func (customTime *CustomTime) MarshalJSON() ([]byte, error) {
	timeAsJson, err := json.Marshal(customTime.Local().Format("02/01/2006 15:04"))
	if err != nil {
		return nil, err
	}

	return timeAsJson, nil
}
