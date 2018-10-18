$ = document.querySelectorAll.bind(window.document);

// Apply Addon
Terminal.applyAddon(attach);
Terminal.applyAddon(fit);
Terminal.applyAddon(fullscreen);
Terminal.applyAddon(search);
Terminal.applyAddon(webLinks);
Terminal.applyAddon(winptyCompat);

var term;
var termFontSize = 15;
var terminalContainer = $("#terminal-container")[0];
var protocol = (location.protocol === 'https:') ? 'wss://' : 'ws://';
var socketURL = protocol + location.hostname + ((location.port) ? (':' + location.port) : '') + '/ws/terminal/';
var socket;

function postData(url, data) {
  return fetch(url, {
      method: 'POST',
      body: JSON.stringify(data),
      headers: {
        'content-type': 'application/json'
      }
    })
    .then(response => response.json())
}

function genSignature(params, appSecert) {
  // sort keys
  var keys = Object.keys(params).sort();

  // join k-v to string
  var paramStr = '';
  keys.map(key => paramStr += key + params[key]);

  return MD5(MD5(paramStr) + appSecert)
}

function connect() {
  var app_key, app_secert, node, container;

  var $app = $("#app")[0];
  if ($app) {
    app_secert = $app.value;
    var options = $app.options;
    app_key = options[options.selectedIndex].text;
  }

  var $node = $("#node")[0]
  if ($node) {
    node = $node.value
  }

  var $container = $("#container")[0]
  if ($container) {
    container = $container.value
  }

  if (!node || !container) {
    alert("node 和 container 都不能为空");
    return
  }

  var execConfig = {
    app_key: app_key,
    host: node,
    container_id: container,
  }
  execConfig.sign = genSignature(execConfig, app_secert);
  postData('/api/sign/exec', execConfig)
    .then(data => {
      if (data.token) {
        createTerminal(data.token)
      } else {
        alert(data.message || 'Token 获取失败')
      }
    })
    .catch(error => console.error(error))
}

function createTerminal(token) {
  // Clean terminal
  if (term) {
    term.dispose();
    socket && socket.close();
    socket = null;
    term = null;
  } else {
    while (terminalContainer.children.length) {
      terminalContainer.removeChild(terminalContainer.children[0]);
    }
  }

  // options refer https://github.com/xtermjs/xterm.js/blob/master/src/Terminal.ts#L77
  term = new Terminal({
    cursorBlink: true, // 光标闪烁,
  });

  term.open(terminalContainer);
  term.winptyCompatInit();
  term.webLinksInit();
  term.fit();
  term.focus();

  socket = new WebSocket(socketURL + token);
  socket.onopen = function () {
    term.attach(socket);
    term._initialized = true;
  }
  socket.onclose = function(e) {
    console.log('close', e)
    if (e.code == 1006) { // 服务端中断
      term.writeln('');
      term.writeln("Disconnect from server...");
    }
  }
  socket.onerror = function(e) {
    console.log('error', e)
  }
}

window.onresize = function () {
  if (term) {
    term.fit();
  }
}
