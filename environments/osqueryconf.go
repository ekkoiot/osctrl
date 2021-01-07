package environments

import (
	"encoding/json"
	"fmt"
)

// OsqueryConf to hold the structure for the configuration
// https://osquery.readthedocs.io/en/stable/deployment/configuration/#configuration-specification
type OsqueryConf struct {
	Options    OptionsConf   `json:"options"`
	Schedule   ScheduleConf  `json:"schedule"`
	Packs      PacksConf     `json:"packs"`
	Decorators DecoratorConf `json:"decorators"`
	ATC        ATCConf       `json:"auto_table_construction"`
}

// OptionsConf for each part of the configuration
type OptionsConf map[string]interface{}

// ScheduleConf to hold all the schedule
// https://osquery.readthedocs.io/en/stable/deployment/configuration/#schedule
type ScheduleConf map[string]ScheduleQuery

// ScheduleQuery to hold the scheduled queries in the configuration
// https://osquery.readthedocs.io/en/stable/deployment/configuration/#schedule
type ScheduleQuery struct {
	Query    string `json:"query,omitempty"`
	Interval int    `json:"interval,omitempty"`
	Removed  bool   `json:"removed,omitempty"`
	Snapshot bool   `json:"snapshot,omitempty"`
	Platform string `json:"platform,omitempty"`
	Version  string `json:"version,omitempty"`
	Shard    int    `json:"shard,omitempty"`
	Denylist bool   `json:"denylist,omitempty"`
}

// PacksConf to hold all the packs in the configuration
// https://osquery.readthedocs.io/en/stable/deployment/configuration/#packs
type PacksConf map[string]interface{}

// PackEntry to hold the struct for a single pack
type PackEntry struct {
	Queries   map[string]ScheduleQuery `json:"queries,omitempty"`
	Platform  string                   `json:"platform,omitempty"`
	Shard     int                      `json:"shard,omitempty"`
	Version   string                   `json:"version,omitempty"`
	Discovery []string                 `json:"discovery,omitempty"`
}

// DecoratorConf to hold the osquery decorators
// https://osquery.readthedocs.io/en/stable/deployment/configuration/#decorator-queries
type DecoratorConf struct {
	Load     []string    `json:"load,omitempty"`
	Always   []string    `json:"always,omitempty"`
	Interval interface{} `json:"interval,omitempty"`
}

// ATCConf to hold all the auto table construction in the configuration
// https://osquery.readthedocs.io/en/stable/deployment/configuration/#automatic-table-construction
type ATCConf map[string]interface{}

// RefreshConfiguration to take all parts and put them together in the configuration
func (environment *Environment) RefreshConfiguration(name string) error {
	env, err := environment.Get(name)
	if err != nil {
		return fmt.Errorf("error structuring environment %v", err)
	}
	_options, err := environment.GenStructOptions([]byte(env.Options))
	if err != nil {
		return fmt.Errorf("error structuring options %v", err)
	}
	_schedule, err := environment.GenStructSchedule([]byte(env.Schedule))
	if err != nil {
		return fmt.Errorf("error structuring schedule %v", err)
	}
	_packs, err := environment.GenStructPacks([]byte(env.Packs))
	if err != nil {
		return fmt.Errorf("error structuring packs %v", err)
	}
	_decorators, err := environment.GenStructDecorators([]byte(env.Decorators))
	if err != nil {
		return fmt.Errorf("error structuring decorators %v", err)
	}
	_ATC, err := environment.GenStructATC([]byte(env.ATC))
	if err != nil {
		return fmt.Errorf("error structuring ATC %v", err)
	}
	conf := OsqueryConf{
		Options:    _options,
		Schedule:   _schedule,
		Packs:      _packs,
		Decorators: _decorators,
		ATC:        _ATC,
	}
	indentedConf, err := environment.GenSerializedConf(conf, true)
	if err != nil {
		return fmt.Errorf("error serializing configuration %v", err)
	}
	if err := environment.DB.Model(&env).Update("configuration", indentedConf).Error; err != nil {
		return fmt.Errorf("Update configuration %v", err)
	}
	return nil
}

// UpdateConfiguration to update configuration for an environment
func (environment *Environment) UpdateConfiguration(name string, cnf OsqueryConf) error {
	indentedConf, err := environment.GenSerializedConf(cnf, true)
	if err != nil {
		return fmt.Errorf("error serializing configuration %v", err)
	}
	if err := environment.DB.Model(&TLSEnvironment{}).Where("name = ?", name).Update("configuration", indentedConf).Error; err != nil {
		return fmt.Errorf("Update configuration %v", err)
	}
	return nil
}

// UpdateConfigurationParts to update all the configuration parts for an environment
func (environment *Environment) UpdateConfigurationParts(name string, cnf OsqueryConf) error {
	indentedOptions, err := environment.GenSerializedConf(cnf.Options, true)
	if err != nil {
		return fmt.Errorf("error serializing options %v", err)
	}
	indentedSchedule, err := environment.GenSerializedConf(cnf.Schedule, true)
	if err != nil {
		return fmt.Errorf("error serializing schedule %v", err)
	}
	indentedPacks, err := environment.GenSerializedConf(cnf.Packs, true)
	if err != nil {
		return fmt.Errorf("error serializing packs %v", err)
	}
	indentedDecorators, err := environment.GenSerializedConf(cnf.Decorators, true)
	if err != nil {
		return fmt.Errorf("error serializing decorators %v", err)
	}
	indentedATC, err := environment.GenSerializedConf(cnf.ATC, true)
	if err != nil {
		return fmt.Errorf("error serializing ATC %v", err)
	}
	if err := environment.DB.Model(&TLSEnvironment{}).Where("name = ?", name).Updates(TLSEnvironment{
		Options:    indentedOptions,
		Schedule:   indentedSchedule,
		Packs:      indentedPacks,
		Decorators: indentedDecorators,
		ATC:        indentedATC}).Error; err != nil {
		return fmt.Errorf("Update parts %v", err)
	}
	return nil
}

// GenSerializedConf to generate a serialized osquery configuration from the structured data
func (environment *Environment) GenSerializedConf(structured interface{}, indent bool) (string, error) {
	indentStr := ""
	if indent {
		indentStr = "  "
	}
	jsonConf, err := json.MarshalIndent(structured, "", indentStr)
	if err != nil {
		return "", err
	}
	return string(jsonConf), nil
}

// GenStructConf to generate the components from the osquery configuration
func (environment *Environment) GenStructConf(configuration []byte) (OsqueryConf, error) {
	var data OsqueryConf
	if err := json.Unmarshal(configuration, &data); err != nil {
		return data, err
	}
	return data, nil
}

// GenStructOptions to generate options from the serialized string
func (environment *Environment) GenStructOptions(configuration []byte) (OptionsConf, error) {
	var data OptionsConf
	if err := json.Unmarshal(configuration, &data); err != nil {
		return data, err
	}
	return data, nil
}

// GenStructSchedule to generate schedule from the serialized string
func (environment *Environment) GenStructSchedule(configuration []byte) (ScheduleConf, error) {
	var data ScheduleConf
	if err := json.Unmarshal(configuration, &data); err != nil {
		return data, err
	}
	return data, nil
}

// GenStructPacks to generate packs from the serialized string
func (environment *Environment) GenStructPacks(configuration []byte) (PacksConf, error) {
	var data PacksConf
	if err := json.Unmarshal(configuration, &data); err != nil {
		return data, err
	}
	return data, nil
}

// GenStructDecorators to generate decorators from the serialized string
func (environment *Environment) GenStructDecorators(configuration []byte) (DecoratorConf, error) {
	var data DecoratorConf
	if err := json.Unmarshal(configuration, &data); err != nil {
		return data, err
	}
	return data, nil
}

// GenStructATC to generate ATC from the serialized string
func (environment *Environment) GenStructATC(configuration []byte) (ATCConf, error) {
	var data ATCConf
	if err := json.Unmarshal(configuration, &data); err != nil {
		return data, err
	}
	return data, nil
}

// GenEmptyConfiguration to generate a serialized string with an empty configuration
func (environment *Environment) GenEmptyConfiguration(indent bool) string {
	cnf := OsqueryConf{
		Options:  OptionsConf{},
		Schedule: ScheduleConf{},
		Packs:    PacksConf{},
		Decorators: DecoratorConf{
			Always: []string{
				DecoratorUsers,
				DecoratorHostname,
				DecoratorLoggedInUser,
				DecoratorOsqueryVersionHash,
				DecoratorMD5Process,
			},
		},
		ATC: ATCConf{},
	}
	str, err := environment.GenSerializedConf(cnf, indent)
	if err != nil {
		return ""
	}
	return str
}

// AddOptionsConf to add a new query to the osquery schedule
func (environment *Environment) AddOptionsConf(name, option string, value interface{}) error {
	env, err := environment.Get(name)
	if err != nil {
		return fmt.Errorf("error structuring environment %v", err)
	}
	// Parse options into struct
	_options, err := environment.GenStructOptions([]byte(env.Options))
	if err != nil {
		return fmt.Errorf("error structuring options %v", err)
	}
	// Add new option
	_options[option] = value
	// Generate serialized indented options
	indentedOptions, err := environment.GenSerializedConf(_options, true)
	if err != nil {
		return fmt.Errorf("error serializing options %v", err)
	}
	// Update options in environment
	if err := environment.UpdateOptions(name, indentedOptions); err != nil {
		return fmt.Errorf("error updating options %v", err)
	}
	// Refresh all configuration
	if err := environment.RefreshConfiguration(name); err != nil {
		return fmt.Errorf("error refreshing configuration %v", err)
	}
	return nil
}

// AddScheduleConfQuery to add a new query to the osquery schedule
func (environment *Environment) AddScheduleConfQuery(name, qName string, query ScheduleQuery) error {
	env, err := environment.Get(name)
	if err != nil {
		return fmt.Errorf("error structuring environment %v", err)
	}
	// Parse options into struct
	_schedule, err := environment.GenStructSchedule([]byte(env.Schedule))
	if err != nil {
		return fmt.Errorf("error structuring schedule %v", err)
	}
	// Add new query
	_schedule[qName] = query
	// Generate serialized indented schedule
	indentedSchedule, err := environment.GenSerializedConf(_schedule, true)
	if err != nil {
		return fmt.Errorf("error serializing schedule %v", err)
	}
	// Update schedule in environment
	if err := environment.UpdateSchedule(name, indentedSchedule); err != nil {
		return fmt.Errorf("error updating schedule %v", err)
	}
	// Refresh all configuration
	if err := environment.RefreshConfiguration(name); err != nil {
		return fmt.Errorf("error refreshing configuration %v", err)
	}
	return nil
}
