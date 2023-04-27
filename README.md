<!-- markdownlint-disable MD040 MD009 -->

# `dly` - your daily note from the command line

This small utility was built because I needed a quick way to add one-line entries to my daily notes, from the command line. These would usually be some clever thoughts I do not want to forget, TODOs, things to buy, ...

I do not want to start Logseq just for that (even though I mapped it to shortcuts) and so `dly` was born.

## Installation

The binary `dly.exe` (or the relevant one for your OS) attempts to read a configuration file located in `.config/dly/dly.yml` in your home directory. It will try hard to find said home directory and will abort if this is not possible. Should you encounter such a case, please open a bug report (a Github Issue).

If the configuration file is not found, a minimal empty one will be created. **It is not yet functional as it**, you need to edit it to add your Logseq daily notes folder (I could automate that if I knew how).

Binaries can be found in the Releases section.

### Building from source

```
go install github.com/wsw70/dly@latest
```

## Usage

Run `dly.exe`, a pop-up window appears to allow you to type your line. Once you are done press `Enter` (or click `OK`) and your daily note is updated.

If you use a tool such as [AutoHotKey](https://www.autohotkey.com/), you could bind a key combination to run `dly`.

Example for a hotkey under `Win-N`:

```
#n::Run "D:\Y\dev-perso\dly\dly.exe"
```

## Debugging

If you want to have a more verbose mode or help debug the program, you can set the environment variable `DLY_DEBUG` to the value `yes`. Below is an example of the consequences of setting it inline in PowerShell:

```
PS D:\Y\dev-perso\dly> go run . poiytporipoerit teprotperoite r eortieproit
2023-01-13T15:05:29+01:00 INFO note D:\Y\Logseq\journals\2023_01_13.md updated

PS D:\Y\dev-perso\dly> $env:DLY_DEBUG = 'yes'; go run . poiytporipoerit teprotperoite r eortieproit
2023-01-13T15:06:27+01:00 DEBUG text starts on a new line
2023-01-13T15:06:27+01:00 DEBUG backup note is C:TEMP\2023_01_13.md
2023-01-13T15:06:27+01:00 INFO note D:\Y\Logseq\journals\2023_01_13.md updated
```

## Configuration file

The configuration file located in `.config/dly/dly.yml` in your home directory is a YAML file that may contain the following directives

| directive | mandatory? | type | meaning |
| --- | --- | --- | :--- |
| `DailyNotesPath` | yes | string | Path to the LogSeq daily notes (typically the `journal` folder in Logseq data directory)
| `FilenameFormat` | yes | string | Format of your daily note, without the `.md` extension. The format follows (weird) Go formatting rules, see the [documentation](https://pkg.go.dev/time) or an [article](https://www.geeksforgeeks.org/time-formatting-in-golang/) for details. As a general rule, when you want to say "the current year" and expected something like `YYYY`, you use `2006` (yes, exactly this string). The "current month" is `01` and the "current day" is `02`. Yes this is insane. The default format (in the auto-generated file) is `2006_01_02` - this corresponds today to `2023_01_13` which in turns points to the file `2023_01_13.md`, which **Logseq** interprets as the date 2023-01-13.|
| `AddTimestamp` | no | bool | Should your line be prefixed with a bolded timestamp? |
| `AppendHashtag` | no | string | add the string as hashtag (example: `from-cli`, note the absence of `#` which will be automatically added) |
| `ShowNotificationOnSuccess`<br>(MS Windows only) | yes | bool | Show a tray notification when the note is added |
| `ShowNotificationOnNewversion`<br>(MS Windows only) | yes | bool | Show a tray notification when a new version of `dly` is available |

## What next?

- automated detection of the daily notes folder
- cleanup of code, including more comments
- more OSes (notably MacOS). It is easy to do but I need someone to test the binary (and information about where the home and temporary directories are)
- anything you can think of

Feel free to open Issues if you find bugs, or start Discussions.

I should probbaly add a license but I do not care, so let it be [WTFPL](https://en.wikipedia.org/wiki/WTFPL).

If you have the irrestible need to share your gratitude, call someone you love or send money to a clever charity that helps with education.
