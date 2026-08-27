#!/bin/sh
# setup.sh - Universal environment setup script for Bee Programming Language
# Adds bin/ to user PATH across POSIX and Windows (sh-compatible)

echo "Setting up Bee Programming Language environment..."

BIN_DIR="$(pwd)/bin"

case ":$PATH:" in
    *":$BIN_DIR:"*)
        echo "bin/ is already in PATH."
        ;;
    *)
        echo "Adding $BIN_DIR to PATH..."
        if [ -n "$ZSH_VERSION" ]; then
            echo "export PATH=\"$BIN_DIR:\$PATH\"" >> ~/.zshrc
            echo "Added to ~/.zshrc. Please run 'source ~/.zshrc'."
        elif [ -n "$BASH_VERSION" ]; then
            echo "export PATH=\"$BIN_DIR:\$PATH\"" >> ~/.bashrc
            echo "Added to ~/.bashrc. Please run 'source ~/.bashrc'."
        else
            echo "Please add $BIN_DIR to your shell PATH manually."
        fi
        ;;
esac

echo "Setup completed successfully."
