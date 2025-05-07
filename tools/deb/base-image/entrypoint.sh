#!/usr/bin/env bash
set -x
set -euo pipefail

PATH=/usr/local/go/bin:$PATH

trap term INT TERM

export GOFLAGS=-mod=vendor

if [[ -n "${USER_NAME:-}" && -n "${USER_UID:-}" && -n "${USER_GID:-}" ]]; then
	echo "Creating user ${USER_NAME} with UID=${USER_UID}, GID=${USER_GID}..."

	if ! getent group "$USER_GID" >/dev/null; then
		groupadd -g "$USER_GID" "$USER_NAME"
	fi

	if ! id -u "$USER_NAME" >/dev/null 2>&1; then
		useradd -m -u "$USER_UID" -g "$USER_GID" -s /bin/bash "$USER_NAME"
	fi

	chown -R "$USER_UID":"$USER_GID" /home/$USER_NAME

	su - "$USER_NAME" -c "echo 'export GOPATH=/home/$USER_NAME/go' >> ~/.bashrc"
	su - "$USER_NAME" -c "echo 'export PATH=\$GOPATH/bin:/usr/local/go/bin:\$PATH' >> ~/.bashrc"
	su - "$USER_NAME" -c "echo 'export PATH=/usr/local/go/bin:\$PATH' >> ~/.bashrc"
	exec su - "$USER_NAME" -c "$@"
else
	echo "Running as default user (root)..."
        exec "$@"
fi
