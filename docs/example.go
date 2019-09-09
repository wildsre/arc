package docs

func Example() string { 
    var doc string
    doc = `Example:
#!/bin/bash
echo "apply"
token=$(curl -sS 'http://127.0.0.1:8023/arc/apply?name=robot@kingsoft.com' |jq -r .data.token)
echo "succ $token"

echo "add data1: app=a1 resource=r1 item=i1 value=v1"
curl -sS "http://127.0.0.1:8023/arc/add?token=${token}&app=a1&resource=r1&item=i1&value=v1"

echo "add data2: app=a1 resource=r1 item=i2 value=vvvv"
curl -sS "http://127.0.0.1:8023/arc/add?token=${token}&app=a1&resource=r1&item=i2&value=vvvv"

echo "list item"
curl -sS "http://127.0.0.1:8023/arc/get?token=${token}&app=all"

echo "get unexist data"
curl -sS "http://127.0.0.1:8023/arc/get?token=${token}&app=a556&resource=r1&item=i1&value=v1"

echo "delete data1: app=a1 resource=r1 item=i1"
curl -sS "http://127.0.0.1:8023/arc/delete?token=${token}&app=a1&resource=r1&item=i1"

echo "list all"
curl -sS "http://127.0.0.1:8023/arc/get?token=${token}&app=all"`
    return doc
}
