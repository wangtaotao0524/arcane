package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	redactionMask     = "XXXXXXXXXX"
	keyAuthOidcConfig = "authOidcConfig"
)

type SettingVariable struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

// IsTrue returns true if the value is a truthy string
func (s SettingVariable) IsTrue() bool {
	ok, _ := strconv.ParseBool(s.Value)
	return ok
}

// AsInt returns the value as an integer
func (s SettingVariable) AsInt() int {
	val, _ := strconv.Atoi(s.Value)
	return val
}

// AsDurationSeconds returns the value as a time.Duration in seconds
func (s SettingVariable) AsDurationSeconds() time.Duration {
	val, err := strconv.Atoi(s.Value)
	if err != nil {
		return 0
	}
	return time.Duration(val) * time.Second
}

type Settings struct {
	// General category
	ProjectsDirectory SettingVariable `key:"projectsDirectory" meta:"label=Projects Directory;type=text;keywords=projects,directory,path,folder,location,storage,files,compose,docker-compose;category=general;description=Configure where project files are stored" catmeta:"id=general;title=General;icon=settings;url=/settings/general;description=Core application settings and configuration"`
	DiskUsagePath     SettingVariable `key:"diskUsagePath" meta:"label=Disk Usage Path;type=text;keywords=disk,usage,path,storage,folder,files;category=general;description=Path used for disk usage calculations"`
	BaseServerURL     SettingVariable `key:"baseServerUrl" meta:"label=Base Server URL;type=text;keywords=base,url,server,domain,host,endpoint,address,link;category=general;description=Set the base URL for the application"`
	EnableGravatar    SettingVariable `key:"enableGravatar" meta:"label=Enable Gravatar;type=boolean;keywords=gravatar,avatar,profile,picture,image,user,photo;category=general;description=Enable Gravatar profile pictures for users"`
	DefaultShell      SettingVariable `key:"defaultShell" meta:"label=Default Shell;type=text;keywords=shell,default,shellpath,path,login;category=general;description=Default shell to use for commands"`
	AccentColor       SettingVariable `key:"accentColor,public,local" meta:"label=Accent Color;type=text;keywords=color,accent,theme,css,appearance,ui;category=general;description=Primary accent color for UI"`

	// Deprecated: OnboardingCompleted is no longer used as of the onboarding removal.
	// This field is kept for backward compatibility and is automatically set to true on startup.
	// This will be removed in a future release.
	OnboardingCompleted SettingVariable `key:"onboardingCompleted,public" meta:"label=Onboarding Completed;type=boolean;keywords=onboarding,completed,setup,first-run;category=general;description=Whether onboarding has been completed"`

	// Deprecated: OnboardingSteps is no longer used as of the onboarding removal.
	// This field is kept for backward compatibility only.
	// This will be removed in a future release.
	OnboardingSteps SettingVariable `key:"onboardingSteps" meta:"label=Onboarding Steps;type=text;keywords=onboarding,steps,progress,guide;category=general;description=Serialized onboarding steps"`

	// Docker category
	AutoUpdate         SettingVariable `key:"autoUpdate" meta:"label=Auto Update;type=boolean;keywords=auto,update,automatic,upgrade,refresh,restart,deploy;category=docker;description=Automatically update containers when new images are available" catmeta:"id=docker;title=Docker;icon=database;url=/settings/docker;description=Configure Docker settings, polling, and auto-updates"`
	AutoUpdateInterval SettingVariable `key:"autoUpdateInterval" meta:"label=Auto Update Interval;type=number;keywords=auto,update,interval,frequency,schedule,automatic,timing;category=docker;description=Interval between automatic updates"`
	PollingEnabled     SettingVariable `key:"pollingEnabled" meta:"label=Enable Polling;type=boolean;keywords=polling,check,monitor,watch,scan,detection,automatic;category=docker;description=Enable automatic checking for image updates"`
	PollingInterval    SettingVariable `key:"pollingInterval" meta:"label=Polling Interval;type=number;keywords=interval,frequency,schedule,time,minutes,period,delay;category=docker;description=How often to check for image updates"`
	PruneMode          SettingVariable `key:"dockerPruneMode" meta:"label=Docker Prune Action;type=select;keywords=prune,cleanup,clean,remove,delete,unused,dangling,space,disk;category=docker;description=Configure how unused Docker images are cleaned up"`
	DockerHost         SettingVariable `key:"dockerHost,public,envOverride" meta:"label=Docker Host;type=text;keywords=docker,host,daemon,socket,unix,remote;category=docker;description=URI for Docker daemon"`

	// Security category
	AuthLocalEnabled   SettingVariable `key:"authLocalEnabled,public" meta:"label=Local Authentication;type=boolean;keywords=local,auth,authentication,username,password,login,credentials;category=security;description=Enable local username/password authentication" catmeta:"id=security;title=Security;icon=shield;url=/settings/security;description=Manage authentication and security settings"`
	AuthOidcEnabled    SettingVariable `key:"authOidcEnabled,public" meta:"label=OIDC Authentication;type=boolean;keywords=oidc,openid,connect,sso,oauth,external,provider,federation;category=security;description=Enable OpenID Connect (OIDC) authentication"`
	AuthSessionTimeout SettingVariable `key:"authSessionTimeout" meta:"label=Session Timeout;type=number;keywords=session,timeout,expire,duration,lifetime,minutes,logout;category=security;description=How long user sessions remain active"`
	AuthPasswordPolicy SettingVariable `key:"authPasswordPolicy" meta:"label=Password Policy;type=select;keywords=password,policy,strength,complexity,requirements,security,rules;category=security;description=Set password strength requirements"`
	AuthOidcConfig     SettingVariable `key:"authOidcConfig,sensitive" meta:"label=OIDC Config;type=text;keywords=oidc,config,client,id,issuer,secret,oauth;category=security;description=OIDC provider configuration"`

	// Navigation category
	MobileNavigationMode         SettingVariable `key:"mobileNavigationMode,public,local" meta:"label=Mobile Navigation Mode;type=select;keywords=mode,style,type,floating,docked,position,layout,design,appearance,bottom;category=navigation;description=Choose between floating or docked navigation on mobile" catmeta:"id=navigation;title=Navigation;icon=navigation;url=/settings/navigation;description=Customize navigation and interface behavior"`
	MobileNavigationShowLabels   SettingVariable `key:"mobileNavigationShowLabels,public,local" meta:"label=Show Navigation Labels;type=boolean;keywords=labels,text,icons,display,show,hide,names,captions,titles,visible,toggle;category=navigation;description=Display text labels alongside navigation icons"`
	MobileNavigationScrollToHide SettingVariable `key:"mobileNavigationScrollToHide,public,local" meta:"label=Scroll to Hide;type=boolean;keywords=scroll,hide,auto-hide,behavior,down,up,automatic,disappear,vanish,minimize,collapse;category=navigation;description=Automatically hide navigation when scrolling down"`
	SidebarHoverExpansion        SettingVariable `key:"sidebarHoverExpansion,public,local" meta:"label=Sidebar Hover Expansion;type=boolean;keywords=sidebar,hover,expansion,expand,desktop,mouse,over,collapsed,collapsible,icon,labels,text,preview,peek,tooltip,overlay,temporary,quick,access,navigation,menu,items,submenu,nested;category=navigation;description=Expand sidebar on hover in desktop mode"`
	GlassEffectEnabled           SettingVariable `key:"glassEffectEnabled,public,local" meta:"label=Glass Effect;type=boolean;keywords=glass,glassmorphism,blur,backdrop,frosted,effect,gradient,ambient,design,ui,appearance,modern,visual,style,theme,transparency,translucent;category=navigation;description=Enable modern glassmorphism design with blur, gradients, and ambient effects"`

	InstanceID SettingVariable `key:"instanceId,internal" meta:"label=Instance ID;type=text;keywords=instance,id,uuid,identifier;category=internal;description=Unique instance identifier"`
}

func (SettingVariable) TableName() string {
	return "settings"
}

func (s *Settings) ToSettingVariableSlice(showAll bool, redactSensitiveValues bool) []SettingVariable {
	cfgValue := reflect.ValueOf(s).Elem()
	cfgType := cfgValue.Type()

	var res []SettingVariable

	for i := 0; i < cfgType.NumField(); i++ {
		field := cfgType.Field(i)

		key, attrs, _ := strings.Cut(field.Tag.Get("key"), ",")
		if key == "" {
			continue
		}

		if !showAll && !strings.Contains(attrs, "public") {
			continue
		}

		value := cfgValue.Field(i).FieldByName("Value").String()
		value = redactSettingValue(key, value, attrs, redactSensitiveValues)

		settingVariable := SettingVariable{
			Key:   key,
			Value: value,
		}
		res = append(res, settingVariable)
	}

	return res
}

func (s *Settings) FieldByKey(key string) (defaultValue string, isPublic bool, isSensitive bool, err error) {
	rv := reflect.ValueOf(s).Elem()
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		tagValue := strings.Split(rt.Field(i).Tag.Get("key"), ",")
		keyFromTag := tagValue[0]
		isPublic = slices.Contains(tagValue, "public")
		isSensitive = slices.Contains(tagValue, "sensitive")

		if keyFromTag != key {
			continue
		}

		valueField := rv.Field(i).FieldByName("Value")
		return valueField.String(), isPublic, isSensitive, nil
	}

	return "", false, false, SettingKeyNotFoundError{field: key}
}

func (s *Settings) IsLocalSetting(key string) bool {
	rt := reflect.TypeOf(s).Elem()

	for i := 0; i < rt.NumField(); i++ {
		tagValue := strings.Split(rt.Field(i).Tag.Get("key"), ",")
		keyFromTag := tagValue[0]

		if keyFromTag == key {
			return slices.Contains(tagValue, "local")
		}
	}

	return false
}

func (s *Settings) UpdateField(key string, value string, noSensitive bool) error {
	rv := reflect.ValueOf(s).Elem()
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		tagValue, attrs, _ := strings.Cut(rt.Field(i).Tag.Get("key"), ",")
		if tagValue != key {
			continue
		}

		// If the field is sensitive and noSensitive is true, we skip that
		if noSensitive && strings.Contains(attrs, "sensitive") {
			return SettingSensitiveForbiddenError{field: key}
		}

		valueField := rv.Field(i).FieldByName("Value")
		if !valueField.CanSet() {
			return fmt.Errorf("field Value in SettingVariable is not settable for config key '%s'", key)
		}

		valueField.SetString(value)
		return nil
	}

	return SettingKeyNotFoundError{field: key}
}

// helper keeps redaction logic in one place; behavior unchanged
func redactSettingValue(key, value, attrs string, redact bool) string {
	if value == "" || !redact || !strings.Contains(attrs, "sensitive") {
		return value
	}

	if key == keyAuthOidcConfig {
		var cfg OidcConfig
		if err := json.Unmarshal([]byte(value), &cfg); err == nil {
			cfg.ClientSecret = ""
			if redacted, err := json.Marshal(cfg); err == nil {
				return string(redacted)
			}
			return redactionMask
		}
		return redactionMask
	}

	return redactionMask
}

type SettingKeyNotFoundError struct {
	field string
}

func (e SettingKeyNotFoundError) Error() string {
	return "cannot find setting key '" + e.field + "'"
}

func (e SettingKeyNotFoundError) Is(target error) bool {
	x := SettingKeyNotFoundError{}
	return errors.As(target, &x)
}

type SettingSensitiveForbiddenError struct {
	field string
}

func (e SettingSensitiveForbiddenError) Error() string {
	return "field '" + e.field + "' is sensitive and can't be updated"
}

func (e SettingSensitiveForbiddenError) Is(target error) bool {
	x := SettingSensitiveForbiddenError{}
	return errors.As(target, &x)
}

type OidcConfig struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	IssuerURL    string `json:"issuerUrl"`
	Scopes       string `json:"scopes"`

	AuthorizationEndpoint string `json:"authorizationEndpoint,omitempty"`
	TokenEndpoint         string `json:"tokenEndpoint,omitempty"`
	UserinfoEndpoint      string `json:"userinfoEndpoint,omitempty"`
	JwksURI               string `json:"jwksUri,omitempty"`

	// Admin mapping: evaluate this claim to grant admin.
	// Examples:
	// - adminClaim: "admin", adminValue: "true"        (boolean or string "true")
	// - adminClaim: "roles", adminValue: "admin"       (array membership)
	// - adminClaim: "realm_access.roles", adminValue: "admin" (Keycloak)
	AdminClaim string `json:"adminClaim,omitempty"`
	AdminValue string `json:"adminValue,omitempty"`
}
