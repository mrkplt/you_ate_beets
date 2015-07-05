package config

import (
  "io/ioutil"
  "gopkg.in/yaml.v2"
  "github.com/mrkplt/you_ate_beets/iffy"
)

type Config struct {
  Anaconda         struct {
    ConsumerKey      string
    ConsumerSecret   string
    AccessToken      string
    AccessSecret     string
  }
  Database         struct {
    Name             string
  }
}

func Secrets() Config {
  filename := "secrets.yaml"
  var config Config

  source, err := ioutil.ReadFile(filename)
  iffy.PanicIf(err)

  err = yaml.Unmarshal(source, &config)
  iffy.PanicIf(err)

  return config
}
