APP_EXECUTABLE := mywebserver
PREFIX := /usr/local
BINDIR := $(PREFIX)/bin
INSTALL := install
INSTALL_PROGRAM := $(INSTALL) -m 755
INSTALL_DATA := $(INSTALL) -m 644

.PHONY = all build run clean
all: build
build:
	GOARCH=amd64 GOOS=linux go build -o $(APP_EXECUTABLE)-linux 
	GOARCH=amd64 GOOS=darwin go build -o $(APP_EXECUTABLE)-darwin 
	GOARCH=amd64 GOOS=windows go build -o $(APP_EXECUTABLE)-windows 
run: build
	./$(APP_EXECUTABLE)-linux
clean:
	rm -f $(APP_EXECUTABLE)-linux $(APP_EXECUTABLE)-darwin $(APP_EXECUTABLE)-windows $(APP_EXECUTABLE)
# install:
# 	$(APP_EXECUTABLE)
# 	@echo "Installing $(APP_EXECUTABLE) to $(BINDIR)"
# 	mkdir -p $(BINDIR)
# 	$(INSTALL_PROGRAM) $(APP_EXECUTABLE) $(BINDIR)
# uninstall:
# 	@echo "Removing $(BINDIR)/$(APP_EXECUTABLE)"
# 	rm -f $(BINDIR)/$(APP_EXECUTABLE)
