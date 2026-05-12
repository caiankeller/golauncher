# golauncher

A minimalist application launcher TUI written in Go with Charm libraries (`bubbletea`, `bubbles`, `lipgloss`).

Built for keyboard-first Linux workflows (especially tiling WMs), `golauncher` gives you a fast searchable app menu from your terminal. No images, no bulky UI, no bs; so simple that it can't be labeled ugly.

> There's no config file. Golauncher will use your terminal theme.

## Requirements

* Linux (Hyprland, i3, Sway, etc.)
* Go `1.25+`
* A terminal emulator

## Installation

### AUR

#### Using yay

```bash
yay -S golauncher-git
```

#### Manual AUR build

```bash
git clone https://aur.archlinux.org/golauncher-git.git
cd golauncher-git
makepkg -si
```

### Manual directly from this repo

```bash
git clone https://github.com/caiankeller/golauncher.git
cd golauncher
make
sudo make install
```

This installs `golauncher` to `/usr/local/bin`. Check [usage](#usage), [controls](#controls) or [Hyprland integration](#integration-with-hyprland)

## Uninstall

### AUR (yay and manual)

```bash
sudo pacman -R golauncher-git
```

### Manual directly from repo

In the cloned folder, run

```bash
sudo make uninstall
```

## Usage

In your terminal, run

```bash
golauncher
```

Press `/` to start searching, type to filter applications, then press `Enter` to launch.

### Available make targets

* `make build` — build the binary
* `make install` — build and install to `/usr/local/bin`
* `make uninstall` — remove from `/usr/local/bin`
* `make clean` — remove local build artifacts

## Integration with Hyprland

Add this to your `hyprland.conf`

```hypr
# Bind golauncher to Super+R
bind = SUPER, R, exec, $terminal --class "golauncher" -e golauncher
```

This binds golauncher to SUPER + R. Feel free to change this to your preferred shortcut. The `$terminal` variable represents your terminal emulator in Hyprland; if you haven't defined it yet, you can add it to your config

```hypr
$terminal = alacritty
```

Alternatively, you can call your terminal directly

```hypr
bind = SUPER, R, exec, alacritty --class "golauncher" -e golauncher
```

then add these rules for a modal-like floating window

```hypr
# Optional modal-like window behavior
windowrule = float, ^(golauncher)$
windowrule = center, ^(golauncher)$
windowrule = size 700 450, ^(golauncher)$
windowrule = dimaround, ^(golauncher)$
windowrule = stayfocused, ^(golauncher)$

```

## Controls

| Key | Action |
| --- | --- |
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `←` / `h` / `PgUp` | Previous page |
| `→` / `l` / `PgDn` | Next page |
| `g` / `Home` | Jump to top |
| `G` / `End` | Jump to bottom |
| `/` | Enter filter mode |
| `Enter` | Launch selected app and exit |
| `Esc` / `q` / `Ctrl+C` | Quit |

## License

[MIT](LICENSE)
