#!/bin/sh

if [ -z "$BIN_DIR" ]; then
  BIN_DIR=/usr/local/bin
fi

echo "Installing gol to $BIN_DIR"

THE_ARCH_BIN=''
THIS_PROJECT_NAME='gol'

THISOS=$(uname -s)
ARCH=$(uname -m)
DEST=$BIN_DIR/$THIS_PROJECT_NAME

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


curl -kL --progress-bar https://github.com/kevincobain2000/$THIS_PROJECT_NAME/releases/latest/download/$THE_ARCH_BIN -o $THIS_PROJECT_NAME
echo "Downloaded $THIS_PROJECT_NAME"
chmod +x $THIS_PROJECT_NAME

SUDO=""

# check if $DEST is writable and suppress an error message
touch "$DEST" 2>/dev/null

# we need sudo powers to write to DEST
if [ $? -eq 1 ]; then
    echo "You do not have permission to write to $DEST, enter your password to grant sudo powers"
    SUDO="sudo"
fi

$SUDO mv $THIS_PROJECT_NAME "$DEST"

echo "Installed successfully to: $DEST"

