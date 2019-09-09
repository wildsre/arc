## ARC(Application Resource Center)

- arc a cmdb with simple api

## Getting started

### Create Public Account

- create public account, record the token

```sh
curl 'http://127.0.0.1/arc/apply?name=public@local'
#token: f8789777c959acfe8ac4726b22f25b68
```

### Add Data

```sh
curl 'http://127.0.0.1/arc/add?app=myapp&resource=config&item=port&item=8000&token=f8789777c959acfe8ac4726b22f25b68'
```

### Get Data

get data without token, the default auth of default is 764, it means anyone have read auth

```sh
curl 'http://127.0.0.1/arc/get?app=myapp&resource=config&item=all'
```

### Update Data

```sh
curl 'http://127.0.0.1/arc/update?app=myapp&resource=config&item=port&value=9100&token=f8789777c959acfe8ac4726b22f25b68'
```


### Delete Data

```sh
curl 'http://127.0.0.1/arc/delete?app=myapp&resource=config&item=port&token=f8789777c959acfe8ac4726b22f25b68'
```
