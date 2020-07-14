#!/usr/bin/env bash

build_one() {
  echo "[linux] building $1/$2"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-s' -o build/bin/linux/$2_d ./$1/$2/main.go
	echo "[windows] building $1/$2"
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '-s' -o build/bin/windows/$2_d.exe ./$1/$2/main.go
	echo "[darwin] building $1/$2"
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags '-s' -o build/bin/darwin/$2_d ./$1/$2/main.go
}

build() {
  if [ ! $2 == "all" ] ;then
    echo "build one"
    build_one $1 $2
  else
    echo "build all"
    for file in `ls $1`
    do
        if test -f $file
        then
            echo $file 是文件
        else
            build_one $1 $file
        fi
    done
  fi
}

if [ ! -n "$1" ] ;then
  echo  $"Usage: $0 {example|server} {dir_name}"
fi

case "$1" in
  examples)
    ;;
  server)
    ;;
  *)
    echo  $"Usage: $0 {examples|server} {dir_name}"
    exit 1
esac

build $1 $2