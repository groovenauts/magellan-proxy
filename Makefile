
PKGDIR=./pkg
BASENAME=magellan-proxy
VERSION=`grep Version version.go | cut -f2 -d\"`
OS=linux
ARCH=amd64

PKGNAME=${BASENAME}-${VERSION}_${OS}-${ARCH}
PKGFILE=${PKGDIR}/${PKGNAME}

all: build

SRCS=http.go magellan-proxy.go message.go trmq.go version.go

build: ${SRCS}
	GOOS=${OS} GOARCH=${ARCH} gom build github.com/groovenauts/magellan-proxy
	- rm -rf ${PKGDIR}
	- mkdir -p ${PKGDIR}
	mv ${BASENAME} ${PKGFILE}

release: build
	ghr -u groovenauts --replace --draft ${VERSION} pkg

prerelease: build
	ghr -u groovenauts --replace --draft --prerelease ${VERSION} pkg

clean:
	rm -rf ${PKGDIR}
	rm -f ${BASENAME}
