const crypto = require("crypto");
const http = require("http");
const url = require('url');

function md5(str) {
  var md5sum = crypto.createHash("md5");
  md5sum.update(str);
  str = md5sum.digest("hex");
  return str;
}

function genSignature(params, appSecret) {
  delete params.sign; // 去除签名字段
  let keys = Object.keys(params).sort();
  let paramStr = "";

  keys.map(key => (paramStr += key + params[key]));

  return md5(md5(paramStr) + appSecret);
}

http.postJSON = function (addr, params) {
  addr = url.parse(addr)
  params = JSON.stringify(params)
  var options = {
    method: 'POST',
    host: addr.hostname,
    port: addr.port,
    path: addr.pathname,
    headers: {
      'Content-Type': 'application/json',
      'Content-Length': params.length
    }
  }

  return new Promise((resolve, reject) => {
    let req = this.request(options, (res) => {
      res.setEncoding('utf8');
      var data = '';
      res.on('data', (chunk) => {
        data += chunk;
      });
      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          headers: res.headers,
          data: data
        })
      });
    });
    req.on('error', (e) => {
      reject(e)
    })
    req.write(params);
    req.end();
  })
}

var appSecret = "bb";
var obj = {
  host: "192.168.33.11",
  container_id: "23deae9fc89d",
  app_key: "aa"
};
obj.sign = genSignature(obj, "bb");

var objJson = JSON.stringify(obj)


http.postJSON("http://127.0.0.1:8888/api/sign/exec", obj)
  .then(resp => {
    resp.data = JSON.parse(resp.data)
    if (resp.statusCode == 200) {
      console.log("ws://127.0.0.1:8888/ws/terminal/%s", resp.data.token)
    } else {
      console.log(resp.data.message)
    }
  })