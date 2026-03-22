package modules

// Module describes a selectable item in the TUI menu.
type Module struct {
	ID          string   // unique key, e.g. "kernel-sysctl"
	Script      string   // relative path, e.g. "kernel/optimize.sh"
	Components  []string // sub-components or nil for standalone
	Label       string   // display name in menu
	Description string   // short description
	Category    string   // "optimization" or "installation"
	OS          string   // "all", "linux", "darwin"
	NeedsSudo   bool     // invoke with sudo bash
}

// AllModules returns the full registry, unfiltered.
func AllModules() []Module {
	return []Module{
		// ── Optimizations ──
		{ID: "gnome", Script: "gnome/optimize.sh", Label: "GNOME Optimize", Description: "disable animations, sounds, hot corners", Category: "optimization", OS: "linux"},
		{ID: "nautilus", Script: "nautilus/optimize.sh", Label: "Nautilus Optimize", Description: "restrict Tracker, limit thumbnails", Category: "optimization", OS: "linux"},
		{ID: "apparmor", Script: "apparmor/setup.sh", Label: "AppArmor Setup", Description: "learning mode with Slack reminder", Category: "optimization", OS: "linux", NeedsSudo: true},
		{ID: "kernel-sysctl", Script: "kernel/optimize.sh", Components: []string{"sysctl"}, Label: "Kernel ▸ sysctl.conf", Description: "network, memory, conntrack tuning", Category: "optimization", OS: "linux"},
		{ID: "kernel-limits", Script: "kernel/optimize.sh", Components: []string{"limits"}, Label: "Kernel ▸ limits", Description: "file descriptor & process limits", Category: "optimization", OS: "linux"},
		{ID: "kernel-scheduler", Script: "kernel/optimize.sh", Components: []string{"scheduler"}, Label: "Kernel ▸ I/O scheduler", Description: "none (SSD/NVMe)", Category: "optimization", OS: "linux"},
		{ID: "kernel-autotune", Script: "kernel/optimize.sh", Components: []string{"autotune"}, Label: "Kernel ▸ autotune", Description: "RAM-based autotune service", Category: "optimization", OS: "linux"},
		{ID: "sshd", Script: "sshd/setup.sh", Label: "SSH ▸ sshd hardening", Description: "disables password auth", Category: "optimization", OS: "linux"},
		// ── Installations ──
		{ID: "shell-zsh", Script: "shell/install.sh", Components: []string{"zsh"}, Label: "Shell ▸ zsh + oh-my-zsh", Description: "", Category: "installation", OS: "all"},
		{ID: "shell-fzf", Script: "shell/install.sh", Components: []string{"fzf"}, Label: "Shell ▸ fzf", Description: "fuzzy finder", Category: "installation", OS: "all"},
		{ID: "shell-starship", Script: "shell/install.sh", Components: []string{"starship"}, Label: "Shell ▸ starship prompt", Description: "", Category: "installation", OS: "all"},
		{ID: "shell-direnv", Script: "shell/install.sh", Components: []string{"direnv"}, Label: "Shell ▸ direnv", Description: "", Category: "installation", OS: "all"},
		{ID: "shell-plugins", Script: "shell/install.sh", Components: []string{"plugins"}, Label: "Shell ▸ zsh plugins", Description: "autosuggestions, syntax-highlighting", Category: "installation", OS: "all"},
		{ID: "shell-nvm", Script: "shell/install.sh", Components: []string{"nvm"}, Label: "Shell ▸ nvm", Description: "Node version manager", Category: "installation", OS: "all"},
		{ID: "shell-git", Script: "shell/install.sh", Components: []string{"git"}, Label: "Shell ▸ git config", Description: "LFS, SSH-over-HTTPS", Category: "installation", OS: "all"},
		{ID: "shell-byobu", Script: "shell/install.sh", Components: []string{"byobu"}, Label: "Shell ▸ byobu + tmux", Description: "", Category: "installation", OS: "linux"},
		{ID: "terminal-ncdu", Script: "terminal/install.sh", Components: []string{"ncdu"}, Label: "Terminal ▸ ncdu", Description: "disk analyzer", Category: "installation", OS: "all"},
		{ID: "yazi", Script: "yazi/install.sh", Label: "Yazi", Description: "terminal file manager", Category: "installation", OS: "all"},
		{ID: "docker", Script: "docker/install.sh", Label: "Docker", Description: "engine, compose, buildx, daemon config", Category: "installation", OS: "all"},
		{ID: "go", Script: "go/install.sh", Label: "Go", Description: "programming language from go.dev", Category: "installation", OS: "all"},
		{ID: "neovim", Script: "neovim/install.sh", Label: "Neovim + LazyVim", Description: "editor with IDE features", Category: "installation", OS: "all"},
		{ID: "browser-chrome", Script: "browsers/install.sh", Components: []string{"chrome"}, Label: "Browser ▸ Google Chrome", Description: "", Category: "installation", OS: "linux"},
		{ID: "browser-brave", Script: "browsers/install.sh", Components: []string{"brave"}, Label: "Browser ▸ Brave", Description: "", Category: "installation", OS: "linux"},
		{ID: "app-signal", Script: "browsers/install.sh", Components: []string{"signal"}, Label: "App ▸ Signal Desktop", Description: "", Category: "installation", OS: "linux"},
		{ID: "peazip", Script: "peazip/install.sh", Label: "PeaZip", Description: "archive manager (200+ formats)", Category: "installation", OS: "linux"},
	}
}

// ForOS returns modules matching the given OS ("linux" or "darwin").
func ForOS(goos string) []Module {
	all := AllModules()
	filtered := make([]Module, 0, len(all))
	for _, m := range all {
		if m.OS == "all" || m.OS == goos {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// NeedsUserInfo returns true if any module in the selection requires the GitInfo screen.
func NeedsUserInfo(selected []Module) bool {
	for _, m := range selected {
		if m.ID == "shell-git" || m.ID == "apparmor" {
			return true
		}
	}
	return false
}

// NeedsWebhook returns true if apparmor is in the selection.
func NeedsWebhook(selected []Module) bool {
	for _, m := range selected {
		if m.ID == "apparmor" {
			return true
		}
	}
	return false
}

// ScriptGroup represents a single script invocation with merged components.
type ScriptGroup struct {
	Script     string
	Components []string
	Label      string
	NeedsSudo  bool
	ModuleIDs  []string
}

// GroupByScript merges selected modules that share the same script path.
func GroupByScript(selected []Module) []ScriptGroup {
	seen := map[string]int{}
	var groups []ScriptGroup

	for _, m := range selected {
		key := m.Script
		if idx, ok := seen[key]; ok && len(m.Components) > 0 {
			groups[idx].Components = append(groups[idx].Components, m.Components...)
			groups[idx].ModuleIDs = append(groups[idx].ModuleIDs, m.ID)
		} else {
			seen[key] = len(groups)
			groups = append(groups, ScriptGroup{
				Script:     m.Script,
				Components: append([]string{}, m.Components...),
				Label:      m.Label,
				NeedsSudo:  m.NeedsSudo,
				ModuleIDs:  []string{m.ID},
			})
		}
	}
	return groups
}
