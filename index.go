package gooey

const INDEX = `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Gooey App</title>
<script>
(function(){var e=window.location.href.substring(7),d=new WebSocket("ws://"+e+"gooeywebsocket"),c=void 0;window.hasOwnProperty("gooey")?c=window.gooey:(c={},window.gooey=c,c.OnMessage=function(a){console.log(a)},c.Send=function(a){1===d.readyState?d.send(JSON.stringify(a)):console.error("[GOOEY] Websocket connection is not open.")},c.IsDisconnected=!1,c.OnDisconnect=function(){console.error("[GOOEY] Disconnected from server.")},c.OpenNewTab=function(){var a=new XMLHttpRequest;a.open("GET",window.location+
"gooeynewtab",!0);a.send()});var f=window.setInterval(function(){3===d.readyState&&(window.clearInterval(f),c.IsDisconnected=!0,c.OnDisconnect())},1500);d.addEventListener("message",function(a){a=JSON.parse(a.data);if(a.hasOwnProperty("GooeyMessage")&&a.hasOwnProperty("GooeyContent")&&"gooey-server-reload-content"===a.GooeyMessage){console.clear();console.log("[GOOEY] Hot reload.");a=a.GooeyContent;""!==a.Body&&(document.body.innerHTML=a.Body);if(""!==a.CSS){var b=document.getElementById("gooey-reload-css-content");
b?b.innerHTML=a.CSS:(b=document.createElement("style"),b.id="gooey-reload-css-content",b.innerHTML=a.CSS,document.head.appendChild(b))}""!==a.Javascript&&((b=document.getElementById("gooey-reload-js-content"))&&document.head.removeChild(b),b=document.createElement("script"),b.id="gooey-reload-js-content",b.innerHTML=a.Javascript,document.head.appendChild(b))}else c.OnMessage(a)})})();
</script>
<script>
(function() {
window.gooey.OnMessage = function(msg) {
var elt = document.getElementById('gooey-message-area');
elt.innerText = msg;
};
})();
</script>
</head>
<body>
<h1>Gooey Sample Index Page</h1>
<div>
<h3>Last Received Gooey Message:</h3>
<div id="gooey-message-area"></div>
<div>
</body>
</html>
`
