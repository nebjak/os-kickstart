#!/bin/bash
set -euo pipefail

# Kernel & network optimization: sysctl, limits, sshd, scheduler, autotune
# Author: Dusan Panic <dpanic@gmail.com>
# Source: https://github.com/dpanic/patchfiles
# Safe to re-run -- idempotent
#
# Usage:
#   ./kernel-optimize.sh                    # apply all optimizations
#   ./kernel-optimize.sh sysctl limits      # apply only sysctl and limits
#
# Requires: sudo (all files are system-level)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

source "$SCRIPT_DIR/lib.sh"

ALL_COMPONENTS=(sysctl limits sshd scheduler autotune)
parse_update_flag "$@"
COMPONENTS=("${_CLEAN_ARGS[@]}")
if [[ ${#COMPONENTS[@]} -eq 0 ]]; then
    COMPONENTS=("${ALL_COMPONENTS[@]}")
fi

want() {
    local c
    for c in "${COMPONENTS[@]}"; do [[ "$c" == "$1" ]] && return 0; done
    return 1
}

STEP=0
count_steps() {
    local total=0
    for c in "${ALL_COMPONENTS[@]}"; do want "$c" && total=$((total + 1)); done
    echo "$total"
}
TOTAL=$(count_steps)
next() { STEP=$((STEP + 1)); echo "[$STEP/$TOTAL] $1..."; }

backup_file() {
    local target="$1"
    if [[ -f "$target" ]]; then
        sudo cp "$target" "${target}.bak-kickstart"
        echo "  backup: ${target}.bak-kickstart"
    fi
}

append_if_missing() {
    local target="$1"
    local marker="$2"
    local content="$3"
    if grep -qF "$marker" "$target" 2>/dev/null; then
        skip "already present in $target"
    else
        backup_file "$target"
        echo "$content" | sudo tee -a "$target" >/dev/null
        echo "  appended to $target"
    fi
}

echo "=== Kernel & Network Optimization ==="
echo "  Components: ${COMPONENTS[*]}"
echo ""

# ── sysctl.conf ───────────────────────────────────────────────────────────────
if want "sysctl"; then
    next "sysctl.conf"

    backup_file /etc/sysctl.conf
    sudo cp "$REPO_DIR/configs/sysctl.conf" /etc/sysctl.conf
    sudo sysctl -p >/dev/null 2>&1 || echo "  warning: some sysctl params may require autotune/reboot"
    echo "  done: /etc/sysctl.conf (from configs/sysctl.conf)"
fi

# ── limits ────────────────────────────────────────────────────────────────────
if want "limits"; then
    next "file descriptor & process limits"

    # limits.conf -- overwrite
    backup_file /etc/security/limits.conf
    sudo tee /etc/security/limits.conf >/dev/null << 'EOF'
# KICKSTART -- file descriptor and process limits from dpanic/patchfiles
* hard nofile 2097152
* soft nofile 2097152
root hard nofile 2097152
root soft nofile 2097152

* soft nproc 65535
* hard nproc 65535
root soft nproc 65535
root hard nproc 65535

* hard stack 131072
* soft stack 131072
EOF
    echo "  done: /etc/security/limits.conf"

    # PAM session modules -- append if missing
    append_if_missing /etc/pam.d/common-session \
        "pam_limits.so" \
        "# KICKSTART -- enable pam_limits for desktop sessions
session required pam_limits.so"

    append_if_missing /etc/pam.d/common-session-noninteractive \
        "pam_limits.so" \
        "# KICKSTART -- enable pam_limits for SSH sessions
session required pam_limits.so"

    # systemd DefaultLimitNOFILE -- append if missing
    append_if_missing /etc/systemd/system.conf \
        "DefaultLimitNOFILE=2097152" \
        "# KICKSTART -- increase systemd file descriptor limit
DefaultLimitNOFILE=2097152"

    append_if_missing /etc/systemd/user.conf \
        "DefaultLimitNOFILE=2097152" \
        "# KICKSTART -- increase systemd user file descriptor limit
DefaultLimitNOFILE=2097152"

    echo "  done: limits + PAM + systemd"
fi

# ── sshd ──────────────────────────────────────────────────────────────────────
if want "sshd"; then
    next "sshd hardening"

    if [[ ! -f /etc/ssh/sshd_config ]]; then
        skip "sshd not installed"
    else
        backup_file /etc/ssh/sshd_config
        sudo tee /etc/ssh/sshd_config >/dev/null << 'EOF'
# KICKSTART -- hardened sshd config from dpanic/patchfiles
Port 22
AddressFamily any
ListenAddress 0.0.0.0

PubkeyAuthentication yes
PasswordAuthentication no
ChallengeResponseAuthentication no
GSSAPIAuthentication no
UsePAM yes

AllowAgentForwarding yes
AllowTcpForwarding yes
X11Forwarding yes

PrintMotd no
PrintLastLog yes
TCPKeepAlive yes
Compression delayed
UseDNS no

AcceptEnv LANG LC_*
Subsystem sftp /usr/lib/openssh/sftp-server

ClientAliveInterval 120
ClientAliveCountMax 40

Ciphers aes128-ctr,aes192-ctr,aes256-ctr,chacha20-poly1305@openssh.com,aes256-gcm@openssh.com,aes128-gcm@openssh.com

HostKeyAlgorithms ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,ssh-rsa,ssh-dss

KexAlgorithms curve25519-sha256@libssh.org,ecdh-sha2-nistp256,ecdh-sha2-nistp384,ecdh-sha2-nistp521,diffie-hellman-group-exchange-sha256

MACs hmac-sha2-256,hmac-sha2-512,hmac-sha2-512-etm@openssh.com,hmac-sha2-256-etm@openssh.com,umac-128-etm@openssh.com,umac-128@openssh.com
EOF

        sudo systemctl restart sshd 2>/dev/null || sudo systemctl restart ssh 2>/dev/null || true
        echo "  done: /etc/ssh/sshd_config (password auth DISABLED)"
    fi
fi

# ── scheduler ─────────────────────────────────────────────────────────────────
if want "scheduler"; then
    next "I/O scheduler (none -- best for SSD/NVMe)"

    sudo tee /etc/udev/rules.d/60-scheduler.rules >/dev/null << 'EOF'
# KICKSTART -- disable I/O scheduler (optimal for SSD/NVMe) from dpanic/patchfiles
ACTION=="add|change", KERNEL=="sd*[!0-9]|sr*|nvme*|mmcblk*", ATTR{queue/scheduler}="none"
EOF

    sudo udevadm control --reload 2>/dev/null || true
    sudo udevadm trigger 2>/dev/null || true
    echo "  done: /etc/udev/rules.d/60-scheduler.rules"
fi

# ── autotune ──────────────────────────────────────────────────────────────────
if want "autotune"; then
    next "RAM-based autotune (conntrack, tw_buckets, file-max)"

    sudo tee /usr/bin/autotune.sh >/dev/null << 'SCRIPT'
#!/usr/bin/env bash
set -euo pipefail

# KICKSTART -- dynamic kernel tuning based on RAM from dpanic/patchfiles
# Tunes: nf_conntrack_max, tcp_max_tw_buckets, fs.file-max

if [ "$EUID" -ne 0 ]; then
    echo "Error: must be run as root" >&2
    exit 1
fi

MIN_CONNTRACK=65536
PER_GB=65536

MEM_KB=$(awk '/MemTotal/ {print $2}' /proc/meminfo)
RAM_GB=$(awk "BEGIN {ram_gb = $MEM_KB / 1024 / 1024; print (ram_gb == int(ram_gb)) ? int(ram_gb) : int(ram_gb) + 1}")
[ "$RAM_GB" -lt 1 ] && RAM_GB=1

TARGET_MAX=$((RAM_GB * PER_GB))
[ "$TARGET_MAX" -lt "$MIN_CONNTRACK" ] && TARGET_MAX=$MIN_CONNTRACK

# conntrack_max
if ! lsmod | grep -q "^nf_conntrack "; then
    modprobe nf_conntrack 2>/dev/null && sleep 1 || true
fi
if [ -f "/proc/sys/net/netfilter/nf_conntrack_max" ]; then
    CURRENT=$(sysctl -n net.netfilter.nf_conntrack_max 2>/dev/null || echo 0)
    if [ "$CURRENT" -ne "$TARGET_MAX" ]; then
        echo "Setting nf_conntrack_max=$TARGET_MAX (RAM=${RAM_GB}G, was=$CURRENT)"
        sysctl -w net.netfilter.nf_conntrack_max="$TARGET_MAX" >/dev/null
    fi
fi

# tcp_max_tw_buckets
TW_CURRENT=$(sysctl -n net.ipv4.tcp_max_tw_buckets 2>/dev/null || echo 0)
if [ "$TW_CURRENT" -ne "$TARGET_MAX" ]; then
    echo "Setting tcp_max_tw_buckets=$TARGET_MAX (RAM=${RAM_GB}G, was=$TW_CURRENT)"
    sysctl -w net.ipv4.tcp_max_tw_buckets="$TARGET_MAX" >/dev/null
fi

# fs.file-max
FILE_MAX_PER_GB=262144
FILE_MAX_TARGET=$((RAM_GB * FILE_MAX_PER_GB))
[ "$FILE_MAX_TARGET" -lt 1048576 ] && FILE_MAX_TARGET=1048576

FM_CURRENT=$(sysctl -n fs.file-max 2>/dev/null || echo 0)
if [ "$FM_CURRENT" -ne "$FILE_MAX_TARGET" ]; then
    echo "Setting fs.file-max=$FILE_MAX_TARGET (RAM=${RAM_GB}G, was=$FM_CURRENT)"
    sysctl -w fs.file-max="$FILE_MAX_TARGET" >/dev/null
fi
SCRIPT

    sudo chmod +x /usr/bin/autotune.sh

    # systemd service
    sudo tee /etc/systemd/system/autotune.service >/dev/null << 'EOF'
[Unit]
Description=Dynamic kernel parameter tuning based on RAM (kickstart)
After=systemd-sysctl.service network-online.target
Requires=systemd-sysctl.service
Wants=network-online.target

[Service]
Type=oneshot
User=root
Group=root
ExecStart=/usr/bin/autotune.sh
RemainAfterExit=yes
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable autotune.service 2>/dev/null || true
    echo "  done: /usr/bin/autotune.sh + autotune.service (enabled)"
fi

echo ""
echo "=== Kernel optimization complete ==="
echo "  Applied: ${COMPONENTS[*]}"
echo ""
echo "  A reboot is recommended to fully apply all changes."
