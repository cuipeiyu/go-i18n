package main

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/cuipeiyu/go-i18n"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func merge() {
	defaultLang := viper.GetString("default")
	if defaultLang == "" {
		logrus.Errorf("default language is empty")
		return
	}

	targetLang := viper.GetStringSlice("target")
	if targetLang == nil {
		logrus.Errorf("target languages is empty")
		return
	}

	for _, lang := range targetLang {
		mergeLang(defaultLang, lang)
	}
}

func mergeLang(source, target string) {
	outdir := viper.GetString("outdir")
	workspace := filepath.Dir(getGoEnv("GOMOD"))

	sourceMap := make(map[string]*i18n.Message)
	targetMap := make(map[string]*i18n.Message)
	middleMap := make(map[string]*i18n.Message)

	{
		filename := filepath.Join(workspace, outdir, source+".yaml")
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			logrus.Fatalf("读取文件失败: %v", err)
		}

		var list []*i18n.Message
		err = yaml.Unmarshal(buf, &list)
		if err != nil {
			logrus.Fatalf("无法识别的文件: %v", err)
		}

		for _, item := range list {
			sourceMap[item.ID] = item
		}
	}
	{
		filename := filepath.Join(workspace, outdir, target+".yaml")
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			logrus.Fatalf("读取文件失败: %v", err)
		}

		var list []*i18n.Message
		err = yaml.Unmarshal(buf, &list)
		if err != nil {
			logrus.Fatalf("无法识别的文件: %v", err)
		}

		for _, item := range list {
			targetMap[item.ID] = item
		}
	}
	{
		filename := filepath.Join(workspace, outdir, target+".todo.yaml")
		buf, err := ioutil.ReadFile(filename)
		if err == nil {
			var list []*i18n.Message
			err = yaml.Unmarshal(buf, &list)
			if err != nil {
				logrus.Fatalf("无法识别的文件: %v", err)
			}

			for _, item := range list {
				middleMap[item.ID] = item
			}
		}
	}

	// 比对文件
	log.Println("比对文件")
}
