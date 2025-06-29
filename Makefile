GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

SDPLUGINDIR=./com.exension.hwinfo.sdPlugin
BUILD_DIR=./build

PROTOS=$(wildcard ./*/**/**/*.proto)
PROTOPB=$(PROTOS:.proto=.pb.go)

.PHONY: build clean proto plugin debug release kill-processes

build: plugin

kill-processes:
	-@taskkill /F /IM StreamDeck.exe /T >nul 2>&1
	-@taskkill /F /IM hwinfo.exe /T >nul 2>&1
	-@taskkill /F /IM hwinfo-plugin.exe /T >nul 2>&1
	-@timeout /t 2 /nobreak >nul

plugin: proto kill-processes
	$(GOBUILD) -o $(SDPLUGINDIR)/hwinfo.exe ./cmd/hwinfo_streamdeck_plugin
	$(GOBUILD) -o $(SDPLUGINDIR)/hwinfo-plugin.exe ./cmd/hwinfo-plugin
	-@install-plugin.bat

proto: $(PROTOPB)

$(PROTOPB): $(PROTOS)
	@mkdir -p .cache/protoc/bin
	.cache/protoc/bin/protoc \
		--go_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:. \
		--go-grpc_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		$(<)

# plugin:
# 	-@kill-streamdeck.bat
# 	@go build -o com.exension.hwinfo.sdPlugin\\hwinfo.exe github.com/shayne/hwinfo-streamdeck/cmd/hwinfo_streamdeck_plugin
# 	@xcopy com.exension.hwinfo.sdPlugin $(APPDATA)\\Elgato\\StreamDeck\\Plugins\\com.exension.hwinfo.sdPlugin\\ /E /Q /Y
# 	@start-streamdeck.bat

debug: kill-processes
	$(GOBUILD) -o $(SDPLUGINDIR)/hwinfo.exe ./cmd/hwinfo_debugger
	$(GOBUILD) -o $(SDPLUGINDIR)/hwinfo-plugin.exe ./cmd/hwinfo-plugin
	-@install-plugin.bat
# @xcopy com.exension.hwinfo.sdPlugin $(APPDATA)\\Elgato\\StreamDeck\\Plugins\\com.exension.hwinfo.sdPlugin\\ /E /Q /Y

release: build
	@mkdir -p $(BUILD_DIR)
	-@rm -f $(BUILD_DIR)/com.exension.hwinfo.streamDeckPlugin
	@DistributionTool.exe -b -i $(SDPLUGINDIR) -o $(BUILD_DIR)

clean: kill-processes
	$(GOCLEAN)
	-@rm -f $(SDPLUGINDIR)/hwinfo.exe
	-@rm -f $(SDPLUGINDIR)/hwinfo-plugin.exe
	-@rm -f $(BUILD_DIR)/com.exension.hwinfo.streamDeckPlugin
