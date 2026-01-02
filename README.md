A tiny Go utility that routes a URL to a Vivaldi browser profile and launches Vivaldi with the appropriate profile. It's useful when you want certain domains always to open in specific Vivaldi profiles (Work, Personal, etc.), or to show a GUI chooser for selected sites.

---

## Features

- Route URLs to a specific Vivaldi profile based on host/domain.
- Optional GUI profile picker for selected hosts (configurable).
- Validates configuration and rejects duplicate URL entries.
- Supports HTTP and HTTPS only; strips userinfo and ports when matching.
- Launches Vivaldi detached using the `--profile-directory` flag.
- Cross-platform start (Windows and Unix supported by platform-specific helpers in cmd).

---

## Quick note

> The router selects the **first valid URL** among command-line arguments and opens that one. If the URL's host is listed in `selectprofile`, a GUI chooser is shown. Otherwise, it uses the matching mapping or the configured default profile.

---

## Installation

### Building

Build from source:

Use goreleaser, then take the zip file content and extract it to a folder

```bash
goreleaser release --snapshot --clean
```

Or build normally then copy all needed files by hand

```bash
# From repo root
go build -o vivaldi-url-profile-router ./...
```

### Installing

Copy everything where you want it to live.

Open `config/add_as_browser.reg`, **edit**, save. Use `reg import add_as_browser.reg` in Powershell to import without UAC.

The binary expects a app.yaml file in the config directory (see configuration section).

---

## Configuration

The app looks for `app.yaml` in a config folder (`viper` is configured to use app.yaml).

Example app.yaml:

```yaml
browser:
  path: C:\Users\example\AppData\Local\Vivaldi\Application\vivaldi.exe
  default: Default # The folder name of the default profile

selectprofile:
  - https://mail.example.com/
  - https://intranet.example.com

mapping:
  - name: "Work"
    folder: "Profile 1"
    urls:
      - https://company.example.com/
      - https://something.else.example.com
```

Fields:
- `browser.path` - full path to Vivaldi executable.
- `browser.default` - profile folder name used when no mapping is matched.
- `selectprofile` - list of URLs whose host should trigger the GUI profile chooser.
- `mapping` - list of mappings:
  - `name` - human-friendly name.
  - `folder` - Vivaldi profile folder name to pass as `--profile-directory`.
  - `urls` - list of URLs (the host is extracted and normalized for matching).

Behavior:
- URL validation is performed at startup. Missing required fields or duplicates will make the app fail early with a helpful error message.
- Only the host is used for matching (lowercased and trailing dot removed). Subdomains are treated distinctly; list them explicitly if required.
- When a host matches `selectprofile`, the GUI shows available profiles (Default plus all configured mappings). Cancel exits without launching.

---

## Usage

Open a URL (first valid URL argument is used):

```bash
# Single URL
vivaldi-url-profile-router "https://company.example.com/some/path"

# Multiple arguments - the router picks the first parseable URL
vivaldi-url-profile-router some "garbage" "https://company.example.com"
```

It launches Vivaldi with a profile flag similar to:

```
vivaldi.exe --profile-directory="Profile 1" https://company.example.com/some/path
```

Exit behavior:
- Success: exits with status 0.
- On configuration or runtime errors: logs the error and exits non-zero. A `vupr_error.log` is written to the user's home directory when there is an error with finding the config files, otherwise there is an `app.log` file in the application directory.

---

## Implementation notes

- The router accepts only `http` and `https` schemes.
- Hosts with userinfo or ports are normalized (credentials removed, port stripped) before matching.
- Duplicate target URLs across mappings are disallowed and reported on startup.
- Code lives under cmd including platform-specific detached start logic (`start_detached_windows.go`, `start_detached_unix.go`).

---

## Changing the application icon

To change the Windows application icon, you need to generate your own COFF (.syso) object.

- Replace the `vivaldi-url-profile-router.png` file with your own
- Install rsrc: `go install github.com/akavel/rsrc@latest`
- Use imagemagick to convert the png to ico: `convert vivaldi-url-profile-router.png -define icon:auto-resize=256,128,64,48,32,16 vivaldi-url-profile-router.ico`
- Generate the COFF object: `rsrc -ico vivaldi-url-profile-router.ico -o vivaldi-url-profile-router_windows_amd64.syso -arch amd64`
- Rebuild the project, see Installation step above

---

## Contributing

Pull requests are welcome, but the tool is only intended for personal use by the author and still in development.

If you are using `go run .` to start the project, make sure to set the environment variable `VUPRDEV` to `true`, otherwise execution will fail.
This is because the location of the config files is relative to the executable, which is not true when using `go run`.
For convenience, there is a `run.sh` that you can use in Git Bash.

---

## License

This project is open source - see the LICENSE file.
