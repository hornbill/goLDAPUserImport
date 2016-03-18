#!/bin/bash
#
# usage: ./golang-crosscompile-build.bash

target=ldap_user_import.go
# Get Version out of target then replace . with _
version=$(go run $target -version)
version=${version//./_}
version=${version// /}
#platforms="darwin/386 darwin/amd64 freebsd/386 freebsd/amd64 freebsd/arm linux/386 linux/amd64 linux/arm windows/386 windows/amd64"
platforms="windows/386 windows/amd64"
echo "Building Version: $version"

for platform in ${platforms}
do
    split=(${platform//\// })
    goos=${split[0]}
    os=${split[0]}
    goarch=${split[1]}
    arch=${split[1]}
    output=ldap_user_import
    package=ldap_user_import
    # add exe to windows output
    [[ "windows" == "$goos" ]] && output="$output.exe"
    [[ "windows" == "$goos" ]] && os="win"
    [[ "386" == "$goarch" ]] && arch="x86"
    [[ "amd64" == "$goarch" ]] && arch="x64"

    destination="builds/$goos/$goarch/$output"

    echo "GOOS=$goos GOARCH=$goarch go build -o $destination $target"
    GOOS=$goos GOARCH=$goarch go build  -o $destination $target

    cp LICENSE.md "builds/$goos/$goarch/LICENSE.md"
    cp README.md "builds/$goos/$goarch/README.md"
    cp conf.json "builds/$goos/$goarch/conf.json"
    cd "builds/$goos/$goarch/"
    echo "zip -j "${package}_${os}_${arch}_v${version}.zip" $output LICENSE.md README.md conf.json"
    zip -r "${package}_${os}_${arch}_v${version}.zip" $output LICENSE.md README.md conf.json
    cp "${package}_${os}_${arch}_v${version}.zip" "../../../${package}_${os}_${arch}_v${version}.zip"
    cd "../../../"
done
rm -rf "builds/"
