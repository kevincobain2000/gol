#!/bin/sh

BIN_DIR=$(pwd)

THE_ARCH_BIN=''
THIS_PROJECT_NAME='gol'

THISOS=$(uname -s)
ARCH=$(uname -m)

case $THISOS in
   Linux*)
      case $ARCH in
        arm64)
          THE_ARCH_BIN="$THIS_PROJECT_NAME-linux-arm64"
          ;;
        aarch64)
          THE_ARCH_BIN="$THIS_PROJECT_NAME-linux-arm64"
          ;;
        armv6l)
          THE_ARCH_BIN="$THIS_PROJECT_NAME-linux-arm"
          ;;
        armv7l)
          THE_ARCH_BIN="$THIS_PROJECT_NAME-linux-arm"
          ;;
        *)
          THE_ARCH_BIN="$THIS_PROJECT_NAME-linux-amd64"
          ;;
      esac
      ;;
   Darwin*)
      case $ARCH in
        arm64)
          THE_ARCH_BIN="$THIS_PROJECT_NAME-darwin-arm64"
          ;;
        *)
          THE_ARCH_BIN="$THIS_PROJECT_NAME-darwin-amd64"
          ;;
      esac
      ;;
   Windows|MINGW64_NT*)
      THE_ARCH_BIN="$THIS_PROJECT_NAME-windows-amd64.exe"
      THIS_PROJECT_NAME="$THIS_PROJECT_NAME.exe"
      ;;
esac

if [ -z "$THE_ARCH_BIN" ]; then
   echo "This script is not supported on $THISOS and $ARCH"
   exit 1
fi

curl -kL --progress-bar https://github.com/kevincobain2000/$THIS_PROJECT_NAME/releases/latest/download/$THE_ARCH_BIN -o "$BIN_DIR"/$THIS_PROJECT_NAME

chmod +x "$BIN_DIR"/$THIS_PROJECT_NAME

echo "Installed successfully to: $BIN_DIR/$THIS_PROJECT_NAME"

