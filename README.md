# tanaage

tanaage is google drive upload client. upload files define  by config file.

# DESCRIPTION

This project is my private study project. API will be changed.

# Usage

```
$ tanaage --config config.yml
```

# Config File Sample

```yaml
client_id: abcdefg123456
client_email: abcdefg123456@exam.ple
private_key: aaaaaaaaaaaaabbbbbbbbbbbcccccccccccddddd
private_key_id: abcdefg123456
private_key: aaaaaaaaaaaaabbbbbbbbbbbcccccccccccddddd
type: "service_account"
folder: "test-folder"

# set dir name and upload dir
uploads:
   - from: img/
     to: image
   - from: movie/hoge
     to: movie
   - from: movie/fuga
     to: movie
```
