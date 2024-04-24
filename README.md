# zsh-manpage-completion-generator

Automatically generate zsh completions from man page using fish shell completion files.  
This program is inspired from [nevesnunes/sh-manpage-completions](https://github.com/nevesnunes/sh-manpage-completions).  
But it supports more completion files and is much simply implemented, faster and easier to use, written in Go.

## Requirements

This program depends on fish shell's manpage converter, [create_manpage_completions.py](https://github.com/fish-shell/fish-shell/blob/master/share/tools/create_manpage_completions.py).  
So you must first install fish shell, See below for installation instructions.

https://github.com/fish-shell/fish-shell

If you do not want to install fish, you can also manually place and run the conversion python script as an example below.

```
# download script
$ sudo wget --backups=1 -P /usr/local/bin/ https://raw.githubusercontent.com/fish-shell/fish-shell/master/share/tools/{create_manpage_completions,deroff}.py
$ sudo chmod a+x /usr/local/bin/{create_manpage_completions,deroff}.py

# create arbitrary fish completion folder
$ mkdir ~/fish_generated_completions

# generate fish completions from manpage
$ create_manpage_completions.py --manpath --cleanup-in ~/fish_generated_completions -d ~/fish_generated_completions --progress

# and then specify -src option to convert
$ zsh-manpage-completion-generator -src ~/fish_generated_completions
```


## Installation

**From binaries:**

Download the binary from [GitHub Releases](https://github.com/umlx5h/zsh-manpage-completion-generator/releases/latest) and place it in your `$PATH`.

Install the latest binary to `/usr/local/bin`:

```bash
curl -L "https://github.com/umlx5h/zsh-manpage-completion-generator/releases/latest/download/zsh-manpage-completion-generator_$(uname -s)_$(uname -m).tar.gz" | tar xz
chmod a+x ./zsh-manpage-completion-generator
sudo mv ./zsh-manpage-completion-generator /usr/local/bin/zsh-manpage-completion-generator
```

**Homebrew:**

```bash
brew install umlx5h/tap/zsh-manpage-completion-generator
```

**AUR (Arch User Repository):**

with any AUR helpers
```
yay -S zsh-manpage-completion-generator-bin
paru -S zsh-manpage-completion-generator-bin
```

**Go install:**

```bash
go install github.com/umlx5h/zsh-manpage-completion-generator@latest
```

## Usage

You must first generate fish competion files.  
It generates completion files, usually under folder `$XDG_DATA_HOME/.local/share/fish/generated_completions`

```console
$ fish -c 'fish_update_completions'
Parsing man pages and writing completions to /home/dummy/.local/share/fish/generated_completions/
```

Then generate zsh completions from fish completions folder.  
By default, it is generated in folder `$XDG_DATA_HOME/.local/share/zsh/generated_man_completions`

```console
$ zsh-manpage-completion-generator
Converting fish completions: /home/dummy/.local/share/fish/generated_completions -> /home/dummy/.local/share/zsh/generated_man_completions
Completed. converted: 2579/2581, skipped: 142
```

Then add completions folder to your `fpath` in `.zshrc`.  
(It is recommended to place them at the end of the `fpath`, so that human-generated completions such as [zsh-users/zsh-completions](https://github.com/zsh-users/zsh-completions) will be preferred.)

```
fpath=(
   $fpath
   $HOME/.local/share/zsh/generated_man_completions
)
compinit # This is not necessary if it is called after this.
```

You may have to force rebuild zcompdump

```console
$ rm -f "${ZDOTDIR-~}/.zcompdump" && compinit
```

It is recommended that `compinit` be called only once because of the startup time.  
You can check the number of calls to `compinit` by using `zprof`.

Note that `fpath` must be added before `compinit`.  
If you are using any zsh framework, check where it is adding `fpath` and calling `compinit`.

## Option

The following command line options are supported.

You can change the path to the fish and zsh completions by specifying `-dst` and `-src`.  
The `-clean` option can be specified to remove zsh completion folder before generating them.  
This is useful for removing completions that are no longer needed.  
(Note that the entire `-dst` folder will be deleted.)

```console
$ zsh-manpage-completion-generator -h
Usage of zsh-manpage-completion-generator:
  -clean
        CAUTION: remove destination folder before converting
  -dst string
        zsh generated_completions destination folder (default "/home/dummy/.local/share/zsh/generated_man_completions")
  -src string
        fish generated_completions src folder (default "/home/dummy/.local/share/fish/generated_completions")
  -verbose
        verbose log
  -version
        show version
```

## Caveat

- Some invalid options may be generated, but there is nothing that can be done about this, as this is also the case with fish.

## Related

- [nevesnunes/sh-manpage-completions](https://github.com/nevesnunes/sh-manpage-completions)
- [RobSis/zsh-completion-generator](https://github.com/RobSis/zsh-completion-generator)

## License

MIT
