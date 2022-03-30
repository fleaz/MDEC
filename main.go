package main

import (
	"net/http"
	"text/template"
    "strings"
    "errors"

    log "github.com/sirupsen/logrus"
    "github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
    "github.com/imdario/mergo"
)

func splitEmailAddress(email string) (string,string,error) {
    idx := strings.LastIndex(email, "@")
    if idx >= 0 {
        return email[:idx], email[idx+1:], nil
    } else {
        return "", "", errors.New("invalid email address")
    }
}

func getDomainConfig(domain string) DomainConfig {
    conf := configList["default"]
    for k, v := range configList{
        if k == domain {
            err := mergo.Merge(&conf, v, mergo.WithOverride)
            if  err != nil {
                log.Fatalf("failed to merge config: %s\n", err)
            }
            break
        }
    }

    // Set correct domain name in config
    conf.Domain = domain
    return conf
}

func AutoConfig(w http.ResponseWriter, r *http.Request) {
	log.Info("autoconfig request")

    query := r.URL.Query()

    // Get channel to send to
    emailaddress := query.Get("emailaddress")
    if emailaddress == "" {
        // return 400 
    }
    _,domainPart, err := splitEmailAddress(emailaddress)
    if err != nil{
        log.Error(err)
        return
    }

    domainConf := getDomainConfig(domainPart)

    t,err := template.ParseFiles("autoconfig.xml")
    if err != nil{
        log.Error(err)
        return
    }
    err = t.Execute(w, domainConf)
    if err != nil{
        log.Error(err)
        return
    }
}

func AutoDiscover(w http.ResponseWriter, r *http.Request) {
	log.Info("autodiscover request")
	http.ServeFile(w, r, "autodiscover.xml")
}

func MobileConfig(w http.ResponseWriter, r *http.Request) {
	log.Info("mobileconfig request")
	http.ServeFile(w, r, "mobileconfig.xml")
}

type MailServer struct {
    Type string `yaml:"type"`
    Host string `yaml:"host"`
    Port int `yaml:"port"`
    SSLMethod string `yaml:"ssl_method"`
    Username string `yaml:"username"`
    PasswordType string `yaml:"password_type"`
}

type DomainConfig struct {
    Domain string
    Incoming MailServer `yaml:"incoming"`
    Outgoing MailServer `yaml:"outgoing"`
}
var configList = make(map[string]DomainConfig)
var k = koanf.New(".")


func main() {
    k.Load(file.Provider("config.yaml"), yaml.Parser())

    k.Unmarshal("domains", &configList)

	// Thunderbird
	http.HandleFunc("/mail/config-v1.1.xml", AutoConfig)
	// Outlook
	http.HandleFunc("/autodiscover/autodiscover.xml", AutoDiscover)
	// iOS
	http.HandleFunc("/mobileconfig", MobileConfig)

    listenAddr := k.String("listen")
    log.Infof("Listing on %q\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
