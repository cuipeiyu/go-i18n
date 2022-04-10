
.PHONY: remove
remove:
	rm -rf locales

.PHONY: trans
trans:
	goi18n walk
	goi18n merge -t zh-CN

.PHONY: run
run:
	go run main.go
