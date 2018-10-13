const crypto = require("crypto");

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

function getPostJsonRequestForHttpie(path, params) {
  var paramsStr = '';
  if (params) {
    paramsStr = Object.keys(params)
      .map(k => {
        let type = typeof k;
        switch (type) {
          case "string":
            return `${k}="${params[k]}"`;
          case "object":
            return ""; // omit
          default:
            return `${k}=${params[k]}`;
        }
      })
      .join(" ");
  }

  return `http POST ${path} ${paramsStr}`;
}
var appSecret = "bb";
var obj = {
  host: "192.168.33.11",
  container_id: "23deae9fc89d",
  app_key: "aa"
};
obj.sign = genSignature(obj, "bb");

var apiSignExecStr = getPostJsonRequestForHttpie(
  "127.0.0.1:8888/api/sign/exec", obj
);

console.log(apiSignExecStr);