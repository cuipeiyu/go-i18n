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
	"gopkg.in/yaml.v2"
)

// 遍历项目文件夹找出代码片段
func walk() {
	// fmt.Println("Print: " + strings.Join(args, " "))
	// log.Println("AllSettings", viper.AllSettings())

	logrus.Debugf("开始遍历文件夹")

	workspace := filepath.Dir(getGoEnv("GOMOD"))
	ignoreTestFiles := viper.GetBool("ignore-test-files")

	paths := viper.GetStringSlice("path")
	if paths == nil {
		paths = []string{
			workspace,
		}
	}

	messages := []*i18n.Message{}
	for _, path := range paths {
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

	// messageTemplates := map[string]*i18n.MessageTemplate{}
	// for _, m := range messages {
	// 	if mt := i18n.NewMessageTemplate(m); mt != nil {
	// 		messageTemplates[m.ID] = mt
	// 	}
	// }

	logrus.Debugf("找到 %d 条", len(messages))
	buf, err := yaml.Marshal(messages)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	// log.Println(viper.AllSettings())

	outdir := viper.GetString("outdir")
	filename := filepath.Join(workspace, outdir, viper.GetString("default")+".yaml")
	// log.Println("-----", outdir, filename)
	err = ioutil.WriteFile(filename, buf, os.ModePerm)
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
