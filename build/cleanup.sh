#!/bin/bash

# Runs as the deb prerm (remove|upgrade|deconfigure|failed-upgrade) and from the
# RPM %preun (number of package instances left after the transaction). Only a
# real removal may touch system state.
case "${1:-}" in
    remove|0) ;;
    *) exit 0 ;;
esac

# RPM has no purge step; the deb postrm drops tracker state on purge instead.
if [ "$1" = "0" ]; then
    rm -f /var/log/gpu-tracker.json /var/log/gpu-tracker.lock
fi

# A failed step must not block uninstall.
if ! /usr/bin/amd-ctk cdi remove; then
    echo "Could not remove /etc/cdi/amd.json; remove it manually." >&2
fi
if ! /usr/bin/amd-ctk runtime configure --remove; then
    echo "Could not remove the amd runtime from /etc/docker/daemon.json; remove it manually." >&2
fi
exit 0
