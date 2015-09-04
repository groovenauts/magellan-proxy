
PKGDIR=./pkg
BASENAME=magellan-proxy
VERSION=`grep Version version.go | cut -f2 -d\"`
OS=linux
ARCH=amd64

PKGNAME=${BASENAME}-${VERSION}_${OS}-${ARCH}.tar.gz
PKGFILE=${PKGDIR}/${PKGNAME}

all: build

SRCS=http.go magellan-proxy.go message.go trmq.go version.go

build: ${SRCS}
	GOOS=${OS} GOARCH=${ARCH} gom build github.com/groovenauts/magellan-proxy
	- mkdir -p ${PKGDIR}
	- rm -f ${PKGFILE}
	tar zcf ${PKGFILE} ${BASENAME}
	rm -f ${BASENAME}

release: build
	ghr -u groovenauts --replace --draft ${VERSION} pkg

prerelease: build
	ghr -u groovenauts --replace --draft --prerelease ${VERSION} pkg


clean:
	rm -rf ${PKGDIR}
	rm -f ${BASENAME}
