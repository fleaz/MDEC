package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/imdario/mergo"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	log "github.com/sirupsen/logrus"
)

func splitEmailAddress(email string) (string, string, error) {
	idx := strings.LastIndex(email, "@")
	if idx >= 0 {
		return email[:idx], email[idx+1:], nil
	} else {
		return "", "", errors.New("invalid email address")
	}
}

func getDomainConfig(emailAddress string) DomainConfig {
	conf := configList["default"]

	_, domainPart, err := splitEmailAddress(emailAddress)
	if err != nil {
		log.Error(err)
		return conf
	}

	for k, v := range configList {
		if k == domainPart {
			err := mergo.Merge(&conf, v, mergo.WithOverride)
			if err != nil {
				log.Fatalf("failed to merge config: %s\n", err)
			}
			break
		}
	}

	// Set correct domain name in config
	conf.Domain = domainPart
	conf.EMail = emailAddress
	return conf
}

func AutoConfig(w http.ResponseWriter, r *http.Request) {
	log.Info("autoconfig request")

	query := r.URL.Query()
	emailaddress := query.Get("emailaddress")
	if emailaddress == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	data, _ := renderConfig(emailaddress, "autoconfig.xml")
	w.Write(data)
}

func AutoDiscover(w http.ResponseWriter, r *http.Request) {
	log.Info("autodiscover request")
	http.ServeFile(w, r, "autodiscover.xml")
}

func renderConfig(emailAddress, templateName string) ([]byte, error) {
	domainConf := getDomainConfig(emailAddress)

	t, err := template.ParseFiles(templateName)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var data bytes.Buffer
	err = t.Execute(&data, domainConf)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return data.Bytes(), nil
}

func MobileConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "mobileconfig.html")
	} else if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		emailaddress := r.FormValue("email")
		if emailaddress == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		data, err := renderConfig(emailaddress, "mobileconfig.xml")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=mobileconfig_%s.xml", emailaddress))
		w.Write(data)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

type MailServer struct {
	Type         string `yaml:"type"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	SSLMethod    string `yaml:"ssl_method"`
	Username     string `yaml:"username"`
	PasswordType string `yaml:"password_type"`
}

type GeneralConfig struct {
	Organization string `yaml:"organization"`
}

type DomainConfig struct {
	Domain   string
	EMail    string
	Incoming MailServer    `yaml:"incoming"`
	Outgoing MailServer    `yaml:"outgoing"`
	General  GeneralConfig `yaml:"general"`
}

var configList = make(map[string]DomainConfig)
var k = koanf.New(".")

func main() {
	k.Load(file.Provider("config.yaml"), yaml.Parser())

	k.Unmarshal("domains", &configList)

	// Thunderbird
	http.HandleFunc("/mail/config-v1.1.xml", AutoConfig)
	// Apple Mail
	http.HandleFunc("/autodiscover/autodiscover.xml", AutoDiscover)
	// iOS
	http.HandleFunc("/mobileconfig", MobileConfig)

	listenAddr := k.String("listen")
	log.Infof("Listing on %q\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
