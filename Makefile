# Build variables
BUILDDIR?=target
LDFLAGS?=
BUILDFLAGS?=-v -a -installsuffix cgo
PROJECT_NAME=main

OBJS := $(addprefix $(BUILDDIR)/,$(PROJECT_NAME))


$(OBJS) : main.go install_deps | $(BUILDDIR)
	go build ${LDFLAGS} ${BUILDFLAGS} $<
	mv ${PROJECT_NAME} $(BUILDDIR)/

%.go:
	echo $@ $^ $<
	golint $@ $<

$(BUILDDIR):
	-@mkdir $(BUILDDIR) 2> /dev/null

.PHONY: clean install_deps test build lint
build: $(OBJS)

lint: **/*.go
	golint $<

clean: install_deps
	go clean
	-rm -f ${BUILDDIR}/${PROJECT_NAME}
	-rm -rf ${BUILDDIR}

test: install_deps
	@go test -v
	@go test -v ./*/*_test.go

install_deps :
	dep ensure