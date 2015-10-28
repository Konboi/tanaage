# tanaage

tanaage is google driver uploader from config file

# DESCRIPTION

This project is my private study project. API will be changed.

# Usage

```
$ tanaage --config config.yml
```

# Config File Sample

```yaml
oauth_token: abcdefg123456
folder: "test-folder"

# set dir name and upload dir
dir:
   - from: img/
     to: image
   - from: movie/hoge
     to: movie
   - from: movie/fuga
     to: movie
```
