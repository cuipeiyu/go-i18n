package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	i18n "github.com/cuipeiyu/go-i18n"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// 遍历项目文件夹找出代码片段
func walk() {
	logrus.Debugf("开始遍历文件夹")

	workspace := filepath.Dir(getGoEnv("GOMOD"))
	ignoreTestFiles := viper.GetBool("ignore-test-files")

	paths := viper.GetStringSlice("path")
	logrus.Debugf("所有路径 %t %#v", paths == nil, paths)
	if len(paths) == 0 {
		paths = []string{
			workspace,
		}
	}

	total := 0

	messages := []*i18n.Message{}
	for _, path := range paths {
		logrus.Debugf("搜索路径 %s", path)
		if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".go" {
				return nil
			}

			// Don't extract from test files.
			if ignoreTestFiles && strings.HasSuffix(path, "_test.go") {
				return nil
			}

			total++

			buf, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			msgs, err := extractMessages(buf)
			if err != nil {
				return err
			}
			messages = append(messages, msgs...)
			return nil
		}); err != nil {
			return
		}
	}

	if messages == nil {
		logrus.Infof("无匹配数据")
		return
	}

	logrus.Debugf("处理 %d 文件", total)
	logrus.Debugf("找到 %d 条", len(messages))

	messageMap := make(M, len(messages))
	for _, item := range messages {
		item.Hash = hash(*item)
		messageMap[item.ID] = *item
	}

	outdir := filepath.Join(workspace, viper.GetString("outdir"))
	filename := filepath.Join(outdir, viper.GetString("default")+".yaml")
	err := messageMap.write2File(filename, "original")
	if err != nil {
		logrus.Errorln(err)
		return
	}
}

// extractMessages extracts messages from the bytes of a Go source file.
func extractMessages(buf []byte) ([]*i18n.Message, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", buf, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	extractor := &extractor{i18nPackageName: i18nPackageName(file)}
	ast.Walk(extractor, file)
	return extractor.messages, nil
}
