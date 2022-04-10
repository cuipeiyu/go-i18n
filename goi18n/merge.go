package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
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

func mergeLang(original, target string) {
	// logrus.Debugf("%s => %s", source, target)
	outdir := viper.GetString("outdir")
	workspace := filepath.Dir(getGoEnv("GOMOD"))

	originalMap := make(M)
	middleMap := make(M)
	targetMap := make(M)

	{
		filename := filepath.Join(workspace, outdir, original+".yaml")
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			logrus.Errorf("来源文件不存在")
			return
		}
		// logrus.Debugf("读取文件 %s 内容", filename)
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			logrus.Fatalf("读取文件失败: %v", err)
		}

		// var list []i18n.Message
		// logrus.Debugf("序列化文件 %s 内容", filename)
		err = yaml.Unmarshal(buf, &originalMap)
		if err != nil {
			logrus.Fatalf("无法识别的文件: %v", err)
		}

		if len(originalMap) == 0 {
			logrus.Info("无数据，跳过")
			return
		}
	}
	{
		filename := filepath.Join(workspace, outdir, target+".todo.yaml")
		// logrus.Debugf("中间文件 %s", filename)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// logrus.Debugf("中间文件不存在")
		} else {
			// logrus.Debugf("读取文件 %s 内容", filename)
			buf, err := ioutil.ReadFile(filename)
			if err == nil {
				// logrus.Debugf("序列化文件 %s 内容", filename)
				err = yaml.Unmarshal(buf, &middleMap)
				if err != nil {
					logrus.Fatalf("无法识别的文件: %v", err)
				}
			}
		}
	}
	{
		filename := filepath.Join(workspace, outdir, target+".yaml")
		// logrus.Debugf("目标文件 %s", filename)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// logrus.Debugf("目标文件不存在")
		} else {
			// logrus.Debugf("读取文件 %s 内容", filename)
			buf, err := ioutil.ReadFile(filename)
			if err != nil {
				logrus.Fatalf("读取文件失败: %v", err)
			}

			// logrus.Debugf("序列化文件 %s 内容", filename)
			err = yaml.Unmarshal(buf, &targetMap)
			if err != nil {
				logrus.Fatalf("无法识别的文件: %v", err)
			}
		}
	}

	// 比对文件
	// logrus.Debugf("比对文件")

	originalMapResult, middleMapResult, targetMapResult := diff(originalMap, middleMap, targetMap)

	// 写入
	{
		if len(targetMapResult) > 0 {
			filename := filepath.Join(workspace, outdir, target+".yaml")
			targetMapResult.write2File(filename, "target")
		}
	}
	{
		filename := filepath.Join(workspace, outdir, target+".todo.yaml")
		if len(middleMapResult) > 0 {
			middleMapResult.write2File(filename, "middle")
		} else {
			_ = os.Remove(filename)
		}
	}
	// 确保其他文件保存后，再处理 original 文件
	{
		if len(originalMapResult) > 0 {
			filename := filepath.Join(workspace, outdir, original+".yaml")
			originalMapResult.write2File(filename, "original")
		}
	}
}

func diff(originalMap, middleMap, targetMap M) (M, M, M) {
	originalMapResult,
		middleMapResult,
		targetMapResult :=
		make(M),
		make(M),
		make(M)

	// 不同情景
	// 情景 1
	// 来源     中间     目标
	// ABC      -       -
	// 需要在【中间】新增
	//
	// 情景 2
	// 来源     中间     目标
	// ABC      ABC     -
	// 判断【来源】和【中间】是否相同，相同不处理，不相同移动【中间】至【目标】
	//
	// 情景 3
	// 来源     中间     目标
	// ABC      -       ABC
	// 判断【来源】和【目标】是否相同，相同则添加【中间】
	//
	// 情景 4
	// 来源     中间     目标
	// ABC      ABC     ABC
	// 判断【中间】和【目标】是否相同，相同则移除【中间】
	//
	// 情景 5
	// 来源     中间     目标
	// -        ABC     -
	// 移除【来源】和【目标】
	//
	// 情景 6
	// 来源     中间     目标
	// -        -       ABC
	// 移除【来源】和【中间】

	for id, org := range originalMap {
		mid, midHas := middleMap[id]
		tar, tarHas := targetMap[id]

		originalMapResult[id] = org

		// 情景 1
		if !midHas && !tarHas {
			// logrus.Debugf("情景1")
			middleMapResult[id] = org
			targetMapResult[id] = org
		}
		// 情景 2
		if midHas && !tarHas {
			// logrus.Debugf("情景2")
			h1 := hash(org)
			h2 := hash(mid)
			if h1 == h2 {
				// org.Hash = h1
				middleMapResult[id] = org // 保留
				targetMapResult[id] = org // 复制
			} else {
				mid.Hash = h1
				targetMapResult[id] = mid // 移动
			}
		}
		// 情景 3
		if !midHas && tarHas {
			// logrus.Debugf("情景3")

			if org.Hash == tar.Hash { // 若相同
				if org.Hash == hash(tar) { // 内容一至，表示没翻译
					middleMapResult[id] = org // 需要翻译
					targetMapResult[id] = tar // 不变
				} else {
					// 视为内容是被正确翻译的
				}
			} else { // 内容不同
				tar.Hash = org.Hash
				middleMapResult[id] = org // 添加需要翻译项
				targetMapResult[id] = tar // 不变
			}
		}
		// 情景 4
		if midHas && tarHas {
			// logrus.Debugf("情景4")
			h1 := hash(mid)
			h2 := hash(tar)
			if org.Hash == tar.Hash { // 正常的翻译
				mid.Hash = org.Hash
				if h1 == h2 { // 需要翻译
					middleMapResult[id] = mid // 不变
				}
				targetMapResult[id] = mid // 不变
			} else {
				if h1 == h2 { // 需要翻译
					tar.Hash = org.Hash
					// middleMapResult[id] = mid // 不变
					targetMapResult[id] = tar // 不变
				} else {
					mid.Hash = org.Hash
					targetMapResult[id] = mid // 移动
				}
			}
		}
	}

	return originalMapResult, middleMapResult, targetMapResult
}

func hash(t i18n.Message) string {
	h := sha1.New()
	// _, _ = io.WriteString(h, t.Description)
	_, _ = io.WriteString(h, t.Zero)
	_, _ = io.WriteString(h, t.One)
	_, _ = io.WriteString(h, t.Two)
	_, _ = io.WriteString(h, t.Few)
	_, _ = io.WriteString(h, t.Many)
	_, _ = io.WriteString(h, t.Other)
	return hex.EncodeToString(h.Sum(nil))
}
