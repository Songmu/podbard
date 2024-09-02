# SKETCH

## Directory Structure

- index.md      - index page
- primcast.yaml - configuration file
- episode/      - eqisode pages
- audio/        - audio files
- template/     - template files
- static/       - static files

## Sub Commmands

### init

```console
$ primcast init .
```

### episode

```
$ primcast episode [-slug=hoge -date=2024-09-01 -title=title] audio/1.mp3
```

create a new epoisode page with the specified audio file.

### build

```
$ primcast build
```

build the site and output to the `public` directory.
