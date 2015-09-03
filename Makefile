
PKGDIR=./pkg
BASENAME=magellan-proxy
VERSION=`grep Version version.go | cut -f2 -d\"`
OS=linux
ARCH=amd64

PKGNAME=${BASENAME}-${VERSION}-${OS}_${ARCH}.zip
PKGFILE=${PKGDIR}/${PKGNAME}

all: build

SRCS=http.go magellan-proxy.go message.go trmq.go version.go

build: ${SRCS}
	GOOS=${OS} GOARCH=${ARCH} gom build github.com/groovenauts/magellan-proxy
	- mkdir -p ${PKGDIR}
	- rm -f ${PKGFILE}
	zip ${PKGFILE} ${BASENAME}
	rm -f ${BASENAME}

clean:
	rm -rf ${PKGDIR}
	rm -f ${BASENAME}
