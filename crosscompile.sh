# git submodule init
# git submodule update

APP=ckptool

echo "About to cross-compile $APP"
if [ "$1" = "" ]; then
        echo "You need to pass in the version number as the first parameter like ./crosscompile 1.00"
        exit
fi

rm -rf snapshot/*

echo "Building Linux amd64"
mkdir snapshot/$APP-$1_linux_amd64
env GOOS=linux GOARCH=amd64 go build -v -o snapshot/$APP-$1_linux_amd64/$APP
exit

echo "" 
echo "Building Linux 386"
mkdir snapshot/$APP-$1_linux_386
env GOOS=linux GOARCH=386 go build -v -o snapshot/$APP-$1_linux_386/$APP

echo "" 
echo "Building Windows x32"
mkdir snapshot/$APP-$1_windows_386
env GOOS=windows GOARCH=386 go build -v -o snapshot/$APP-$1_windows_386/$APP.exe

echo "" 
echo "Building Windows x64"
mkdir snapshot/$APP-$1_windows_amd64
env GOOS=windows GOARCH=amd64 go build -v -o snapshot/$APP-$1_windows_amd64/$APP.exe

echo "" 
echo "Building Darwin x64"
mkdir snapshot/$APP-$1_darwin_amd64
env GOOS=darwin GOARCH=amd64 go build -v -o snapshot/$APP-$1_darwin_amd64/$APP

